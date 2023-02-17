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

package peridotadminv1

import (
	"context"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	adminpb "peridot.resf.org/peridot/admin/pb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (s *Server) AddUpdateInformation(ctx context.Context, req *adminpb.AddUpdateInformationRequest) (*peridotpb.AsyncTask, error) {
	if err := req.ValidateAll(); err != nil {
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

	task, err := tx.CreateTask(adminUser, "noarch", peridotpb.TaskType_TASK_TYPE_UPDATEINFO, &req.ProjectId, nil)
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
			TaskQueue: "yumrepofs",
		},
		s.temporalWorker.WorkflowController.UpdateInfoWorkflow,
		req,
		task,
	)
	if err != nil {
		s.log.Errorf("could not start sync workflow in AddUpdateInformation: %v", err)
		_ = s.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_FAILED)
		return nil, err
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}
