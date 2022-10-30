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
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"log"
	apolloconnector "peridot.resf.org/apollo/db/connector"
	"peridot.resf.org/apollo/worker"
	commonpb "peridot.resf.org/common"
	"peridot.resf.org/temporalutils"
	"peridot.resf.org/utils"
)

var root = &cobra.Command{
	Use: "apollostarter",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	cnf.DefaultPort = 31209

	cnf.DatabaseName = utils.Pointer[string]("apollo")
	cnf.Name = "apollostarter"

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

	w, err := worker.NewWorker(&worker.NewWorkerInput{
		Temporal:  c,
		Database:  db,
		TaskQueue: "apollo-v1-main-queue",
	})
	defer w.Client.Close()

	// Poll Red Hat for new CVEs and advisories every two hours
	cveWfOpts := client.StartWorkflowOptions{
		ID:           "cron_cve_mirror",
		TaskQueue:    w.TaskQueue,
		CronSchedule: "0 */2 * * *",
	}
	_, err = w.Client.ExecuteWorkflow(context.Background(), cveWfOpts, w.WorkflowController.PollRedHatCVEsWorkflow)
	if err != nil {
		log.Fatalf("unable to start cve workflow: %v", err)
	}
	errataWfOpts := client.StartWorkflowOptions{
		ID:           "cron_errata_mirror",
		TaskQueue:    w.TaskQueue,
		CronSchedule: "0 */2 * * *",
	}
	_, err = w.Client.ExecuteWorkflow(context.Background(), errataWfOpts, w.WorkflowController.PollRedHatErrataWorkflow)
	if err != nil {
		log.Fatalf("unable to start errata workflow: %v", err)
	}

	// Poll unresolved CVE status and update every hour
	cveStatusWfOpts := client.StartWorkflowOptions{
		ID:           "cron_cve_status",
		TaskQueue:    w.TaskQueue,
		CronSchedule: "0 */1 * * *",
	}
	_, err = w.Client.ExecuteWorkflow(context.Background(), cveStatusWfOpts, w.WorkflowController.UpdateCVEStateWorkflow)
	if err != nil {
		log.Fatalf("unable to start cve status workflow: %v", err)
	}

	// Check if CVE is fixed downstream every 10 minutes
	cveDownstreamWfOpts := client.StartWorkflowOptions{
		ID:           "cron_cve_downstream",
		TaskQueue:    w.TaskQueue,
		CronSchedule: "*/10 * * * *",
	}
	_, err = w.Client.ExecuteWorkflow(context.Background(), cveDownstreamWfOpts, w.WorkflowController.DownstreamCVECheckWorkflow)
	if err != nil {
		log.Fatalf("unable to start cve downstream workflow: %v", err)
	}

	// Auto create advisory for fixed CVEs every 30 minutes
	cveAdvisoryWfOpts := client.StartWorkflowOptions{
		ID:           "cron_cve_advisory",
		TaskQueue:    w.TaskQueue,
		CronSchedule: "*/10 * * * *",
	}
	_, err = w.Client.ExecuteWorkflow(context.Background(), cveAdvisoryWfOpts, w.WorkflowController.AutoCreateAdvisoryWorkflow)
	if err != nil {
		log.Fatalf("unable to start cve advisory workflow: %v", err)
	}

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
}

func main() {
	utils.Main()
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
