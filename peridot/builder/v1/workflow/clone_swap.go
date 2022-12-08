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

package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"time"
)

type CloneSwapStep1 struct {
	Batches [][]string
}

func (c *Controller) getSingleProject(projectId string) (*models.Project, error) {
	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(projectId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not list projects")
	}
	if len(projects) == 0 {
		return nil, errors.New("no projects found")
	}

	p := projects[0]
	return &p, nil
}

func (c *Controller) CloneSwapWorkflow(ctx workflow.Context, req *peridotpb.CloneSwapRequest, task *models.Task) (*peridotpb.CloneSwapTask, error) {
	var ret peridotpb.CloneSwapTask
	deferTask, errorDetails, err := c.commonCreateTask(task, &ret)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	var step1 CloneSwapStep1
	syncCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 25 * time.Minute,
		StartToCloseTimeout:    60 * time.Minute,
		TaskQueue:              c.mainQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err = workflow.ExecuteActivity(syncCtx, c.CloneSwapActivity, req, task).Get(ctx, &step1)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	// We're going to create a workflow for each batch of builds.
	for _, batch := range step1.Batches {
		yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue: "yumrepofs",
		})
		taskID := task.ID.String()
		updateRepoRequest := &UpdateRepoRequest{
			ProjectID:        req.TargetProjectId.Value,
			BuildIDs:         batch,
			Delete:           false,
			TaskID:           &taskID,
			NoDeletePrevious: true,
		}
		updateRepoTask := &yumrepofspb.UpdateRepoTask{}
		err = workflow.ExecuteChildWorkflow(yumrepoCtx, c.RepoUpdaterWorkflow, updateRepoRequest).Get(yumrepoCtx, updateRepoTask)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}
	}

	var flattenedBuilds []string
	for _, batch := range step1.Batches {
		flattenedBuilds = append(flattenedBuilds, batch...)
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	ret.TargetProjectId = req.TargetProjectId.Value
	ret.SrcProjectId = req.SrcProjectId.Value
	ret.BuildIdsLayered = flattenedBuilds

	return &ret, nil
}

func (c *Controller) CloneSwapActivity(ctx context.Context, req *peridotpb.CloneSwapRequest, task *models.Task) (*CloneSwapStep1, error) {
	srcProject, err := c.getSingleProject(req.SrcProjectId.Value)
	if err != nil {
		return nil, err
	}
	targetProject, err := c.getSingleProject(req.TargetProjectId.Value)
	if err != nil {
		return nil, err
	}

	// We're going to fetch all repositories in the source project, and then
	// copy them to the target project.
	srcRepos, err := c.db.FindRepositoriesForProject(srcProject.ID.String(), nil, false)
	if err != nil {
		return nil, err
	}

	beginTx, err := c.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "could not begin transaction")
	}
	tx := c.db.UseTransaction(beginTx)

	for _, srcRepo := range srcRepos {
		_ = c.logToMon(
			[]string{fmt.Sprintf("Processing %s/%s", srcProject.Name, srcRepo.Name)},
			task.ID.String(),
			utils.NullStringToEmptyString(task.ParentTaskId),
		)

		// Check if the repo exists in the target project.
		dbRepo, err := tx.GetRepository(nil, &srcRepo.Name, &req.TargetProjectId.Value)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "could not get repository")
		}
		if dbRepo == nil {
			// The repo doesn't exist in the target project, so we need to create it.
			dbRepo, err = tx.CreateRepositoryWithPackages(srcRepo.Name, req.TargetProjectId.Value, false, srcRepo.Packages)
			if err != nil {
				return nil, errors.Wrap(err, "could not create repository")
			}
			_ = c.logToMon(
				[]string{fmt.Sprintf("Created %s/%s", targetProject.Name, dbRepo.Name)},
				task.ID.String(),
				utils.NullStringToEmptyString(task.ParentTaskId),
			)
		}

		allArches := append(srcProject.Archs, "src")
		for _, a := range allArches {
			allArches = append(allArches, a+"-debug")
		}

		for _, arch := range allArches {
			srcRepoLatestRevision, err := tx.GetLatestActiveRepositoryRevision(srcRepo.ID.String(), arch)
			if err != nil && err != sql.ErrNoRows {
				return nil, errors.Wrap(err, "could not get latest active repository revision")
			}
			if srcRepoLatestRevision == nil {
				_ = c.logToMon(
					[]string{fmt.Sprintf("Skipping %s/%s/%s because it has no active revisions", srcProject.Name, srcRepo.Name, arch)},
					task.ID.String(),
					utils.NullStringToEmptyString(task.ParentTaskId),
				)
				continue
			}

			id := uuid.New()
			_, err = tx.CreateRevisionForRepository(
				id.String(),
				dbRepo.ID.String(),
				arch,
				srcRepoLatestRevision.RepomdXml,
				srcRepoLatestRevision.PrimaryXml,
				srcRepoLatestRevision.FilelistsXml,
				srcRepoLatestRevision.OtherXml,
				srcRepoLatestRevision.UpdateinfoXml,
				srcRepoLatestRevision.ModuleDefaultsYaml,
				srcRepoLatestRevision.ModulesYaml,
				srcRepoLatestRevision.GroupsXml,
				srcRepoLatestRevision.UrlMappings.String(),
			)
			if err != nil {
				return nil, errors.Wrap(err, "could not create repository revision")
			}

			_ = c.logToMon(
				[]string{fmt.Sprintf("Created revision %s for %s/%s/%s", id.String(), targetProject.Name, srcRepo.Name, arch)},
				task.ID.String(),
				utils.NullStringToEmptyString(task.ParentTaskId),
			)
		}
	}

	// Now let's get all succeeded builds for the target project.
	builds, err := tx.GetSuccessfulBuildIDsAsc(req.TargetProjectId.Value)
	if err != nil {
		return nil, errors.Wrap(err, "could not list builds")
	}

	// We're creating batches of 200 builds to process at a time.
	var syncBatches [][]string
	for i := 0; i < len(builds); i += 200 {
		end := i + 200
		if end > len(builds) {
			end = len(builds)
		}
		syncBatches = append(syncBatches, builds[i:end])
	}

	_ = c.logToMon(
		[]string{
			fmt.Sprintf("Created %d batches", len(syncBatches)),
			"Following builds will be synced:",
		},
		task.ID.String(),
		utils.NullStringToEmptyString(task.ParentTaskId),
	)
	for _, id := range builds {
		_ = c.logToMon(
			[]string{fmt.Sprintf("\t* %s", id)},
			task.ID.String(),
			utils.NullStringToEmptyString(task.ParentTaskId),
		)
	}

	err = beginTx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "could not commit transaction")
	}

	return &CloneSwapStep1{
		Batches: syncBatches,
	}, nil
}
