// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	// "strconv"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"openapi.peridot.resf.org/peridotopenapi"
)

var taskLogs = &cobra.Command{
	Use:  "logs [name-or-buildId]",
	Args: cobra.ExactArgs(1),
	Run:  taskLogsMn,
}

var (
	architecture    string
	cwd             string
	combined        bool
	taskLogFileName string
	attrs           int
	succeeded       bool
	cancelled       bool
	failed          bool
)

func init() {
	taskLogs.Flags().StringVarP(&architecture, "architecture", "a", "", "(inop) filter by architecture")
	taskLogs.Flags().StringVarP(&cwd, "cwd", "C", "", "change working directory for ouput")
	taskLogs.Flags().BoolVarP(&combined, "combined", "c", false, "dump all logs to one file")
	taskLogs.Flags().BoolVar(&succeeded, "succeeded", true, "only query successful tasks")
	taskLogs.Flags().BoolVar(&cancelled, "cancelled", false, "only query cancelled tasks")
	taskLogs.Flags().BoolVar(&failed, "failed", false, "only query failed tasks")
	taskLogs.MarkFlagsMutuallyExclusive("cancelled", "failed", "succeeded")
}

func taskLogsMn(_ *cobra.Command, args []string) {
	// Ensure project id exists
	projectId := mustGetProjectID()

	buildIdOrPackageName := args[0]
	var buildId string

	err := uuid.Validate(buildIdOrPackageName)
	if err == nil {
		buildId = buildIdOrPackageName
	} else {
		// argument is not a uuid, try to look up the most recent build for a package with said name
		var status string
		if failed {
			status = string(peridotopenapi.FAILED)
		}
		if cancelled {
			status = string(peridotopenapi.CANCELED)
		}
		buildId, err = getLatestBuildTaskIdForPackageName(projectId, buildIdOrPackageName, status)
		if err != nil {
			errFatal(err)
		}
	}

	if cwd != "" {
		err := os.Chdir(cwd)
		if err != nil {
			errFatal(fmt.Errorf("Error during chdir: %w", err.Error()))
		}
	}

	if combined {
		// open and close the file to truncate it
		taskLogFileName = fmt.Sprintf("%s.log", buildId)
		attrs = os.O_RDWR | os.O_APPEND | os.O_CREATE
		if _, err := os.Stat(taskLogFileName); !errors.Is(err, os.ErrNotExist) {
			file, err := os.OpenFile(taskLogFileName, os.O_RDWR|os.O_TRUNC, 0666)
			if err != nil {
				errFatal(err)
			}
			defer file.Close()
			log.Printf("Truncating %s because combined logs were requested", taskLogFileName)
		}
	}

	// Wait for build to finish
	taskCl := getClient(serviceTask).(peridotopenapi.TaskServiceApi)
	log.Printf("Checking if parent task %s is finished\n", buildId)

	const (
		retryInterval = 5 * time.Second
		maxRetries    = 5
	)
	var retryCount = 0

	for {
		res, _, err := taskCl.GetTask(getContext(), projectId, buildId).Execute()
		if err != nil {
			log.Printf("Error getting task: %s", err.Error())
			if retryCount < maxRetries {
				retryCount++
				time.Sleep(retryInterval)
				continue
			}
			errFatal(fmt.Errorf("max retries reached"))
		}

		task := res.GetTask()
		if !task.GetDone() {
			log.Printf("task not done after %v retries: %v", retryCount, err.Error())
			if retryCount < maxRetries {
				retryCount++
				time.Sleep(retryInterval)
				continue
			}
			errFatal(fmt.Errorf("max retries reached"))
		}

		for _, t := range task.GetSubtasks() {
			taskType, ok := t.GetTypeOk()

			if !ok {
				continue
			}

			switch *taskType {
			case peridotopenapi.BUILD_ARCH:
				// NOTE(neil): 2024-07-25 - ignore error as it tries to unsuccessfully unmarshall json from logs
				taskId := t.GetId()
				taskArch := t.GetArch()

				_, resp, _ := taskCl.StreamTaskLogs(getContext(), projectId, taskId).Execute()

				defer resp.Body.Close()
				if resp != nil && resp.StatusCode == 200 {
					// log.Printf("%v", resp.Status)
					if !combined {
						taskLogFileName = fmt.Sprintf("%s_%s-%s.log", buildId, taskId, taskArch)
						attrs = os.O_RDWR | os.O_CREATE | os.O_TRUNC
					}

					status, ok := t.GetStatusOk()
					if !ok {
						errFatal(fmt.Errorf("unable to get status for task: %v", status))
					}

					statusString := string(*status.Ptr())
					statusString = statusString[strings.LastIndex(statusString, "_")+1:]

					log.Printf("Writing logs for task (arch=%s,tid=%s,status=%s) to %v", taskArch, taskId, statusString, taskLogFileName)

					file, err := os.OpenFile(taskLogFileName, attrs, 0666)
					if err != nil {
						errFatal(err)
					}
					defer file.Close()

					_, err = file.ReadFrom(resp.Body)
					if err != nil {
						errFatal(err)
					}
				}
			}
		}
		break
	}
}
