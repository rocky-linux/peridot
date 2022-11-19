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
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"path/filepath"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

type FutureContext struct {
	Future    workflow.Future
	Ctx       workflow.Context
	TaskQueue string
	Task      *models.Task
}

type sideEffectBuild struct {
	Task  *models.Task
	Build *models.Build
}

type channelArchBuild struct {
	err          error
	newArtifacts []*peridotpb.TaskArtifact
}

func getSrpmBuildArches(project *models.Project, metadata *anypb.Any) ([]string, error) {
	// add all project arches first
	arches := project.Archs

	// let's check srpm for incompatible arches / noarch
	rpmArtifactMetadata := &peridotpb.RpmArtifactMetadata{}
	err := metadata.UnmarshalTo(rpmArtifactMetadata)
	if err != nil {
		return nil, err
	}
	if len(rpmArtifactMetadata.BuildArch) > 0 {
		arches = rpmArtifactMetadata.BuildArch
	}
	// if exclusive to specific arches, then only build for those
	if len(rpmArtifactMetadata.ExclusiveArch) > 0 {
		var newArches []string
		for _, arch := range arches {
			if utils.StrContains(arch, rpmArtifactMetadata.ExclusiveArch) {
				newArches = append(newArches, arch)
			}
		}
		arches = newArches
	}
	if len(rpmArtifactMetadata.ExcludeArch) > 0 {
		var newArches []string
		for _, arch := range arches {
			if !utils.StrContains(arch, rpmArtifactMetadata.ExcludeArch) {
				newArches = append(newArches, arch)
			}
		}
		arches = newArches
	}
	// if noarch isn't excluded and is in build arches or exclusive arches, then add it
	// the kernel package for example has a separate build process for noarch
	if !utils.StrContains("noarch", rpmArtifactMetadata.ExcludeArch) && (utils.StrContains("noarch", rpmArtifactMetadata.BuildArch) || utils.StrContains("noarch", rpmArtifactMetadata.ExclusiveArch)) {
		arches = append(arches, "noarch")
	}

	// verify that we're not adding any arches that are not in the project
	var finalArches []string
	for _, arch := range arches {
		if arch == "noarch" || utils.StrContains(arch, project.Archs) {
			// double check that we're not adding duplicate arches
			if !utils.StrContains(arch, finalArches) {
				finalArches = append(finalArches, arch)
			}
		}
	}

	if len(finalArches) == 0 {
		return nil, errors.New("no arches found for project")
	}
	arches = finalArches

	return arches, nil
}

// BuildBatchWorkflow is a workflow that builds a batch of packages (executes BuildWorkflow as a child workflow)
func (c *Controller) BuildBatchWorkflow(ctx workflow.Context, req *peridotpb.SubmitBuildBatchRequest, newBuildReqs []string, newModuleBuildReqs []string, buildBatchId string, user *utils.ContextUser) error {
	var futures []FutureContext

	for _, buildPkg := range newBuildReqs {
		buildReq := &peridotpb.SubmitBuildRequest{
			ProjectId: req.ProjectId,
			Package: &peridotpb.SubmitBuildRequest_PackageName{
				PackageName: wrapperspb.String(buildPkg),
			},
		}
		triggerCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue:           c.mainQueue,
			WorkflowTaskTimeout: 3 * time.Hour,
		})
		futures = append(futures, FutureContext{
			Ctx:       triggerCtx,
			Future:    workflow.ExecuteChildWorkflow(triggerCtx, c.TriggerBuildFromBatchWorkflow, buildReq, buildBatchId, false, user),
			TaskQueue: c.mainQueue,
		})
	}
	for _, buildPkg := range newModuleBuildReqs {
		buildReq := &peridotpb.SubmitBuildRequest{
			ProjectId: req.ProjectId,
			Package: &peridotpb.SubmitBuildRequest_PackageName{
				PackageName: wrapperspb.String(buildPkg),
			},
		}
		triggerCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue:           c.mainQueue,
			WorkflowTaskTimeout: 3 * time.Hour,
		})
		futures = append(futures, FutureContext{
			Ctx:       triggerCtx,
			Future:    workflow.ExecuteChildWorkflow(triggerCtx, c.TriggerBuildFromBatchWorkflow, buildReq, buildBatchId, true, user),
			TaskQueue: c.mainQueue,
		})
	}

	// Build failures doesn't mean a batch trigger has failed
	// A batch can contain failed builds
	for _, future := range futures {
		_ = future.Future.Get(future.Ctx, nil)
	}

	return nil
}

