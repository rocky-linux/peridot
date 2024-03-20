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
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"strings"
)

func (s *Server) CreateProject(ctx context.Context, req *peridotpb.CreateProjectRequest) (*peridotpb.CreateProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := req.Project.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectGlobal, ObjectIdPeridot, PermissionManage); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	beginTx, err := s.db.Begin()
	if err != nil {
		s.log.Errorf("CreateProject: beginTx: %v", err)
		return nil, utils.InternalError
	}
	tx := s.db.UseTransaction(beginTx)

	p, err := tx.CreateProject(req.GetProject())
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, status.Error(codes.AlreadyExists, "project with name already exists")
		}
		s.log.Errorf("CreateProject: %v", err)
		return nil, status.Error(codes.Internal, "failed to create project")
	}

	_, err = s.authz.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: ObjectProject,
						ObjectId:   p.ID.String(),
					},
					Relation: "manager",
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: SubjectUser,
							ObjectId:   user.ID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		_ = beginTx.Rollback()
		s.log.Errorf("CreateProject: authz-create-relationships: %v", err)
		return nil, status.Error(codes.Internal, "failed to create project")
	}

	_, err = tx.CreateRepositoryWithPackages("all", p.ID.String(), false, []string{})
	if err != nil {
		_ = beginTx.Rollback()
		s.log.Errorf("CreateProject: db-create-repository: %v", err)
		return nil, status.Error(codes.Internal, "failed to create project")
	}

	err = beginTx.Commit()
	if err != nil {
		s.log.Errorf("CreateProject: commit: %v", err)
		return nil, status.Error(codes.Internal, "failed to create project")
	}

	return &peridotpb.CreateProjectResponse{
		Project: p.ToProto(),
	}, nil
}

func (s *Server) UpdateProject(ctx context.Context, req *peridotpb.UpdateProjectRequest) (*peridotpb.UpdateProjectResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := req.Project.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionManage); err != nil {
		return nil, err
	}

	p, err := s.db.UpdateProject(req.ProjectId, req.GetProject())
	if err != nil {
		s.log.Errorf("UpdateProject: %v", err)
		return nil, status.Error(codes.Internal, "failed to update project")
	}

	return &peridotpb.UpdateProjectResponse{
		Project: p.ToProto(),
	}, nil
}

func (s *Server) ListProjects(ctx context.Context, req *peridotpb.ListProjectsRequest) (*peridotpb.ListProjectsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	resources, err := s.lookupResources(ctx, ObjectProject, PermissionView)
	if err != nil {
		return nil, err
	}
	// Subject doesn't have access to any projects
	if len(resources) == 0 {
		return &peridotpb.ListProjectsResponse{}, nil
	}

	projects, err := s.db.ListProjects(&peridotpb.ProjectFilters{
		Ids: utils.Take[string](resources, "global"),
	})
	if err != nil {
		s.log.Errorf("could not list projects: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}

	return &peridotpb.ListProjectsResponse{
		Projects: projects.ToProto(),
		Total:    int64(len(projects)),
		Current:  int64(len(projects)),
		Page:     0,
	}, nil
}

func (s *Server) GetProject(ctx context.Context, req *peridotpb.GetProjectRequest) (*peridotpb.GetProjectResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.Id.Value, PermissionView); err != nil {
		return nil, err
	}

	projects, err := s.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.Id.Value),
	})
	if err != nil {
		return nil, utils.CouldNotFindObject
	}
	if len(projects) == 0 {
		return nil, utils.CouldNotFindObject
	}

	return &peridotpb.GetProjectResponse{
		Project: projects[0].ToProto(),
	}, nil
}

