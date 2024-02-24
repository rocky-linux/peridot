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
	"github.com/spf13/cobra"
	"log"
	"openapi.peridot.resf.org/peridotopenapi"
)

var projectCreateHashedRepos = &cobra.Command{
	Use:  "create-hashed-repos [repositories]",
	Args: cobra.MinimumNArgs(1),
	Run:  projectCreateHashedReposMn,
}

func projectCreateHashedReposMn(_ *cobra.Command, args []string) {
	projectID := mustGetProjectID()

	taskCl := getClient(serviceTask).(peridotopenapi.TaskServiceApi)
	cl := getClient(serviceProject).(peridotopenapi.ProjectServiceApi)

	hashedRes, _, err := cl.CreateHashedRepositories(getContext(), projectID).
		Body(peridotopenapi.ProjectServiceCreateHashedRepositoriesBody{
			Repositories: &args,
		}).
		Execute()
	errFatal(err)

	// Wait for task to complete
	log.Printf("Waiting for hashed operation %s to finish\n", hashedRes.GetTaskId())
	for {
		res, _, err := taskCl.GetTask(getContext(), projectID, hashedRes.GetTaskId()).Execute()
		errFatal(err)
		task := res.GetTask()
		if task.GetDone() {
			if task.GetSubtasks()[0].GetStatus() == peridotopenapi.SUCCEEDED {
				log.Printf("Hashed operation %s finished successfully\n", hashedRes.GetTaskId())
				break
			} else {
				log.Printf("Hashed operation %s failed with status %s\n", hashedRes.GetTaskId(), task.GetSubtasks()[0].GetStatus())
				break
			}
		}
	}
}
