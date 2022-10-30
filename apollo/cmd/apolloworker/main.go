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

package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"log"
	apolloconnector "peridot.resf.org/apollo/db/connector"
	"peridot.resf.org/apollo/rherrata"
	"peridot.resf.org/apollo/rhsecurity"
	"peridot.resf.org/apollo/worker"
	"peridot.resf.org/apollo/workflow"
	commonpb "peridot.resf.org/common"
	"peridot.resf.org/temporalutils"
	"peridot.resf.org/utils"
	"sync"
)

var root = &cobra.Command{
	Use: "apolloworker",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	cnf.DefaultPort = 29209

	cnf.DatabaseName = utils.Pointer[string]("apollo")
	cnf.Name = "apolloworker"

	pflags := root.PersistentFlags()
	pflags.String("vendor", "Rocky Enterprise Software Foundation", "Vendor name that is publishing the advisories")

	temporalutils.AddFlags(root.PersistentFlags())
	utils.AddFlags(root.PersistentFlags(), cnf)
}

func mn(_ *cobra.Command, _ []string) {
	c, err := temporalutils.NewClient(client.Options{})
	if err != nil {
		logrus.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	db := apolloconnector.MustAuto()

	options := []workflow.Option{
		workflow.WithSecurityAPI(rhsecurity.NewAPIClient(rhsecurity.NewConfiguration()).DefaultApi),
		workflow.WithErrataAPI(rherrata.NewClient()),
	}

	w, err := worker.NewWorker(
		&worker.NewWorkerInput{
			Temporal:  c,
			Database:  db,
			TaskQueue: "apollo-v1-main-queue",
		},
		options...,
	)
	defer w.Client.Close()

	w.Worker.RegisterWorkflow(w.WorkflowController.AutoCreateAdvisoryWorkflow)
	w.Worker.RegisterWorkflow(w.WorkflowController.DownstreamCVECheckWorkflow)
	w.Worker.RegisterWorkflow(w.WorkflowController.PollRedHatCVEsWorkflow)
	w.Worker.RegisterWorkflow(w.WorkflowController.PollRedHatErrataWorkflow)
	w.Worker.RegisterWorkflow(w.WorkflowController.UpdateCVEStateWorkflow)

	w.Worker.RegisterActivity(w.WorkflowController.AutoCreateAdvisoryActivity)
	w.Worker.RegisterActivity(w.WorkflowController.GetAllShortCodesActivity)
	w.Worker.RegisterActivity(w.WorkflowController.DownstreamCVECheckActivity)
	w.Worker.RegisterActivity(w.WorkflowController.PollCVEProcessShortCodeActivity)
	w.Worker.RegisterActivity(w.WorkflowController.ProcessRedHatErrataShortCodeActivity)
	w.Worker.RegisterActivity(w.WorkflowController.UpdateCVEStateActivity)

	w.Worker.RegisterWorkflow(w.WorkflowController.CollectCVEDataWorkflow)
	w.Worker.RegisterActivity(w.WorkflowController.CollectCVEDataActivity)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		w.Run()
		wg.Done()
	}()

	go func() {
		// only added so we get a health endpoint
		s := utils.NewGRPCServer(
			nil,
			func(r *utils.Register) {
				err := commonpb.RegisterHealthCheckServiceHandlerFromEndpoint(r.Context, r.Mux, r.Endpoint, r.Options)
				if err != nil {
					logrus.Fatalf("could not register health service: %v", err)
				}
			},
			func(r *utils.RegisterServer) {
				commonpb.RegisterHealthCheckServiceServer(r.Server, &utils.HealthServer{})
			},
		)
		s.WaitGroup.Wait()
		wg.Done()
	}()

	wg.Wait()
}

func main() {
	utils.Main()
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
