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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"openapi.peridot.resf.org/peridotopenapi"
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
var skipStep string

func init() {
	buildRpmImport.Flags().BoolVar(&buildRpmImportForceOverride, "force-override", true, "Force override even if version exists (default: true)")
	buildRpmImport.Flags().StringVarP(&skipStep, "skip", "s", "", "which step to skip")
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

	var skipUpload = false
	var skipImport = false

	if skipStep != "" {
		switch strings.ToLower(skipStep) {
		case "upload":
			skipUpload = true
		case "import":
			skipImport = true
		default:
			log.Fatalf("invalid skip step: %s", skipStep)
		}
	}

	// Ensure all args are valid files
	for _, arg := range args {
		if !isFile(arg) {
			log.Fatalf("%s is not a valid file", arg)
		}
	}

	// Upload blobs to lookaside and wait for operation to finish
	var blobs []string
	projectCl := getClient(serviceProject).(peridotopenapi.ProjectServiceApi)
	for _, arg := range args {

		bts, err := os.ReadFile(arg)
		errFatal(err)

		hash := sha256.Sum256(bts)
		shasum := hex.EncodeToString(hash[:])

		if !skipUpload {
			base64EncodedBytes := base64.StdEncoding.EncodeToString(bts)
			_, _, err := projectCl.LookasideFileUpload(getContext()).Body(peridotopenapi.V1LookasideFileUploadRequest{
				File: &base64EncodedBytes,
			}).Execute()
			errFatal(err)
			log.Printf("Uploaded %s to lookaside", arg)
		}
		log.Printf("Will upload %s to lookaside for %s", shasum, arg)
		blobs = append(blobs, shasum)
	}

	if skipImport {
		return
	}

	taskCl := getClient(serviceTask).(peridotopenapi.TaskServiceApi)

	log.Println("Triggering RPM batch import")

	cl := getClient(serviceBuild).(peridotopenapi.BuildServiceApi)
	importRes, _, err := cl.RpmLookasideBatchImport(getContext(), projectId).
		Body(peridotopenapi.BuildServiceRpmLookasideBatchImportBody{
			LookasideBlobs: &blobs,
			ForceOverride:  &buildRpmImportForceOverride,
		}).Execute()
	errFatal(err)

	// Wait for import to finish
	log.Printf("Waiting for import %s to finish\n", importRes.GetTaskId())
	for {
		res, _, err := taskCl.GetTask(getContext(), projectId, importRes.GetTaskId()).Execute()
		if err != nil {
			log.Printf("Error getting task: %s", err.Error())
			time.Sleep(5 * time.Second)
		}
		task := res.GetTask()
		if task.GetDone() {
			if task.GetSubtasks()[0].GetStatus() == peridotopenapi.SUCCEEDED {
				log.Printf("Import %s finished successfully\n", importRes.GetTaskId())
				break
			} else {
				log.Fatalf("Import %s failed with status %s\n", importRes.GetTaskId(), task.GetSubtasks()[0].GetStatus())
			}
		}

		time.Sleep(5 * time.Second)
	}
}
