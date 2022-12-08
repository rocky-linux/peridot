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
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"log"
	"os"
	commonpb "peridot.resf.org/common"
	builderv1 "peridot.resf.org/peridot/builder/v1"
	peridotcommon "peridot.resf.org/peridot/common"
	serverconnector "peridot.resf.org/peridot/db/connector"
	peridotimplv1 "peridot.resf.org/peridot/impl/v1"
	"peridot.resf.org/temporalutils"
	"peridot.resf.org/utils"
	"strings"
	"sync"
)

var root = &cobra.Command{
	Use: "peridotephemeral",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	cnf.DefaultPort = 15201

	dname := "peridot"
	cnf.DatabaseName = &dname
	cnf.Name = "peridotephemeral"

	root.PersistentFlags().String("builder-oci-image-x86_64", "", "Builder OCI image to spawn (x86_64)")
	root.PersistentFlags().String("builder-oci-image-aarch64", "", "Builder OCI image to spawn (aarch64)")
	root.PersistentFlags().String("builder-oci-image-ppc64le", "", "Builder OCI image to spawn (ppc64le)")
	root.PersistentFlags().String("builder-oci-image-s390x", "", "Builder OCI image to spawn (s390x)")

	root.PersistentFlags().Bool("k8s-supports-cross-platform-no-affinity", false, "All Kubernetes nodes supports cross-platform so no affinity rules required (Ex. M1 Docker Desktop)")
	root.PersistentFlags().Bool("provision-only", false, "Provision only mode only provisions ephemeral resources. Only used for extarches (s390x and ppc64le)")

	temporalutils.AddFlags(root.PersistentFlags())
	peridotcommon.AddFlags(root.PersistentFlags())
	utils.AddFlags(root.PersistentFlags(), cnf)
}

func mn(_ *cobra.Command, _ []string) {
	c, err := temporalutils.NewClient(client.Options{})
	if err != nil {
		log.Fatalf("could not create temporal client: %v", err)
	}

	if viper.GetBool("provision-only") {
		peridotimplv1.MainTaskQueue = strings.ToLower(fmt.Sprintf("peridot-provision-only-%s", os.Getenv("PERIDOT_SITE")))
	}

	w, err := builderv1.NewWorker(serverconnector.MustAuto(), c, peridotimplv1.MainTaskQueue, nil)
	if err != nil {
		logrus.Fatalf("could not init worker: %v", err)
	}
	defer w.Client.Close()

	// Ephemeral
	if !viper.GetBool("provision-only") {
		w.Worker.RegisterWorkflow(w.WorkflowController.BuildWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.BuildModuleWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.BuildModuleStreamWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.ImportPackageWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.BuildBatchWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.TriggerBuildFromBatchWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.ImportPackageBatchWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.TriggerImportFromBatchWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.SyncCatalogWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.RpmImportWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.RpmLookasideBatchImportWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.CreateHashedRepositoriesWorkflow)
		w.Worker.RegisterWorkflow(w.WorkflowController.CloneSwapWorkflow)
		w.Worker.RegisterActivity(w.WorkflowController.CloneSwapActivity)
	}
	w.Worker.RegisterWorkflow(w.WorkflowController.ProvisionWorkerWorkflow)
	w.Worker.RegisterWorkflow(w.WorkflowController.DestroyWorkerWorkflow)
	w.Worker.RegisterActivity(w.WorkflowController.CreateK8sPodActivity)
	w.Worker.RegisterActivity(w.WorkflowController.DeleteK8sPodActivity)

	// Logs
	w.Worker.RegisterActivity(w.WorkflowController.IngestLogsActivity)

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
		logrus.Fatal(err)
	}
}
