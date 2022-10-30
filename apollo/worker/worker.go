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

package worker

import (
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	apollodb "peridot.resf.org/apollo/db"
	"peridot.resf.org/apollo/workflow"
)

type Worker struct {
	Client             client.Client
	TaskQueue          string
	WorkflowController *workflow.Controller
	Worker             worker.Worker

	log *logrus.Logger
}

type NewWorkerInput struct {
	Temporal  client.Client
	Database  apollodb.Access
	TaskQueue string
}

func NewWorker(input *NewWorkerInput, workflowOpts ...workflow.Option) (*Worker, error) {
	log := logrus.New()

	controller, err := workflow.NewController(&workflow.NewControllerInput{
		Temporal:  input.Temporal,
		Database:  input.Database,
		MainQueue: input.TaskQueue,
	}, workflowOpts...)
	if err != nil {
		return nil, err
	}

	return &Worker{
		Client:             input.Temporal,
		TaskQueue:          input.TaskQueue,
		WorkflowController: controller,
		Worker:             worker.New(input.Temporal, input.TaskQueue, worker.Options{}),
		log:                log,
	}, nil
}

func (w *Worker) Run() {
	err := w.Worker.Run(worker.InterruptCh())
	if err != nil {
		w.log.Fatalf("could not run worker: %v", err)
	}
}
