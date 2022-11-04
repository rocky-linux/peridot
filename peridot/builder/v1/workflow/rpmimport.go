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
	"archive/tar"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/cavaliergopher/rpm"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/rpmbuild"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

type RpmImportActivityTaskStage1 struct {
	Build *models.Build
}

type RpmImportUploadWrapper struct {
	Upload *UploadActivityResult
	TaskID string
}

func (c *Controller) RpmImportWorkflow(ctx workflow.Context, req *peridotpb.RpmImportRequest, task *models.Task) (*peridotpb.RpmImportTask, error) {
	var ret peridotpb.RpmImportTask
	deferTask, errorDetails, err := c.commonCreateTask(task, &ret)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	importTaskQueue, cleanupWorker, err := c.provisionWorker(ctx, &ProvisionWorkerRequest{
		TaskId:       task.ID.String(),
		ParentTaskId: task.ParentTaskId,
		Purpose:      "rpmimport",
		Arch:         "noarch",
		ProjectId:    req.ProjectId,
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	defer cleanupWorker()

	var importRes RpmImportActivityTaskStage1
	importCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour,
		HeartbeatTimeout:    20 * time.Second,
		TaskQueue:           importTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err = workflow.ExecuteActivity(importCtx, c.RpmImportActivity, req, task.ID.String(), false).Get(ctx, &importRes)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	uploadArchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 12 * time.Hour,
		StartToCloseTimeout:    24 * time.Hour,
		HeartbeatTimeout:       2 * time.Minute,
		TaskQueue:              importTaskQueue,
	})

	var res []*UploadActivityResult
	err = workflow.ExecuteActivity(uploadArchCtx, c.UploadArchActivity, req.ProjectId, task.ID.String()).Get(ctx, &res)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	for _, result := range res {
		err = c.db.AttachTaskToBuild(importRes.Build.ID.String(), result.Subtask.ID.String())
		if err != nil {
			err = status.Errorf(codes.Internal, "could not attach task to build: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}
		if result.Skip {
			continue
		}
	}

	taskID := task.ID.String()
	yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue: "yumrepofs",
	})
	updateRepoRequest := &UpdateRepoRequest{
		ProjectID:        req.ProjectId,
		BuildIDs:         []string{importRes.Build.ID.String()},
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

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	ret.RepoChanges = updateRepoTask
	return &ret, nil
}

func (c *Controller) RpmLookasideBatchImportWorkflow(ctx workflow.Context, req *peridotpb.RpmLookasideBatchImportRequest, task *models.Task) (*peridotpb.RpmLookasideBatchImportTask, error) {
	var ret peridotpb.RpmLookasideBatchImportTask
	deferTask, errorDetails, err := c.commonCreateTask(task, &ret)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	importTaskQueue, cleanupWorker, err := c.provisionWorker(ctx, &ProvisionWorkerRequest{
		TaskId:       task.ID.String(),
		ParentTaskId: task.ParentTaskId,
		Purpose:      "batchrpmimport",
		Arch:         "noarch",
		ProjectId:    req.ProjectId,
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	defer cleanupWorker()

	taskID := task.ID.String()
	var stage1 *RpmImportActivityTaskStage1
	for _, blob := range req.LookasideBlobs {
		var importRes RpmImportActivityTaskStage1
		importCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: time.Hour,
			HeartbeatTimeout:    20 * time.Second,
			TaskQueue:           importTaskQueue,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 1,
			},
		})
		blobReq := &peridotpb.RpmImportRequest{
			ProjectId:     req.ProjectId,
			Rpms:          blob,
			ForceOverride: req.ForceOverride,
		}
		err = workflow.ExecuteActivity(importCtx, c.RpmImportActivity, blobReq, task.ID.String(), true, stage1).Get(ctx, &importRes)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}
		if stage1 == nil {
			stage1 = &importRes
		}
	}

	var res []*RpmImportUploadWrapper
	uploadArchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 12 * time.Hour,
		StartToCloseTimeout:    24 * time.Hour,
		HeartbeatTimeout:       2 * time.Minute,
		TaskQueue:              importTaskQueue,
	})

	var interimRes []*UploadActivityResult
	err = workflow.ExecuteActivity(uploadArchCtx, c.UploadArchActivity, req.ProjectId, task.ID.String()).Get(ctx, &interimRes)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}
	for _, ires := range interimRes {
		res = append(res, &RpmImportUploadWrapper{
			Upload: ires,
			TaskID: task.ID.String(),
		})
	}

	for _, result := range res {
		if result.Upload.Skip {
			continue
		}
		err = c.db.AttachTaskToBuild(stage1.Build.ID.String(), result.Upload.Subtask.ID.String())
		if err != nil {
			err = status.Errorf(codes.Internal, "could not attach task to build: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}
	}

	yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue: "yumrepofs",
	})
	updateRepoRequest := &UpdateRepoRequest{
		ProjectID:        req.ProjectId,
		BuildIDs:         []string{stage1.Build.ID.String()},
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

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	ret.RepoChanges = updateRepoTask
	return &ret, nil
}

