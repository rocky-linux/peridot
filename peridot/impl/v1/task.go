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
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"peridot.resf.org/peridot/builder/v1/workflow"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

func (s *Server) ListTasks(ctx context.Context, req *peridotpb.ListTasksRequest) (*peridotpb.ListTasksResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	tasks, err := s.db.ListTasks(&req.ProjectId.Value, page, limit)
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	var total int64
	if len(tasks) > 0 {
		total = tasks[0].Total
	} else {
		total, err = s.db.ImportCountInProject(req.ProjectId.Value)
		if err != nil {
			s.log.Errorf("could not count imports: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	var asyncTasks []*peridotpb.AsyncTask
	for _, task := range tasks {
		taskProto, err := task.ToProto(true)
		if err != nil {
			return nil, err
		}
		asyncTasks = append(asyncTasks, &peridotpb.AsyncTask{
			TaskId:   task.ID.String(),
			Subtasks: []*peridotpb.Subtask{taskProto},
			Done:     task.FinishedAt.Valid,
		})
	}

	return &peridotpb.ListTasksResponse{
		Tasks: asyncTasks,
		Total: total,
		Size:  limit,
		Page:  page,
	}, nil
}

func (s *Server) GetTask(ctx context.Context, req *peridotpb.GetTaskRequest) (*peridotpb.GetTaskResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId.Value, PermissionView); err != nil {
		return nil, err
	}

	tasks, err := s.db.GetTask(req.Id, req.ProjectId.Value)
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	if len(tasks) == 0 {
		return nil, utils.CouldNotFindObject
	}

	parentTask := tasks[0]

	tasksProto, err := tasks.ToProto(false)
	if err != nil {
		return nil, err
	}

	return &peridotpb.GetTaskResponse{
		Task: &peridotpb.AsyncTask{
			TaskId:   parentTask.ID.String(),
			Subtasks: tasksProto,
			Done:     parentTask.FinishedAt.Valid,
		},
	}, nil
}

func (s *Server) StreamTaskLogs(req *peridotpb.StreamTaskLogsRequest, stream peridotpb.TaskService_StreamTaskLogsServer) error {
	ctx := stream.Context()

	if err := req.ValidateAll(); err != nil {
		return err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return err
	}

	var taskId *string = nil
	var parentTaskId *string = nil
	if req.Parent {
		parentTaskId = &req.Id
	} else {
		taskId = &req.Id
	}
	_, err := s.db.GetTask(req.Id, req.ProjectId)
	if err != nil {
		s.log.Errorf("error getting task: %s", err)
		return utils.InternalError
	}

	offset := int64(0)

	for {
		if ctx.Err() == context.Canceled {
			return nil
		}
		logs, err := s.db.GetLogsForTaskIdOrParentTaskId(taskId, parentTaskId, &offset)
		if err != nil {
			s.log.Errorf("error getting logs for task %s: %s", req.Id, err)
			return utils.InternalError
		}
		if len(logs) > 0 {
			offset += int64(len(logs))

			var reducedLines []string
			for _, line := range logs {
				reducedLines = append(reducedLines, line...)
			}
			err = stream.Send(&httpbody.HttpBody{
				ContentType: "text/plain",
				Data:        []byte(strings.Join(reducedLines, "\n")),
			})
			if err != nil {
				return err
			}
		}

		task, err := s.db.GetTask(req.Id, req.ProjectId)
		if err != nil {
			s.log.Errorf("error getting task: %s", err)
			return utils.InternalError
		}
		if task[0].FinishedAt.Valid {
			return nil
		}
		time.Sleep(4 * time.Second)
	}
}

func (s *Server) CancelTask(ctx context.Context, req *peridotpb.CancelTaskRequest) (*peridotpb.CancelTaskResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionBuild); err != nil {
		return nil, err
	}

	tasks, err := s.db.GetTask(req.Id, req.ProjectId)
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	if len(tasks) == 0 {
		return nil, utils.CouldNotFindObject
	}

	for _, task := range tasks {
		if task.Type == peridotpb.TaskType_TASK_TYPE_WORKER_PROVISION {
			if task.Metadata.Valid {
				anyPb := &anypb.Any{}
				err := protojson.Unmarshal(task.Metadata.JSONText, anyPb)
				if err != nil {
					s.log.Errorf("error unmarshalling task metadata: %s", err)
					return nil, utils.InternalError
				}
				provisionWorkerMetadata := &peridotpb.ProvisionWorkerMetadata{}
				err = anyPb.UnmarshalTo(provisionWorkerMetadata)
				if err != nil {
					s.log.Errorf("error unmarshalling task metadata: %s", err)
					return nil, utils.InternalError
				}

				_, err = s.temporal.ExecuteWorkflow(
					context.Background(),
					client.StartWorkflowOptions{
						TaskQueue: MainTaskQueue,
					},
					s.temporalWorker.WorkflowController.DestroyWorkerWorkflow,
					&workflow.ProvisionWorkerRequest{
						TaskId:    provisionWorkerMetadata.TaskId,
						Purpose:   provisionWorkerMetadata.Purpose,
						ProjectId: req.ProjectId,
					},
				)
				if err != nil {
					s.log.Errorf("error destroying worker: %s", err)
					return nil, utils.InternalError
				}
			}
		}
	}

	task := tasks[0]

	err = s.temporal.CancelWorkflow(context.Background(), task.ID.String(), "")
	if err != nil {
		s.log.Errorf("error canceling workflow: %s", err)
		return nil, utils.InternalError
	}

	return &peridotpb.CancelTaskResponse{}, nil
}
