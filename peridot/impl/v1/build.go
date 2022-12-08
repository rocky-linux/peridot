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

package peridotimplv1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

func (s *Server) ListBuilds(ctx context.Context, req *peridotpb.ListBuildsRequest) (*peridotpb.ListBuildsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	builds, err := s.db.ListBuilds(req.Filters, req.ProjectId, page, limit)
	if err != nil {
		s.log.Errorf("could not list builds: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	var total int64
	if len(builds) > 0 {
		total = builds[0].Total
	} else {
		total, err = s.db.BuildCountInProject(req.ProjectId)
		if err != nil {
			s.log.Errorf("could not count builds: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	buildsPb, err := builds.ToProto()
	if err != nil {
		s.log.Errorf("could not convert builds to proto: %v", err)
		return nil, utils.InternalError
	}

	return &peridotpb.ListBuildsResponse{
		Builds: buildsPb,
		Total:  total,
		Size:   limit,
		Page:   page,
	}, nil
}

func (s *Server) GetBuild(ctx context.Context, req *peridotpb.GetBuildRequest) (*peridotpb.GetBuildResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	imp, err := s.db.GetBuild(req.ProjectId, req.BuildId)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}

	buildPb, err := imp.ToProto()
	if err != nil {
		s.log.Errorf("could not convert build to proto: %v", err)
		return nil, utils.InternalError
	}

	return &peridotpb.GetBuildResponse{
		Build: buildPb,
	}, nil
}

func (s *Server) ListBuildBatches(ctx context.Context, req *peridotpb.ListBuildBatchesRequest) (*peridotpb.ListBuildBatchesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	builds, err := s.db.ListBuildBatches(req.ProjectId, nil, page, limit)
	if err != nil {
		s.log.Errorf("could not list builds: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	var total int64
	if len(builds) > 0 {
		total = builds[0].Total
	} else {
		total, err = s.db.BuildBatchCountInProject(req.ProjectId)
		if err != nil {
			s.log.Errorf("could not count builds: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	return &peridotpb.ListBuildBatchesResponse{
		BuildBatches: builds.ToProto(),
		Total:        total,
		Size:         limit,
		Page:         page,
	}, nil
}

func (s *Server) GetBuildBatch(ctx context.Context, req *peridotpb.GetBuildBatchRequest) (*peridotpb.GetBuildBatchResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	imp, err := s.db.GetBuildBatch(req.ProjectId, req.BuildBatchId, req.Filter, page, limit)
	if err != nil {
		s.log.Errorf("could not get build batch: %v", err)
		return nil, utils.CouldNotFindObject
	}
	var total int64
	if len(imp) > 0 {
		total = imp[0].Total
	} else {
		total, err = s.db.BuildsInBatchCount(req.ProjectId, req.BuildBatchId)
		if err != nil {
			s.log.Errorf("could not count builds in batch: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	listed, err := s.db.ListBuildBatches(req.ProjectId, &req.BuildBatchId, 0, 1)
	if err != nil || len(listed) == 0 {
		if err != nil {
			s.log.Errorf("could not list build batches: %v", err)
		}
		return nil, utils.CouldNotRetrieveObjects
	}

	buildsPb, err := imp.ToProto()
	if err != nil {
		s.log.Errorf("could not convert builds to proto: %v", err)
		return nil, utils.InternalError
	}

	return &peridotpb.GetBuildBatchResponse{
		Builds:    buildsPb,
		Pending:   listed[0].Pending,
		Running:   listed[0].Running,
		Succeeded: listed[0].Succeeded,
		Failed:    listed[0].Failed,
		Canceled:  listed[0].Canceled,
		Total:     total,
		Size:      limit,
		Page:      page,
	}, nil
}

func (s *Server) SubmitBuild(ctx context.Context, req *peridotpb.SubmitBuildRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionBuild); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := s.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		s.log.Errorf("could not list projects in SubmitBuild: %v", err)
		return nil, utils.InternalError
	}
	if len(projects) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "project %s does not exist", req.ProjectId)
	}

	filters := &peridotpb.PackageFilters{}
	switch p := req.Package.(type) {
	case *peridotpb.SubmitBuildRequest_PackageId:
		filters.Id = p.PackageId
	case *peridotpb.SubmitBuildRequest_PackageName:
		filters.NameExact = p.PackageName
	}

	pkgs, err := s.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
	if len(pkgs) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "package is not enabled in project %s", req.ProjectId)
	}
	pkg := pkgs[0]

	packageType := pkg.PackageType
	if pkg.PackageTypeOverride.Valid {
		packageType = peridotpb.PackageType(pkg.PackageTypeOverride.Int32)
	}

	if packageType == peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_COMPONENT {
		return nil, status.Errorf(codes.InvalidArgument, "package %s is a module component", pkg.Name)
	}

	revisions, err := s.db.GetLatestImportRevisionsForPackageInProject(pkgs[0].Name, req.ProjectId)
	if err != nil {
		s.log.Errorf("failed to get import revision: %v", err)
		return nil, utils.InternalError
	}
	if len(revisions) == 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "package %s has no import revisions in project %s", pkg.Name, req.ProjectId)
	}

	rollback := true
	beginTx, err := s.db.Begin()
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	defer func() {
		if rollback {
			_ = beginTx.Rollback()
		}
	}()
	tx := s.db.UseTransaction(beginTx)

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_BUILD, &req.ProjectId, nil)
	if err != nil {
		s.log.Errorf("could not create build task in SubmitBuild: %v", err)
		return nil, status.Error(codes.InvalidArgument, "could not create build task")
	}

	metadataAnyPb, err := anypb.New(&peridotpb.PackageOperationMetadata{
		PackageName: pkg.Name,
	})
	if err != nil {
		return nil, err
	}
	err = tx.SetTaskMetadata(task.ID.String(), metadataAnyPb)
	if err != nil {
		s.log.Errorf("could not set task metadata in TriggerBuildFromBatchWorkflow: %v", err)
		return nil, status.Error(codes.Internal, "could not set task metadata")
	}

	var importRevision *models.ImportRevision
	if req.ScmHash == nil {
		// Each import can have multiple revisions
		importRevisions, err := s.db.GetLatestImportRevisionsForPackageInProject(pkg.Name, req.ProjectId)
		if err != nil {
			s.log.Errorf("failed to get import revision: %v", err)
			return nil, status.Error(codes.InvalidArgument, "could not get import revision")
		}

		// Construct the branch this project is targeting
		// Using the constructed branch, we will check if there is a revision for the branch
		// In fork mode, if the project is configured correctly the upstream will have a revision for the branch
		upstreamBranch := fmt.Sprintf("%s%d%s", projects[0].TargetBranchPrefix, projects[0].MajorVersion, projects[0].BranchSuffix.String)
		for _, revision := range importRevisions {
			if revision.ScmBranchName == upstreamBranch {
				importRevision = &*&revision
				break
			}
		}
	} else {
		importRevision, err = s.db.GetImportRevisionByScmHash(req.ScmHash.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				activityErr := fmt.Errorf("import revision not found for scm hash %s", req.ScmHash.Value)
				return nil, activityErr
			}
		}
	}
	if importRevision == nil {
		return nil, errors.New("could not find upstream branch")
	}

	taskProto, err := task.ToProto(true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not marshal task: %v", err)
	}

	// Check if all branches are modular (that means it's only a module component/module)
	allStream := true
	for _, revision := range revisions {
		if !revision.Modular && !strings.Contains(revision.ScmBranchName, "-stream-") {
			allStream = false
		}
	}
	// Force module variant if all branches are modular
	if !req.ModuleVariant && allStream {
		req.ModuleVariant = true
	}

	if (packageType == peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK || packageType == peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK_MODULE || packageType == peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT) && req.ModuleVariant {
		rollback = false
		err = beginTx.Commit()
		if err != nil {
			return nil, status.Error(codes.Internal, "could not save, try again")
		}

		_, err = s.temporal.ExecuteWorkflow(
			context.Background(),
			client.StartWorkflowOptions{
				ID:                  task.ID.String(),
				TaskQueue:           MainTaskQueue,
				WorkflowTaskTimeout: 3 * time.Hour,
			},
			s.temporalWorker.WorkflowController.BuildModuleWorkflow,
			req,
			task,
			&peridotpb.ExtraBuildOptions{},
		)
		if err != nil {
			return nil, err
		}
	}

	if packageType != peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK && packageType != peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT && len(req.Branches) == 0 && !allStream && !req.ModuleVariant {
		build, err := tx.CreateBuild(pkg.ID.String(), importRevision.PackageVersionId, task.ID.String(), req.ProjectId)
		if err != nil {
			s.log.Errorf("could not create build: %v", err)
			return nil, status.Error(codes.InvalidArgument, "could not create build")
		}

		rollback = false
		err = beginTx.Commit()
		if err != nil {
			return nil, status.Error(codes.Internal, "could not save, try again")
		}

		_, err = s.temporal.ExecuteWorkflow(
			context.Background(),
			client.StartWorkflowOptions{
				ID:                  task.ID.String(),
				TaskQueue:           MainTaskQueue,
				WorkflowTaskTimeout: 3 * time.Hour,
			},
			s.temporalWorker.WorkflowController.BuildWorkflow,
			req,
			task,
			&peridotpb.ExtraBuildOptions{
				ReusableBuildId: build.ID.String(),
			},
		)
		if err != nil {
			return nil, err
		}
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}

func (s *Server) SubmitBuildBatch(ctx context.Context, req *peridotpb.SubmitBuildBatchRequest) (*peridotpb.SubmitBuildBatchResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionBuild); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := s.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		s.log.Errorf("could not list projects in SubmitBuild: %v", err)
		return nil, utils.InternalError
	}
	if len(projects) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "project %s does not exist", req.ProjectId)
	}

	pkgs, err := s.db.GetPackagesInProject(&peridotpb.PackageFilters{}, req.ProjectId, 0, -1)
	var newBuildReqs []string
	var newModuleBuildReqs []string
	for _, buildReq := range req.Builds {
		isEnabled := false
		var dbPkg *models.Package
		for _, pkg := range pkgs {
			switch p := buildReq.Package.(type) {
			case *peridotpb.SubmitBuildRequest_PackageId:
				if p.PackageId.Value == pkg.ID.String() {
					isEnabled = true
					dbPkg = &pkg
				}
			case *peridotpb.SubmitBuildRequest_PackageName:
				if p.PackageName.Value == pkg.Name {
					isEnabled = true
					dbPkg = &pkg
				}
			}
			if isEnabled {
				break
			}
		}

		if !isEnabled {
			var pkgIdentifier string
			switch p := buildReq.Package.(type) {
			case *peridotpb.SubmitBuildRequest_PackageId:
				pkgIdentifier = p.PackageId.Value
			case *peridotpb.SubmitBuildRequest_PackageName:
				pkgIdentifier = p.PackageName.Value
			}
			return nil, status.Errorf(codes.InvalidArgument, "package %s is not enabled in project %s", pkgIdentifier, req.ProjectId)
		} else {
			packageType := dbPkg.PackageType
			if dbPkg.PackageTypeOverride.Valid {
				packageType = peridotpb.PackageType(dbPkg.PackageTypeOverride.Int32)
			}

			// Skip module components as they can't build on their own
			if packageType != peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_COMPONENT {
				if packageType == peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK || packageType == peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK_MODULE || packageType == peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT {
					newModuleBuildReqs = append(newModuleBuildReqs, dbPkg.Name)
				} else {
					newBuildReqs = append(newBuildReqs, dbPkg.Name)
				}
			}
		}
	}

	buildBatchId, err := s.db.CreateBuildBatch(req.ProjectId)
	if err != nil {
		s.log.Errorf("could not create build batch in SubmitBuildBatch: %v", err)
		return nil, status.Error(codes.Internal, "could not create build batch")
	}

	_, err = s.temporal.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue:           MainTaskQueue,
			WorkflowTaskTimeout: 3 * time.Hour,
		},
		s.temporalWorker.WorkflowController.BuildBatchWorkflow,
		req,
		newBuildReqs,
		newModuleBuildReqs,
		buildBatchId,
		user,
	)
	if err != nil {
		s.log.Errorf("could not start build batch workflow in SubmitBuildBatch: %v", err)
		return nil, status.Error(codes.Internal, "could not start build batch workflow")
	}

	return &peridotpb.SubmitBuildBatchResponse{
		BuildBatchId: buildBatchId,
	}, nil
}