func (c *Controller) RpmImportActivity(ctx context.Context, req *peridotpb.RpmImportRequest, taskID string, setTaskStatus bool, stage1 *RpmImportActivityTaskStage1) (*RpmImportActivityTaskStage1, error) {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(4 * time.Second)
		}
	}()

	var buf bytes.Buffer
	bts, err := c.storage.ReadObject(req.Rpms)
	if err != nil {
		return nil, fmt.Errorf("failed to read RPMs tarball: %v", err)
	}
	buf.Write(bts)

	var rpms []*rpm.Package
	rpmBufs := map[string][]byte{}

	if strings.HasSuffix(req.Rpms, ".tar") {
		c.log.Infof("Reading tar: %s", req.Rpms)

		tr := tar.NewReader(&buf)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			var nBuf bytes.Buffer
			if _, err := io.Copy(&nBuf, tr); err != nil {
				return nil, err
			}
			c.log.Infof("Detected RPM: %s", hdr.Name)
			rpmBufs[hdr.Name] = nBuf.Bytes()
		}
		for _, b := range rpmBufs {
			p, err := rpm.Read(bytes.NewBuffer(b))
			if err != nil {
				return nil, err
			}
			rpms = append(rpms, p)
		}
	} else {
		c.log.Infof("Reading RPM: %s", req.Rpms)
		p, err := rpm.Read(&buf)
		if err != nil {
			return nil, err
		}
		rpms = append(rpms, p)

		realName := p.String() + ".rpm"
		if p.SourceRPM() == "" && p.Architecture() == "i686" {
			realName = strings.ReplaceAll(realName, ".i686", ".src")
		}
		rpmBufs[realName] = bts
	}

	var nvr string
	for _, rpmObj := range rpms {
		realNvr := rpmObj.String()
		if rpmObj.SourceRPM() == "" && rpmObj.Architecture() == "i686" {
			realNvr = strings.ReplaceAll(realNvr, ".i686", ".src")
		}
		if nvr == "" {
			nvr = rpmObj.SourceRPM()
			if nvr == "" && rpmObj.Architecture() == "i686" {
				nvr = realNvr
			}

			break
		} else {
			if nvr != rpmObj.SourceRPM() && nvr != fmt.Sprintf("%s.rpm", realNvr) {
				return nil, fmt.Errorf("only include RPMs from one package")
			}
		}
	}
	if !rpmutils.NVR().MatchString(nvr) {
		return nil, fmt.Errorf("invalid SNVR: %s", nvr)
	}

	var nvrMatch []string
	if rpmutils.NVRUnusualRelease().MatchString(nvr) {
		nvrMatch = rpmutils.NVRUnusualRelease().FindStringSubmatch(nvr)
	} else {
		nvrMatch = rpmutils.NVR().FindStringSubmatch(nvr)
	}
	srcNvra := fmt.Sprintf("%s-%s-%s.src", nvrMatch[1], nvrMatch[2], nvrMatch[3])

	beginTx, err := c.db.Begin()
	if err != nil {
		c.log.Errorf("failed to start transaction: %v", err)
		return nil, fmt.Errorf("failed to begin transaction")
	}
	tx := c.db.UseTransaction(beginTx)

	exists, err := tx.NVRAExists(srcNvra)
	if err != nil {
		c.log.Errorf("failed to check if NVRA exists: %v", err)
		return nil, fmt.Errorf("failed to check if NVRA exists")
	}
	shouldLock := true
	if exists && !req.ForceOverride {
		return nil, fmt.Errorf("NVRA already exists")
	}
	if exists && req.ForceOverride {
		shouldLock = false
	}

	pkgs, err := c.db.GetPackagesInProject(
		&peridotpb.PackageFilters{
			NameExact: wrapperspb.String(nvrMatch[1]),
		},
		req.ProjectId,
		0,
		1,
	)
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, utils.CouldNotRetrieveObjects
	}
	pkg := pkgs[0]

	metadataAnyPb, err := anypb.New(&peridotpb.PackageOperationMetadata{
		PackageName: pkg.Name,
	})
	if err != nil {
		return nil, err
	}
	err = tx.SetTaskMetadata(taskID, metadataAnyPb)
	if err != nil {
		c.log.Errorf("could not set task metadata: %v", err)
		return nil, status.Error(codes.Internal, "could not set task metadata")
	}

	var build *models.Build
	if stage1 == nil {
		var packageVersionId string
		packageVersionId, err = tx.GetPackageVersionId(pkg.ID.String(), nvrMatch[2], nvrMatch[3])
		if err != nil {
			if err == sql.ErrNoRows {
				packageVersionId, err = tx.CreatePackageVersion(pkg.ID.String(), nvrMatch[2], nvrMatch[3])
				if err != nil {
					err = status.Errorf(codes.Internal, "could not create package version: %v", err)
					return nil, err
				}
			} else {
				err = status.Errorf(codes.Internal, "could not get package version id: %v", err)
				return nil, err
			}
		}

		// todo(mustafa): Add published check, as well as limitations for overriding existing versions
		// TODO URGENT: Don't allow nondeterministic behavior regarding versions
		err = tx.AttachPackageVersion(req.ProjectId, pkg.ID.String(), packageVersionId, false)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not attach package version: %v", err)
			return nil, err
		}

		build, err = tx.CreateBuild(pkg.ID.String(), packageVersionId, taskID, req.ProjectId)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not create build")
			return nil, err
		}
	} else {
		build = stage1.Build
	}

	targetDir := filepath.Join(rpmbuild.GetCloneDirectory(), "RPMS")
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		err = status.Errorf(codes.Internal, "could not create target directory: %v", err)
		return nil, err
	}
	for k, v := range rpmBufs {
		err = ioutil.WriteFile(filepath.Join(targetDir, k), v, 0644)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not write RPM: %v", err)
			return nil, err
		}
	}

	if shouldLock {
		err = tx.LockNVRA(srcNvra)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not lock NVRA: %v", err)
			return nil, err
		}
	}

	if setTaskStatus {
		err = tx.SetTaskStatus(taskID, peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not set task status: %v", err)
			return nil, err
		}
	}

	err = beginTx.Commit()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not commit transaction: %v", err)
	}

	return &RpmImportActivityTaskStage1{
		Build: build,
	}, nil
}
