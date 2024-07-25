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
	"strconv"
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
)

func init() {
	taskLogs.Flags().StringVarP(&architecture, "architecture", "a", "", "(inop) filter by architecture")
	taskLogs.Flags().BoolVarP(&combined, "combined", "c", false, "dump all logs to one file")
	taskLogs.Flags().StringVarP(&cwd, "cwd", "C", "", "change working directory for ouput")
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
		// projectCl := getClient(serviceProject).(peridotopenapi.ProjectServiceApi)
		packageCl := getClient(servicePackage).(peridotopenapi.PackageServiceApi)
		buildCl := getClient(serviceBuild).(peridotopenapi.BuildServiceApi)

		_, _, err := packageCl.GetPackage(getContext(), projectId, "name", buildIdOrPackageName).Execute()
		if err != nil {
			errFatal(err)
		}
		// var pkg peridotopenapi.V1Package = *res.Package
		// pkgId := pkg.GetId()

		// try to get the latest builds for the package
		res, _, err := buildCl.ListBuilds(
			getContext(),
			projectId).FiltersStatus(string(peridotopenapi.SUCCEEDED)).FiltersPackageName(buildIdOrPackageName).Execute()
		if err != nil {
			errFatal(err)
		}

		// TODO(neil): why is Total a string?
		total, err := strconv.Atoi(*res.Total)
		if err != nil {
			errFatal(err)
		}

		// TODO(neil): support pagination
		if total > int(*res.Size) {
			panic("result set larger than one page")
		}

		if total > 0 {
			builds := *res.Builds

			// for _, build := range builds {
			// 	buildjson, _ := build.MarshalJSON()
			// 	log.Printf("build: %s", buildjson)
			// }

			// after sorting, the first build is the latest
			buildId = builds[0].GetTaskId()
		}
	}

	if cwd != "" {
		os.Chdir(cwd)
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
	log.Printf("Checking if build %s is finished\n", buildId)

	for {
		res, _, err := taskCl.GetTask(getContext(), projectId, buildId).Execute()
		if err != nil {
			log.Printf("Error getting task: %s", err.Error())
			time.Sleep(5 * time.Second)
		}
		task := res.GetTask()
		if task.GetDone() {
			for _, t := range task.GetSubtasks() {
				taskType, ok := t.GetTypeOk()

				if !ok {
					continue
				}

				switch *taskType {
				case peridotopenapi.BUILD_ARCH:
					// NOTE(neil): 2024-07-25 - ignore error as it tries to unsuccessfully unmarshall json from logs
					_, resp, _ := taskCl.StreamTaskLogs(getContext(), projectId, t.GetId()).Execute()

					defer resp.Body.Close()
					if resp != nil && resp.StatusCode == 200 {
						// log.Printf("%v", resp.Status)
						if !combined {
							taskLogFileName = fmt.Sprintf("%s_%s-%s.log", buildId, t.GetId(), t.GetArch())
							attrs = os.O_RDWR | os.O_CREATE | os.O_TRUNC
						}
						log.Printf("Writing logs for task (arch=%s,tid=%s) to %v", t.GetArch(), t.GetId(), taskLogFileName)

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

		time.Sleep(5 * time.Second)
	}
}