func (s *Server) RpmImport(ctx context.Context, req *peridotpb.RpmImportRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionBuild); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := s.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		s.log.Errorf("could not list projects in SubmitBuild: %v", err)
		return nil, utils.InternalError
	}
	if len(projects) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "project %s does not exist", req.ProjectId)
	}

	rollback := true
	beginTx, err := s.db.Begin()
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	defer func() {
		if rollback {
			_ = beginTx.Rollback()
		}
	}()
	tx := s.db.UseTransaction(beginTx)

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_RPM_IMPORT, &req.ProjectId, nil)
	if err != nil {
		s.log.Errorf("could not create build task in RpmImport: %v", err)
		return nil, status.Error(codes.InvalidArgument, "could not create rpm import task")
	}

	taskProto, err := task.ToProto(true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not marshal task: %v", err)
	}

	rollback = false
	err = beginTx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not save, try again")
	}

	_, err = s.temporal.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.RpmImportWorkflow,
		req,
		task,
	)
	if err != nil {
		s.log.Errorf("could not start rpm import workflow in RpmImport: %v", err)
		return nil, status.Error(codes.Internal, "could not start rpm import workflow")
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}

func (s *Server) RpmLookasideBatchImport(ctx context.Context, req *peridotpb.RpmLookasideBatchImportRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionBuild); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := s.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		s.log.Errorf("could not list projects in RpmLookasideBatchImport: %v", err)
		return nil, utils.InternalError
	}
	if len(projects) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "project %s does not exist", req.ProjectId)
	}

	rollback := true
	beginTx, err := s.db.Begin()
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	defer func() {
		if rollback {
			_ = beginTx.Rollback()
		}
	}()
	tx := s.db.UseTransaction(beginTx)

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_RPM_LOOKASIDE_BATCH_IMPORT, &req.ProjectId, nil)
	if err != nil {
		s.log.Errorf("could not create build task in RpmImport: %v", err)
		return nil, status.Error(codes.InvalidArgument, "could not create rpm import task")
	}

	taskProto, err := task.ToProto(true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not marshal task: %v", err)
	}

	rollback = false
	err = beginTx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not save, try again")
	}

	_, err = s.temporal.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.RpmLookasideBatchImportWorkflow,
		req,
		task,
	)
	if err != nil {
		s.log.Errorf("could not start rpm lookaside batch import workflow in RpmImport: %v", err)
		return nil, status.Error(codes.Internal, "could not start rpm lookaside batch import workflow")
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}
