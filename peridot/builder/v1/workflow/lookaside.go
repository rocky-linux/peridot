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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"strings"
	"time"
)

func (c *Controller) LookasideFileUploadWorkflow(ctx workflow.Context, req *peridotpb.LookasideFileUploadRequest, task *models.Task) (*peridotpb.LookasideFileUploadTask, error) {
	ret := &peridotpb.LookasideFileUploadTask{}
	deferTask, errorDetails, err := c.commonCreateTask(task, ret)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	uploadTaskQueue, cleanupWorker, err := c.provisionWorker(ctx, &ProvisionWorkerRequest{
		TaskId:       task.ID.String(),
		ParentTaskId: task.ParentTaskId,
		Purpose:      "lookaside",
		Arch:         "noarch",
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	defer cleanupWorker()

	uploadCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Hour,
		HeartbeatTimeout:    20 * time.Second,
		TaskQueue:           uploadTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err = workflow.ExecuteActivity(uploadCtx, c.LookasideFileUploadActivity, req, task.ID.String()).Get(ctx, ret)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return ret, nil
}

func (c *Controller) LookasideFileUploadActivity(ctx context.Context, req *peridotpb.LookasideFileUploadRequest, taskID string) (*peridotpb.LookasideFileUploadTask, error) {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(4 * time.Second)
		}
	}()

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

	exists, err := c.storage.Exists(sha256Sum)
	if err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			return nil, err
		}
	}
	if exists {
		return &peridotpb.LookasideFileUploadTask{
			Digest: sha256Sum,
		}, nil
	}

	_, err = c.storage.PutObjectBytes(sha256Sum, base64DecodedFile)
	if err != nil {
		return nil, err
	}

	return &peridotpb.LookasideFileUploadTask{
		Digest: sha256Sum,
	}, nil
}
