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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"openapi.peridot.resf.org/peridotopenapi"
	"os"
	"peridot.resf.org/utils"
	"time"
)

type LookasideUploadTask struct {
	Task struct {
		Subtasks []struct {
			Response struct {
				Digest string `json:"digest"`
			} `json:"response"`
		} `json:"subtasks"`
	} `json:"task"`
}

var buildRpmImport = &cobra.Command{
	Use:  "rpm-import [*.rpm]",
	Args: cobra.MinimumNArgs(1),
	Run:  buildRpmImportMn,
}

var buildRpmImportForceOverride bool

func init() {
	buildRpmImport.Flags().BoolVar(&buildRpmImportForceOverride, "force-override", true, "Force override even if version exists (default: true)")
}

func isFile(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}

func buildRpmImportMn(_ *cobra.Command, args []string) {
	// Ensure project id exists
	projectId := mustGetProjectID()
	_ = projectId

	// Ensure all args are valid files
	for _, arg := range args {
		if !isFile(arg) {
			log.Fatalf("%s is not a valid file", arg)
		}
	}

	// Upload blobs to lookaside and wait for operation to finish
	var operations []string
	projectCl := getClient(serviceProject).(peridotopenapi.ProjectServiceApi)
	for _, arg := range args {
		bts, err := ioutil.ReadFile(arg)
		errFatal(err)
		base64EncodedBytes := base64.StdEncoding.EncodeToString(bts)

		res, _, err := projectCl.LookasideFileUpload(getContext()).Body(peridotopenapi.V1LookasideFileUploadRequest{
			File: &base64EncodedBytes,
		}).Execute()
		errFatal(err)
		log.Printf("Uploading %s to lookaside with task id %s\n", arg, res.GetTaskId())
		operations = append(operations, res.GetTaskId())
	}

	log.Println("Waiting for upload tasks to finish...")

	// Wait for tasks to reach success state
	taskCl := getClient(serviceTask).(peridotopenapi.TaskServiceApi)
	var doneOperations []string
	var blobs []string
	for {
		didBreak := false
		for _, op := range operations {
			log.Printf("Waiting for %s to finish\n", op)
			if len(doneOperations) == len(operations) {
				didBreak = true
				break
			}
			if utils.StrContains(op, doneOperations) {
				continue
			}

			res, resp, err := taskCl.GetTask(getContext(), "global", op).Execute()
			errFatal(err)
			task := res.GetTask()
			if task.GetDone() {
				subtask := task.GetSubtasks()[0]
				if subtask.GetStatus() == peridotopenapi.SUCCEEDED {
					b, err := ioutil.ReadAll(resp.Body)
					errFatal(err)

					var subtaskFull LookasideUploadTask
					errFatal(json.Unmarshal(b, &subtaskFull))

					blobs = append(blobs, subtaskFull.Task.Subtasks[0].Response.Digest)
					doneOperations = append(doneOperations, op)
					log.Printf("Task %s finished successfully\n", op)
				} else if subtask.GetStatus() != peridotopenapi.RUNNING || subtask.GetStatus() != peridotopenapi.PENDING {
					errFatal(fmt.Errorf("subtask %s failed with status %s", op, subtask.GetStatus()))
				}
			}

			time.Sleep(2 * time.Second)
		}
		if didBreak {
			break
		}
	}

	log.Println("Upload tasks finished")
	log.Println("Triggering RPM batch import")

	cl := getClient(serviceBuild).(peridotopenapi.BuildServiceApi)
	_, _, err := cl.RpmLookasideBatchImport(getContext(), projectId).
		Body(peridotopenapi.InlineObject4{
			LookasideBlobs: &blobs,
			ForceOverride:  &buildRpmImportForceOverride,
		}).Execute()
	errFatal(err)
}