// TriggerBuildFromBatchWorkflow is a sub-workflow to create a task and trigger a build
func (c *Controller) TriggerBuildFromBatchWorkflow(ctx workflow.Context, req *peridotpb.SubmitBuildRequest, buildBatchId string, module bool, user *utils.ContextUser) error {
	var sideEffect sideEffectBuild
	sideEffectCall := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
			Id: wrapperspb.String(req.ProjectId),
		})
		if err != nil {
			return nil
		}
		if len(projects) != 1 {
			return nil
		}
		project := projects[0]

		filters := &peridotpb.PackageFilters{}
		switch p := req.Package.(type) {
		case *peridotpb.SubmitBuildRequest_PackageId:
			filters.Id = p.PackageId
		case *peridotpb.SubmitBuildRequest_PackageName:
			filters.NameExact = p.PackageName
		}

		pkgs, err := c.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
		if err != nil {
			return nil
		}
		if len(pkgs) != 1 {
			return nil
		}
		pkg := pkgs[0]

		var importRevision *models.ImportRevision

		if req.ScmHash == nil {
			// Each import can have multiple revisions
			importRevisions, err := c.db.GetLatestImportRevisionsForPackageInProject(pkg.Name, req.ProjectId)
			if err != nil {
				return nil
			}

			// Construct the branch this project is targeting
			// Using the constructed branch, we will check if there is a revision for the branch
			// In fork mode, if the project is configured correctly the upstream will have a revision for the branch
			upstreamBranch := fmt.Sprintf("%s%d%s", project.TargetBranchPrefix, project.MajorVersion, project.BranchSuffix.String)
			for _, revision := range importRevisions {
				if revision.ScmBranchName == upstreamBranch {
					importRevision = &*&revision
					break
				}
			}
		} else {
			importRevision, err = c.db.GetImportRevisionByScmHash(req.ScmHash.Value)
			if err != nil {
				if err == sql.ErrNoRows {
					activityErr := fmt.Errorf("import revision not found for scm hash %s", req.ScmHash.Value)
					c.log.Errorf("%s", activityErr.Error())
					return nil
				}
			}
		}
		if importRevision == nil {
			return nil
		}

		beginTx, err := c.db.Begin()
		if err != nil {
			c.log.Errorf("could not start transaction: %v", err)
			return nil
		}
		tx := c.db.UseTransaction(beginTx)

		task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_BUILD, &req.ProjectId, nil)
		if err != nil {
			c.log.Errorf("could not create build task in TriggerBuildFromBatchWorkflow: %v", err)
			return nil
		}

		metadataAnyPb, err := anypb.New(&peridotpb.PackageOperationMetadata{
			PackageName: pkg.Name,
		})
		if err != nil {
			return nil
		}
		err = tx.SetTaskMetadata(task.ID.String(), metadataAnyPb)
		if err != nil {
			c.log.Errorf("could not set task metadata in TriggerBuildFromBatchWorkflow: %v", err)
			return nil
		}

		var build *models.Build
		if module {
			err = beginTx.Commit()
			if err != nil {
				c.log.Errorf("could not commit transaction: %v", err)
				return nil
			}
		} else {
			build, err = tx.CreateBuild(pkg.ID.String(), importRevision.PackageVersionId, task.ID.String(), req.ProjectId)
			if err != nil {
				c.log.Errorf("could not create build in TriggerBuildFromBatchWorkflow: %v", err)
				return nil
			}

			err = tx.AttachBuildToBatch(build.ID.String(), buildBatchId)
			if err != nil {
				c.log.Errorf("could not attach build to batch: %v", err)
				return nil
			}

			err = beginTx.Commit()
			if err != nil {
				c.log.Errorf("could not commit transaction: %v", err)
				return nil
			}
		}

		return &sideEffectBuild{
			Task:  task,
			Build: build,
		}
	})
	err := sideEffectCall.Get(&sideEffect)
	if err != nil {
		return err
	}
	if sideEffect.Task == nil {
		return fmt.Errorf("could not trigger build")
	}

	buildCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:          sideEffect.Task.ID.String(),
		TaskQueue:           c.mainQueue,
		WorkflowTaskTimeout: 3 * time.Hour,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})

	if module {
		return workflow.ExecuteChildWorkflow(buildCtx, c.BuildModuleWorkflow, req, sideEffect.Task, &peridotpb.ExtraBuildOptions{
			BuildBatchId: buildBatchId,
		}).Get(buildCtx, nil)
	}

	return workflow.ExecuteChildWorkflow(buildCtx, c.BuildWorkflow, req, sideEffect.Task, &peridotpb.ExtraBuildOptions{
		BuildBatchId:    buildBatchId,
		ReusableBuildId: sideEffect.Build.ID.String(),
	}).Get(buildCtx, nil)
}

