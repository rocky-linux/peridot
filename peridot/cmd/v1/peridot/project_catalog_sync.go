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
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"openapi.peridot.resf.org/peridotopenapi"
	"time"
)

var projectCatalogSync = &cobra.Command{
	Use: "catalog-sync",
	Run: projectCatalogSyncMn,
}

var (
	scmURL    string
	scmBranch string
)

func init() {
	projectCatalogSync.Flags().StringVar(&scmURL, "scm-url", "", "SCM URL (defaults to TARGET/peridot-config)")
	projectCatalogSync.Flags().StringVar(&scmBranch, "scm-branch", "", "SCM branch (defaults to {TARGET_BRANCH_PREFIX}{MAJOR_VERSION})")
}

func projectCatalogSyncMn(_ *cobra.Command, _ []string) {
	// Ensure project id exists
	projectId := mustGetProjectID()
	projectCl := getClient(serviceProject).(peridotopenapi.ProjectServiceApi)

	if scmURL == "" || scmBranch == "" {
		projectRes, _, err := projectCl.GetProject(getContext(), projectId).Execute()
		errFatal(err)
		p := projectRes.GetProject()

		if scmURL == "" {
			scmURL = fmt.Sprintf("%s/%s/peridot-config", p.GetTargetGitlabHost(), p.GetTargetPrefix())
		}
		if scmBranch == "" {
			scmBranch = fmt.Sprintf("%s%d", p.GetTargetBranchPrefix(), p.GetMajorVersion())
		}
	}

	body := peridotopenapi.InlineObject5{
		ScmUrl: &scmURL,
		Branch: &scmBranch,
	}
	req := projectCl.SyncCatalog(getContext(), projectId).Body(body)
	syncRes, _, err := req.Execute()
	errFatal(err)

	// Wait for sync to finish
	taskCl := getClient(serviceTask).(peridotopenapi.TaskServiceApi)
	log.Printf("Waiting for sync %s to finish\n", syncRes.GetTaskId())
	for {
		res, _, err := taskCl.GetTask(getContext(), projectId, syncRes.GetTaskId()).Execute()
		if err != nil {
			log.Printf("Error getting task: %s", err.Error())
			time.Sleep(5 * time.Second)
		}
		task := res.GetTask()
		if task.GetDone() {
			if task.GetSubtasks()[0].GetStatus() == peridotopenapi.SUCCEEDED {
				log.Printf("Sync %s finished successfully\n", syncRes.GetTaskId())
				break
			} else {
				log.Fatalf("Sync %s failed with status %s\n", syncRes.GetTaskId(), task.GetSubtasks()[0].GetStatus())
			}
		}

		time.Sleep(5 * time.Second)
	}
}
