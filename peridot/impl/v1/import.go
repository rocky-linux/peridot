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
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"peridot.resf.org/peridot/builder/v1/workflow"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (s *Server) ListImports(ctx context.Context, req *peridotpb.ListImportsRequest) (*peridotpb.ListImportsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	imports, err := s.db.ListImports(req.ProjectId, page, limit)
	if err != nil {
		s.log.Errorf("could not list imports: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	var total int64
	if len(imports) > 0 {
		total = imports[0].Total
	} else {
		total, err = s.db.ImportCountInProject(req.ProjectId)
		if err != nil {
			s.log.Errorf("could not count imports: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	importsPb, err := imports.ToProto()
	if err != nil {
		s.log.Errorf("could not convert imports to proto: %v", err)
		return nil, utils.InternalError
	}

	return &peridotpb.ListImportsResponse{
		Imports: importsPb,
		Total:   total,
		Size:    limit,
		Page:    page,
	}, nil
}

func (s *Server) GetImport(ctx context.Context, req *peridotpb.GetImportRequest) (*peridotpb.GetImportResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	imp, err := s.db.GetImport(req.ProjectId, req.ImportId)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}

	importPb, err := imp.ToProto()
	if err != nil {
		s.log.Errorf("could not convert import to proto: %v", err)
		return nil, utils.InternalError
	}

	return &peridotpb.GetImportResponse{
		Import: importPb,
	}, nil
}

func (s *Server) ListImportBatches(ctx context.Context, req *peridotpb.ListImportBatchesRequest) (*peridotpb.ListImportBatchesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	imports, err := s.db.ListImportBatches(req.ProjectId, nil, page, limit)
	if err != nil {
		s.log.Errorf("could not list imports: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	var total int64
	if len(imports) > 0 {
		total = imports[0].Total
	} else {
		total, err = s.db.ImportBatchCountInProject(req.ProjectId)
		if err != nil {
			s.log.Errorf("could not count imports: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	return &peridotpb.ListImportBatchesResponse{
		ImportBatches: imports.ToProto(),
		Total:         total,
		Size:          limit,
		Page:          page,
	}, nil
}

func (s *Server) GetImportBatch(ctx context.Context, req *peridotpb.GetImportBatchRequest) (*peridotpb.GetImportBatchResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	imp, err := s.db.GetImportBatch(req.ProjectId, req.ImportBatchId, req.Filter, page, limit)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}
	var total int64
	if len(imp) > 0 {
		total = imp[0].Total
	} else {
		total, err = s.db.BuildsInBatchCount(req.ProjectId, req.ImportBatchId)
		if err != nil {
			s.log.Errorf("could not count builds in batch: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	listed, err := s.db.ListImportBatches(req.ProjectId, &req.ImportBatchId, 0, 1)
	if err != nil || len(listed) == 0 {
		if err != nil {
			s.log.Errorf("could not list import batches: %v", err)
		}
		return nil, utils.CouldNotRetrieveObjects
	}

	importsPb, err := imp.ToProto()
	if err != nil {
		s.log.Errorf("could not convert imports to proto: %v", err)
		return nil, utils.InternalError
	}

	return &peridotpb.GetImportBatchResponse{
		Imports:   importsPb,
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

func (s *Server) ImportPackage(ctx context.Context, req *peridotpb.ImportPackageRequest) (*peridotpb.AsyncTask, error) {
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
		s.log.Errorf("could not list projects in ImportPackage: %v", err)
		return nil, utils.InternalError
	}
	if len(projects) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "project %s does not exist", req.ProjectId)
	}

	filters := &peridotpb.PackageFilters{}
	switch p := req.Package.(type) {
	case *peridotpb.ImportPackageRequest_PackageId:
		filters.Id = p.PackageId
	case *peridotpb.ImportPackageRequest_PackageName:
		filters.NameExact = p.PackageName
	}

	pkgs, err := s.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
	if len(pkgs) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "package is not enabled in project %s", req.ProjectId)
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

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_IMPORT, &req.ProjectId, nil)
	if err != nil {
		s.log.Errorf("could not create import task in ImportPackage: %v", err)
		return nil, status.Error(codes.InvalidArgument, "could not create import task")
	}

	metadataAnyPb, err := anypb.New(&peridotpb.PackageOperationMetadata{
		PackageName: pkgs[0].Name,
	})
	if err != nil {
		return nil, err
	}
	err = tx.SetTaskMetadata(task.ID.String(), metadataAnyPb)
	if err != nil {
		return nil, err
	}

	taskProto, err := task.ToProto(false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not marshal task: %v", err)
	}

	imp, err := tx.CreateImport(workflow.GetTargetScmUrl(&projects[0], pkgs[0].Name, "rpms"), task.ID.String(), pkgs[0].ID.String(), req.ProjectId)
	if err != nil {
		s.log.Errorf("could not create import in ImportPackage: %v", err)
		return nil, status.Errorf(codes.Internal, "could not create import")
	}

	rollback = false
	err = beginTx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not save, try again")
	}

	_, err = s.temporal.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        task.ID.String(),
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.ImportPackageWorkflow,
		req,
		task,
		imp,
	)
	if err != nil {
		s.log.Errorf("could not start import workflow in ImportPackage: %v", err)
		_ = s.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_FAILED)
		return nil, err
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}

func (s *Server) ImportPackageBatch(ctx context.Context, req *peridotpb.ImportPackageBatchRequest) (*peridotpb.ImportPackageBatchResponse, error) {
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
		s.log.Errorf("could not list projects: %v", err)
		return nil, utils.InternalError
	}
	if len(projects) != 1 {
		return nil, status.Errorf(codes.InvalidArgument, "project %s does not exist", req.ProjectId)
	}

	pkgs, err := s.db.GetPackagesInProject(&peridotpb.PackageFilters{}, req.ProjectId, 0, -1)
	if err != nil {
		s.log.Errorf("could not list packages: %v", err)
		return nil, utils.InternalError
	}
	for _, importReq := range req.Imports {
		isEnabled := false
		for _, pkg := range pkgs {
			switch p := importReq.Package.(type) {
			case *peridotpb.ImportPackageRequest_PackageId:
				if p.PackageId.Value == pkg.ID.String() {
					isEnabled = true
				}
			case *peridotpb.ImportPackageRequest_PackageName:
				if p.PackageName.Value == pkg.Name {
					isEnabled = true
				}
			}
			if isEnabled {
				break
			}
		}

		if !isEnabled {
			var pkgIdentifier string
			switch p := importReq.Package.(type) {
			case *peridotpb.ImportPackageRequest_PackageId:
				pkgIdentifier = p.PackageId.Value
			case *peridotpb.ImportPackageRequest_PackageName:
				pkgIdentifier = p.PackageName.Value
			}
			return nil, status.Errorf(codes.InvalidArgument, "package %s is not enabled in project %s", pkgIdentifier, req.ProjectId)
		}
	}

	importBatchId, err := s.db.CreateImportBatch(req.ProjectId)
	if err != nil {
		s.log.Errorf("could not create import batch in ImportPackageBatch: %v", err)
		return nil, status.Error(codes.Internal, "could not create import batch")
	}

	_, err = s.temporal.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.ImportPackageBatchWorkflow,
		req,
		importBatchId,
		user,
	)
	if err != nil {
		s.log.Errorf("could not start import batch workflow in ImportPackageBatch: %v", err)
		return nil, status.Error(codes.Internal, "could not start import batch workflow")
	}

	return &peridotpb.ImportPackageBatchResponse{
		ImportBatchId: importBatchId,
	}, nil
}