// BuildWorkflow initiates a new artifact build, which can currently be a standard package and a module.
// Modules trigger build activities for each component
// todo(mustafa): More info
func (c *Controller) BuildWorkflow(ctx workflow.Context, req *peridotpb.SubmitBuildRequest, task *models.Task, extraOptions *peridotpb.ExtraBuildOptions) (*peridotpb.SubmitBuildTask, error) {
	submitBuildTask := peridotpb.SubmitBuildTask{
		ChecksDisabled: req.DisableChecks,
		ParentTaskId:   utils.NullStringValueP(task.ParentTaskId),
	}

	deferTask, errorDetails, err := c.commonCreateTask(task, nil)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	filters := &peridotpb.PackageFilters{}
	switch p := req.Package.(type) {
	case *peridotpb.SubmitBuildRequest_PackageId:
		filters.Id = p.PackageId
	case *peridotpb.SubmitBuildRequest_PackageName:
		filters.NameExact = p.PackageName
	}

	pkgs, err := c.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if len(pkgs) != 1 {
		setPackageNotFoundError(errorDetails, req.ProjectId, ErrorDomainBuildsPeridot)
		return nil, utils.CouldNotRetrieveObjects
	}
	pkg := pkgs[0]

	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if len(projects) != 1 {
		setInternalError(errorDetails, errors.New("project could not be found"))
		return nil, utils.CouldNotRetrieveObjects
	}
	project := projects[0]

	var importRevision *models.ImportRevision

	if req.ScmHash == nil {
		// Each import can have multiple revisions
		importRevisions, err := c.db.GetLatestImportRevisionsForPackageInProject(pkg.Name, req.ProjectId)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, err
		}

		// Construct the branch this project is targeting
		// Using the constructed branch, we will check if there is a revision for the branch
		// In fork mode, if the project is configured correctly the upstream will have a revision for the branch
		upstreamBranch := fmt.Sprintf("%s%d%s", project.TargetBranchPrefix, project.MajorVersion, project.BranchSuffix.String)
		for _, revision := range importRevisions {
			if revision.ScmBranchName == upstreamBranch {
				importRevision = &*&revision
				break
			}
		}
	} else {
		importRevision, err = c.db.GetImportRevisionByScmHash(req.ScmHash.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				activityErr := fmt.Errorf("import revision not found for scm hash %s", req.ScmHash.Value)
				setActivityError(errorDetails, activityErr)
				return nil, activityErr
			}
		}
	}
	if importRevision == nil {
		setActivityError(errorDetails, errors.New("could not find upstream branch"))
		return nil, errors.New("could not find upstream branch")
	}
	packageVersion, err := c.db.GetPackageVersion(importRevision.PackageVersionId)
	if err != nil {
		err = fmt.Errorf("could not get package version for import revision %s: %v", importRevision.ID, err)
		setInternalError(errorDetails, err)
		return nil, err
	}

	// If there is a parent task, we need to use that ID as the parent task ID
	taskID := task.ID.String()
	if task.ParentTaskId.Valid {
		taskID = task.ParentTaskId.String
	}

	// Create a side repo if the build request specifies side NVRs
	// Packages specified here will be excluded from the main repo
	if len(req.SideNvrs) > 0 {
		var buildNvrs []models.Build
		var excludes []string
		for _, sideNvr := range req.SideNvrs {
			if !rpmutils.NVRNoArch().MatchString(sideNvr) {
				err = fmt.Errorf("invalid side NVR: %s", sideNvr)
				setActivityError(errorDetails, err)
				return nil, err
			}
			nvrMatch := rpmutils.NVRNoArch().FindStringSubmatch(sideNvr)
			nvrName := nvrMatch[1]
			nvrVersion := nvrMatch[2]
			nvrRelease := nvrMatch[3]

			buildNvrsFromBuild, err := c.db.GetBuildByPackageNameAndVersionAndRelease(nvrName, nvrVersion, nvrRelease)
			if err != nil {
				if err == sql.ErrNoRows {
					err = fmt.Errorf("side NVR %s not found in project %s", sideNvr, req.ProjectId)
					setActivityError(errorDetails, err)
					return nil, err
				}
				err = fmt.Errorf("could not get build for side NVR %s: %v", sideNvr, err)
				setInternalError(errorDetails, err)
				return nil, err
			}

			for _, buildNvr := range buildNvrsFromBuild {
				artifacts, err := c.db.GetArtifactsForBuild(buildNvr.ID.String())
				if err != nil {
					err = fmt.Errorf("could not get artifacts for build %s: %v", buildNvr.ID, err)
					setInternalError(errorDetails, err)
					return nil, err
				}

				for _, artifact := range artifacts {
					artifactName := strings.TrimPrefix(artifact.Name, fmt.Sprintf("%s/", buildNvr.TaskId))
					if rpmutils.NVR().MatchString(artifactName) {
						excludes = append(excludes, rpmutils.NVR().FindStringSubmatch(artifactName)[1])
					}
				}
			}

			buildNvrs = append(buildNvrs, buildNvrsFromBuild...)
		}

		repo, err := c.db.CreateRepositoryWithPackages(uuid.New().String(), req.ProjectId, true, []string{})
		if err != nil {
			c.log.Errorf("failed to create repository: %v", err)
			return nil, status.Error(codes.Internal, "failed to create repository for side nvrs")
		}

		if extraOptions == nil {
			extraOptions = &peridotpb.ExtraBuildOptions{}
		}
		if extraOptions.ExtraYumrepofsRepos == nil {
			extraOptions.ExtraYumrepofsRepos = []*peridotpb.ExtraYumrepofsRepo{}
		}
		if extraOptions.ExcludePackages == nil {
			extraOptions.ExcludePackages = []string{}
		}
		extraOptions.ExtraYumrepofsRepos = append(extraOptions.ExtraYumrepofsRepos, &peridotpb.ExtraYumrepofsRepo{
			Name:           repo.Name,
			ModuleHotfixes: true,
			IgnoreExclude:  true,
		})
		extraOptions.ExcludePackages = append(extraOptions.ExcludePackages, excludes...)

		var buildIds []string
		for _, build := range buildNvrs {
			buildIds = append(buildIds, build.ID.String())
		}

		yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue: "yumrepofs",
		})
		updateRepoRequest := &UpdateRepoRequest{
			ProjectID:        req.ProjectId,
			TaskID:           &taskID,
			BuildIDs:         buildIds,
			Delete:           false,
			ForceRepoId:      repo.ID.String(),
			ForceNonModular:  true,
			DisableSigning:   true,
			DisableSetActive: true,
		}
		updateRepoTask := &yumrepofspb.UpdateRepoTask{}
		err = workflow.ExecuteChildWorkflow(yumrepoCtx, c.RepoUpdaterWorkflow, updateRepoRequest).Get(ctx, updateRepoTask)
		if err != nil {
			return nil, err
		}
	}

	buildID := extraOptions.ReusableBuildId
	if buildID == "" {
		err = errors.New("reusable build id not found")
		setActivityError(errorDetails, err)
		return nil, err
	}

	projectId := project.ID.String()
	var srpmTask models.Task
	srpmTaskEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		newTask, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_BUILD_SRPM, &projectId, &taskID)
		if err != nil {
			return &models.Task{}
		}
		return newTask
	})
	err = srpmTaskEffect.Get(&srpmTask)
	if err != nil {
		setInternalError(errorDetails, fmt.Errorf("could not get srpm task effect: %v", err))
		return nil, err
	}
	if !srpmTask.ProjectId.Valid {
		setInternalError(errorDetails, fmt.Errorf("could not create srpm task"))
		return nil, err
	}

	// Provision a new worker specifically to build the SRPM
	// SRPMs can be built with any architecture
	srpmTaskQueue, cleanupSrpm, err := c.provisionWorker(ctx, &ProvisionWorkerRequest{
		TaskId:       srpmTask.ID.String(),
		ParentTaskId: sql.NullString{String: taskID, Valid: true},
		Purpose:      "srpm",
		Arch:         "noarch",
		ProjectId:    req.ProjectId,
		Privileged:   true,
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	defer cleanupSrpm()

	srpmCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 30 * time.Minute,
		StartToCloseTimeout:    time.Hour,
		HeartbeatTimeout:       2 * time.Minute,
		TaskQueue:              srpmTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	upstreamPrefix := fmt.Sprintf("%s/%s", project.TargetGitlabHost, project.TargetPrefix)
	err = workflow.ExecuteActivity(srpmCtx, c.BuildSRPMActivity, upstreamPrefix, importRevision.ScmHash, project.ID.String(), pkg.Name, packageVersion, srpmTask, extraOptions).Get(srpmCtx, nil)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	uploadSrpmCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 30 * time.Minute,
		StartToCloseTimeout:    time.Hour,
		HeartbeatTimeout:       time.Minute,
		TaskQueue:              srpmTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})
	var uploadSRPMResult UploadActivityResult
	err = workflow.ExecuteActivity(uploadSrpmCtx, c.UploadSRPMActivity, project.ID.String(), taskID).Get(uploadSrpmCtx, &uploadSRPMResult)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}
	cleanupSrpm()

	// todo(mustafa): Currently we do not support partial arch builds
	// Partial arch builds are when you introduce new architectures to a project
	// and be able to build them without affecting the existing architectures.
	// That is an important feature for Peridot, but not strictly necessary as of now.
	nvr := strings.TrimSuffix(filepath.Base(uploadSRPMResult.ObjectName), ".rpm")

	firstNvrCheckSideEffectCall := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		exists, err := c.db.NVRAExists(nvr)
		if err != nil {
			err = fmt.Errorf("could not check if NVR exists: %v", err)
			return err
		}

		return exists
	})
	var res interface{}
	err = firstNvrCheckSideEffectCall.Get(&res)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if err, ok := res.(error); ok {
		setActivityError(errorDetails, err)
		return nil, err
	}
	if exists, ok := res.(bool); ok && exists {
		err = fmt.Errorf("NVR %s already locked", nvr)
		task.Status = peridotpb.TaskStatus_TASK_STATUS_CANCELED
		_ = c.logToMon([]string{err.Error()}, task.ID.String(), taskID)
		setActivityError(errorDetails, err)
		return nil, err
	}

	subtask, err := c.db.GetTask(uploadSRPMResult.Subtask.ID.String(), utils.Pointer(project.ID.String()))
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	metadata := &anypb.Any{}
	err = protojson.Unmarshal(subtask[0].Metadata.JSONText, metadata)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	arches, err := getSrpmBuildArches(&project, metadata)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	// Invoke build for all archs
	var artifacts []*peridotpb.TaskArtifact
	archChannel := workflow.NewChannel(ctx)
	for _, archTop := range arches {
		archGoCtx := workflow.WithValue(ctx, "arch", archTop)
		workflow.Go(archGoCtx, func(ctx workflow.Context) {
			arch := ctx.Value("arch").(string)
			ret := &channelArchBuild{}
			defer archChannel.Send(ctx, ret)

			var archTask models.Task
			archTaskEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
				newTask, err := c.db.CreateTask(nil, arch, peridotpb.TaskType_TASK_TYPE_BUILD_ARCH, &projectId, &taskID)
				if err != nil {
					return &models.Task{}
				}
				return newTask
			})
			err := archTaskEffect.Get(&archTask)
			if err != nil || !archTask.ProjectId.Valid {
				ret.err = fmt.Errorf("failed to create arch task: %s", err)
				return
			}

			workerReq := &ProvisionWorkerRequest{
				TaskId:       archTask.ID.String(),
				ParentTaskId: sql.NullString{String: taskID, Valid: true},
				Purpose:      "b-" + arch,
				Arch:         arch,
				ProjectId:    req.ProjectId,
				HighResource: true,
				Privileged:   true,
			}
			archTaskQueue, cleanupArch, err := c.provisionWorker(ctx, workerReq)
			if err != nil {
				ret.err = fmt.Errorf("failed to provision worker: %s", err)
				return
			}
			defer cleanupArch()

			archCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToStartTimeout: 12 * time.Hour,
				StartToCloseTimeout:    48 * time.Hour,
				HeartbeatTimeout:       2 * time.Hour,
				TaskQueue:              archTaskQueue,
				RetryPolicy: &temporal.RetryPolicy{
					MaximumAttempts: 1,
				},
			})
			err = workflow.ExecuteActivity(archCtx, c.BuildArchActivity, project.ID.String(), pkg.Name, req.DisableChecks, packageVersion, uploadSRPMResult, archTask, arch, extraOptions).Get(archCtx, nil)
			if err != nil {
				ret.err = fmt.Errorf("failed to build arch %s: %s", arch, err)
				return
			}

			uploadArchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToStartTimeout: 12 * time.Hour,
				StartToCloseTimeout:    24 * time.Hour,
				HeartbeatTimeout:       2 * time.Minute,
				TaskQueue:              archTaskQueue,
			})

			var res []*UploadActivityResult
			err = workflow.ExecuteActivity(uploadArchCtx, c.UploadArchActivity, project.ID.String(), taskID).Get(ctx, &res)
			if err != nil {
				ret.err = fmt.Errorf("failed to upload arch %s: %s", arch, err)
				return
			}

			exists, err := c.db.NVRAExists(strings.TrimSuffix(filepath.Base(uploadSRPMResult.ObjectName), ".rpm"))
			if err != nil {
				ret.err = err
				return
			}

			// Only refresh SRPM if we don't already have it in the repo tree
			// This is to enable us to rebuild archs we don't have while keeping
			// the integrity of the already published repository
			if !exists {
				res = append(res, &uploadSRPMResult)
			}

			var newArtifacts []*peridotpb.TaskArtifact
			for _, result := range res {
				err = c.db.AttachTaskToBuild(buildID, result.Subtask.ID.String())
				if err != nil {
					ret.err = fmt.Errorf("failed to attach task to build: %s", err)
					return
				}
				if result.Skip {
					continue
				}
				subtask, err := c.db.GetTask(result.Subtask.ID.String(), utils.Pointer(project.ID.String()))
				if err != nil {
					ret.err = fmt.Errorf("failed to get task: %s", err)
					return
				}
				metadata := &anypb.Any{}
				err = protojson.Unmarshal(subtask[0].Metadata.JSONText, metadata)
				if err != nil {
					ret.err = fmt.Errorf("failed to unmarshal metadata: %s", err)
					return
				}
				ret.newArtifacts = append(newArtifacts, &peridotpb.TaskArtifact{
					TaskId:     result.Subtask.ID.String(),
					Name:       result.ObjectName,
					HashSha256: result.HashSha256,
					Arch:       result.Arch,
					Metadata:   metadata,
				})
				artifacts = append(artifacts, &peridotpb.TaskArtifact{
					TaskId:     result.Subtask.ID.String(),
					Name:       result.ObjectName,
					HashSha256: result.HashSha256,
					Arch:       result.Arch,
				})
			}
		})
	}

	var artifactNames []string
	archWaitChannel := workflow.NewChannel(ctx)
	for i := 0; i < len(arches); i++ {
		workflow.Go(ctx, func(ctx workflow.Context) {
			var res channelArchBuild
			archChannel.Receive(ctx, &res)

			if res.err != nil {
				archWaitChannel.Send(ctx, res.err)
				return
			}

			for _, artifact := range res.newArtifacts {
				if utils.StrContains(artifact.Name, artifactNames) {
					continue
				}
				artifactNames = append(artifactNames, artifact.Name)
			}

			archWaitChannel.Send(ctx, nil)
		})
	}

	var archErrs []error
	for i := 0; i < len(arches); i++ {
		var res interface{}
		archWaitChannel.Receive(ctx, &res)
		if err, ok := res.(error); ok {
			if err != nil {
				archErrs = append(archErrs, err)
			}
		}
	}
	isError := false
	isCanceled := false
	var retErr error
	for _, archErr := range archErrs {
		if archErr != nil {
			if !strings.Contains(archErr.Error(), "canceled") {
				isError = true
				retErr = archErr
			}
			if !isError && strings.Contains(archErr.Error(), "canceled") {
				isCanceled = true
				retErr = archErr
			}
		}
	}
	if retErr != nil {
		setActivityError(errorDetails, retErr)
		if isCanceled {
			task.Status = peridotpb.TaskStatus_TASK_STATUS_CANCELED
		} else {
			task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED
		}
		_ = c.db.SetTaskStatus(task.ID.String(), task.Status)

		return nil, retErr
	}

	submitBuildTask = peridotpb.SubmitBuildTask{
		BuildId:        buildID,
		BuildTaskId:    task.ID.String(),
		PackageName:    pkg.Name,
		ImportRevision: importRevision.ToProto(),
		Artifacts:      artifacts,
		ChecksDisabled: req.DisableChecks,
		Modular:        req.ModuleVariant,
		ParentTaskId:   utils.NullStringValueP(task.ParentTaskId),
		RepoChanges:    &yumrepofspb.UpdateRepoTask{Changes: []*yumrepofspb.RepositoryChange{}},
	}
	sbtAny, err := anypb.New(&submitBuildTask)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	err = c.db.SetTaskResponse(task.ID.String(), sbtAny)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_RUNNING

	// Save once here so RepoUpdaterWorkflow can use it
	deferTask()

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	if !extraOptions.DisableYumrepofsUpdates && !req.SetInactive {
		yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue: "yumrepofs",
		})
		updateRepoRequest := &UpdateRepoRequest{
			ProjectID: req.ProjectId,
			BuildIDs:  []string{buildID},
			Delete:    false,
			TaskID:    &taskID,
		}
		updateRepoTask := &yumrepofspb.UpdateRepoTask{}
		err = workflow.ExecuteChildWorkflow(yumrepoCtx, c.RepoUpdaterWorkflow, updateRepoRequest).Get(yumrepoCtx, updateRepoTask)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}

		submitBuildTask.RepoChanges = updateRepoTask
	}

	// Lock NVR only once
	effectCallNVRA := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		err = c.db.LockNVRA(nvr)
		if err != nil {
			// retry once again, stupid error
			err = c.db.LockNVRA(nvr)
			if err != nil {
				err = fmt.Errorf("failed to lock nvr: %s", err)
				setInternalError(errorDetails, err)
				return err
			}
		}

		return nil
	})
	var resNVRA interface{}
	err = effectCallNVRA.Get(&resNVRA)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if err, ok := resNVRA.(error); ok {
		setActivityError(errorDetails, err)
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &submitBuildTask, nil
}
