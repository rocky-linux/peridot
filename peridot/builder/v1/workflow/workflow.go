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
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	serverdb "peridot.resf.org/peridot/db"
	"peridot.resf.org/peridot/db/models"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	"peridot.resf.org/peridot/lookaside"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/plugin"
	"peridot.resf.org/peridot/rpmbuild"
)

const (
	ErrorDomainTasksPeridot   = "tasks.peridot.resf.org"
	ErrorDomainBuildsPeridot  = "builds.peridot.resf.org"
	ErrorDomainImportsPeridot = "imports.peridot.resf.org"

	ErrorReasonInternalError       = "internal-error"
	ErrorReasonCouldNotFindPackage = "could not find specified package"
	ErrorReasonActivityFailed      = "activity failed in asynctask"
)

type Controller struct {
	internalMonBuffer []string

	temporal  client.Client
	db        serverdb.Access
	storage   lookaside.Storage
	log       *logrus.Logger
	rpmbuild  rpmbuild.Access
	plugins   []plugin.Plugin
	mainQueue string
	keykeeper keykeeperpb.KeykeeperServiceClient
	dynamodb  *dynamodb.DynamoDB
	unshared  bool
}

type MonLogger struct {
	TaskID       string
	ParentTaskID string
	Controller   *Controller
}

func NewController(c client.Client, db serverdb.Access, storage lookaside.Storage, mainQueue string, keykeeperClient keykeeperpb.KeykeeperServiceClient, ddb *dynamodb.DynamoDB, plugins ...plugin.Plugin) *Controller {
	return &Controller{
		temporal:  c,
		db:        db,
		storage:   storage,
		log:       logrus.New(),
		rpmbuild:  rpmbuild.New(osfs.New(".")),
		plugins:   plugins,
		mainQueue: mainQueue,
		keykeeper: keykeeperClient,
		dynamodb:  ddb,
	}
}

func setInternalError(errorDetails *peridotpb.TaskErrorDetails, err error) {
	errorDetails.ErrorInfo = &errdetails.ErrorInfo{
		Reason:   ErrorReasonInternalError,
		Domain:   ErrorDomainTasksPeridot,
		Metadata: nil,
	}
	errorDetails.ErrorType = &peridotpb.TaskErrorDetails_DebugInfo{
		DebugInfo: &errdetails.DebugInfo{
			StackEntries: nil,
			Detail:       err.Error(),
		},
	}
}

func setPackageNotFoundError(errorDetails *peridotpb.TaskErrorDetails, projectId string, domain string) {
	errorDetails.ErrorInfo = &errdetails.ErrorInfo{
		Reason:   ErrorReasonCouldNotFindPackage,
		Domain:   domain,
		Metadata: nil,
	}
	errorDetails.ErrorType = &peridotpb.TaskErrorDetails_PreconditionFailure{
		PreconditionFailure: &errdetails.PreconditionFailure{
			Violations: []*errdetails.PreconditionFailure_Violation{
				{
					Type:        "CouldNotFindPackage",
					Subject:     fmt.Sprintf("/v1/projects/%s/packages", projectId),
					Description: "Project does not contain the specified package",
				},
			},
		},
	}
}

func setActivityError(errorDetails *peridotpb.TaskErrorDetails, err error) {
	errorDetails.ErrorInfo = &errdetails.ErrorInfo{
		Reason: ErrorReasonActivityFailed,
		Domain: ErrorDomainTasksPeridot,
		Metadata: map[string]string{
			"activity_message": err.Error(),
		},
	}
}

func (c *Controller) commonCreateTask(task *models.Task, taskResponse proto.Message) (func(), *peridotpb.TaskErrorDetails, error) {
	errorDetails := peridotpb.TaskErrorDetails{}

	err := c.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		setInternalError(&errorDetails, err)
		return func() {}, &errorDetails, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	deferFunc := func() {
		if taskResponse != nil {
			taskResponseAny, err := anypb.New(taskResponse)
			if err != nil {
				c.log.Errorf("could not create anypb for task: %v", err)
				task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED
			} else {
				err = c.db.SetTaskResponse(task.ID.String(), taskResponseAny)
				if err != nil {
					c.log.Errorf("could not set task info: %v", err)
					task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED
				}
			}
		}

		if task.Status == peridotpb.TaskStatus_TASK_STATUS_FAILED {
			if errorDetails.ErrorType != nil {
				switch x := errorDetails.ErrorType.(type) {
				case *peridotpb.TaskErrorDetails_DebugInfo:
					if x.DebugInfo.Detail == "canceled" {
						task.Status = peridotpb.TaskStatus_TASK_STATUS_CANCELED
					}
				}
			}
			if errorDetails.ErrorInfo != nil {
				if errorDetails.ErrorInfo.Metadata["activity_message"] == "canceled" {
					task.Status = peridotpb.TaskStatus_TASK_STATUS_CANCELED
				}
			}

			setError := true
			if errorDetails.ErrorInfo == nil {
				task.Status = peridotpb.TaskStatus_TASK_STATUS_RUNNING
				setError = false
			}

			if setError {
				anyErrorDetails, err := anypb.New(&errorDetails)
				if err == nil {
					err = c.db.SetTaskResponse(task.ID.String(), anyErrorDetails)
					if err != nil {
						c.log.Errorf("could not set error metadata: %v", err)
					}
				}
			}
		}

		err := c.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			c.log.Errorf("could not set task status: %v", err)
		}

		taskDb, err := c.db.GetTask(task.ID.String(), task.ProjectId.String)
		if err != nil {
			c.log.Errorf("could not get task: %v", err)
			return
		}

		if len(taskDb) > 1 {
			for i, taskFromDb := range taskDb {
				if i == 0 {
					continue
				}
				if taskFromDb.Status != peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED && taskFromDb.Status != peridotpb.TaskStatus_TASK_STATUS_CANCELED && taskFromDb.Status != peridotpb.TaskStatus_TASK_STATUS_FAILED {
					_ = c.db.SetTaskStatus(taskFromDb.ID.String(), peridotpb.TaskStatus_TASK_STATUS_FAILED)
				}
			}
		}
	}

	return deferFunc, &errorDetails, nil
}

func (c *Controller) preExecPlugins(activityType string) error {
	for _, p := range c.plugins {
		if err := p.PreExec(&plugin.PreExecArgs{
			ActivityType: activityType,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) postExecPlugins(activityType string) error {
	for _, p := range c.plugins {
		if err := p.PostExec(&plugin.PostExecArgs{
			ActivityType: activityType,
			ResultsDir:   rpmbuild.GetCloneDirectory(),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) logToMon(lines []string, taskId string, parentTaskId string) error {
	if parentTaskId == "" {
		parentTaskId = taskId
	}
	if err := c.db.InsertLogs(lines, taskId, parentTaskId); err != nil {
		return err
	}

	return nil
}