func (s *Server) ListRepositories(ctx context.Context, req *peridotpb.ListRepositoriesRequest) (*peridotpb.ListRepositoriesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionView); err != nil {
		return nil, err
	}

	repos, err := s.db.FindRepositoriesForProject(req.ProjectId.Value, nil, false)
	if err != nil {
		s.log.Errorf("could not list repositories: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}

	return &peridotpb.ListRepositoriesResponse{
		Repositories: repos.ToProto(),
	}, nil
}

func (s *Server) GetRepository(ctx context.Context, req *peridotpb.GetRepositoryRequest) (*peridotpb.GetRepositoryResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionView); err != nil {
		return nil, err
	}

	repos, err := s.db.FindRepositoriesForProject(req.ProjectId.Value, &req.Id.Value, false)
	if err != nil {
		s.log.Errorf("could not list repositories: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	if len(repos) == 0 {
		return nil, utils.CouldNotFindObject
	}

	return &peridotpb.GetRepositoryResponse{
		Repository: repos[0].ToProto(),
	}, nil
}

func (s *Server) ListExternalRepositories(ctx context.Context, req *peridotpb.ListExternalRepositoriesRequest) (*peridotpb.ListExternalRepositoriesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionView); err != nil {
		return nil, err
	}

	repos, err := s.db.GetExternalRepositoriesForProject(req.ProjectId.Value)
	if err != nil {
		s.log.Errorf("could not list repositories: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}

	return &peridotpb.ListExternalRepositoriesResponse {
		Repositories: repos.ToProto(),
	}, nil
}

func (s *Server) GetExternalRepository(ctx context.Context, req *peridotpb.GetExternalRepositoryRequest) (*peridotpb.GetExternalRepositoryResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionView); err != nil {
		return nil, err
	}

	repos, err := s.db.GetExternalRepositoryForProject(req.ProjectId.Value, &req.Id.Value, false)
	if err != nil {
		s.log.Errorf("could not list repositories: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	if len(repos) == 0 {
		return nil, utils.CouldNotFindObject
	}

	return &peridotpb.GetRepositoryResponse{
		Repository: repos[0].ToProto(),
	}, nil
}

func (s *Server) GetProjectCredentials(ctx context.Context, req *peridotpb.GetProjectCredentialsRequest) (*peridotpb.GetProjectCredentialsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionManage); err != nil {
		return nil, err
	}

	creds, err := s.db.GetProjectKeys(req.ProjectId.Value)
	if err != nil && err != sql.ErrNoRows {
		s.log.Errorf("could not get project credentials: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	if err == sql.ErrNoRows {
		return &peridotpb.GetProjectCredentialsResponse{}, nil
	}

	return &peridotpb.GetProjectCredentialsResponse{
		GitlabUsername: wrapperspb.String(creds.GitlabUsername),
	}, nil
}

func (s *Server) SetProjectCredentials(ctx context.Context, req *peridotpb.SetProjectCredentialsRequest) (*peridotpb.SetProjectCredentialsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionManage); err != nil {
		return nil, err
	}

	err := s.db.SetProjectKeys(req.ProjectId.Value, req.GitlabUsername.Value, req.GitlabPassword.Value)
	if err != nil {
		s.log.Errorf("could not set project credentials: %v", err)
		return nil, utils.CouldNotUpdateObject
	}

	return &peridotpb.SetProjectCredentialsResponse{
		GitlabUsername: req.GitlabUsername,
	}, nil
}

func (s *Server) SyncCatalog(ctx context.Context, req *peridotpb.SyncCatalogRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionManage); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
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

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_SYNC_CATALOG, &req.ProjectId.Value, nil)
	if err != nil {
		s.log.Errorf("could not create task: %v", err)
		return nil, utils.InternalError
	}

	taskProto, err := task.ToProto(false)
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
			ID:        task.ID.String(),
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.SyncCatalogWorkflow,
		req,
		task,
	)
	if err != nil {
		s.log.Errorf("could not start sync workflow in SyncCatalog: %v", err)
		_ = s.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_FAILED)
		return nil, err
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}

func (s *Server) CreateHashedRepositories(ctx context.Context, req *peridotpb.CreateHashedRepositoriesRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionManage); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
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

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_CREATE_HASHED_REPOSITORIES, &req.ProjectId.Value, nil)
	if err != nil {
		s.log.Errorf("could not create task: %v", err)
		return nil, utils.InternalError
	}

	taskProto, err := task.ToProto(false)
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
			ID:        task.ID.String(),
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.CreateHashedRepositoriesWorkflow,
		req,
		task,
	)
	if err != nil {
		s.log.Errorf("could not start workflow: %v", err)
		_ = s.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_FAILED)
		return nil, err
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}

func (s *Server) LookasideFileUpload(ctx context.Context, req *peridotpb.LookasideFileUploadRequest) (*peridotpb.LookasideFileUploadResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectGlobal, ObjectIdPeridot, PermissionManage); err != nil {
		return nil, err
	}

	base64DecodedFile, err := base64.StdEncoding.DecodeString(req.File)
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()
	_, err = hasher.Write(base64DecodedFile)
	if err != nil {
		return nil, err
	}
	sha256Sum := hex.EncodeToString(hasher.Sum(nil))

	exists, err := s.storage.Exists(sha256Sum)
	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			return nil, err
		}
	}
	if exists {
		return &peridotpb.LookasideFileUploadResponse{
			Digest: sha256Sum,
		}, nil
	}

	_, err = s.storage.PutObjectBytes(sha256Sum, base64DecodedFile)
	if err != nil {
		return nil, err
	}

	return &peridotpb.LookasideFileUploadResponse{
		Digest: sha256Sum,
	}, nil
}

func (s *Server) CloneSwap(ctx context.Context, req *peridotpb.CloneSwapRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.TargetProjectId.Value, PermissionManage); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.SrcProjectId.Value, PermissionView); err != nil {
		return nil, err
	}
	user, err := utils.UserFromContext(ctx)
	if err != nil {
		return nil, err
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

	task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_CLONE_SWAP, &req.TargetProjectId.Value, nil)
	if err != nil {
		s.log.Errorf("could not create task: %v", err)
		return nil, utils.InternalError
	}

	taskProto, err := task.ToProto(false)
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
			ID:        task.ID.String(),
			TaskQueue: MainTaskQueue,
		},
		s.temporalWorker.WorkflowController.CloneSwapWorkflow,
		req,
		task,
	)
	if err != nil {
		s.log.Errorf("could not start workflow: %v", err)
		_ = s.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_FAILED)
		return nil, err
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}
