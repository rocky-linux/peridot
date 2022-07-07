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

package builderv1

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"peridot.resf.org/peridot/builder/v1/workflow"
	serverdb "peridot.resf.org/peridot/db"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	"peridot.resf.org/peridot/lookaside/s3"
	"peridot.resf.org/peridot/plugin"
	"time"
)

type Worker struct {
	Client             client.Client
	TaskQueue          string
	WorkflowController *workflow.Controller
	Worker             worker.Worker

	log *logrus.Logger
}

type ExtraReq struct {
	KeykeeperClient keykeeperpb.KeykeeperServiceClient
	DynamoDB        *dynamodb.DynamoDB
}

func NewWorker(db serverdb.Access, c client.Client, taskQueue string, extraReq *ExtraReq, plugins ...plugin.Plugin) (*Worker, error) {
	log := logrus.New()
	storage, err := s3.New(osfs.New("/"))
	if err != nil {
		return nil, err
	}

	if extraReq == nil {
		extraReq = &ExtraReq{}
	}

	return &Worker{
		Client:             c,
		TaskQueue:          taskQueue,
		WorkflowController: workflow.NewController(c, db, storage, taskQueue, extraReq.KeykeeperClient, extraReq.DynamoDB, plugins...),
		Worker: worker.New(c, taskQueue, worker.Options{
			DeadlockDetectionTimeout: 15 * time.Minute,
		}),
		log: log,
	}, nil
}

func (w *Worker) Run() {
	err := w.Worker.Run(worker.InterruptCh())
	if err != nil {
		w.log.Fatalf("could not run worker: %v", err)
	}
}
