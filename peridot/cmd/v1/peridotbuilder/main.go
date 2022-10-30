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
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	builderv1 "peridot.resf.org/peridot/builder/v1"
	peridotcommon "peridot.resf.org/peridot/common"
	serverconnector "peridot.resf.org/peridot/db/connector"
	"peridot.resf.org/peridot/db/models"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	"peridot.resf.org/peridot/plugin"
	"peridot.resf.org/servicecatalog"
	"peridot.resf.org/temporalutils"
	"peridot.resf.org/utils"
)

var root = &cobra.Command{
	Use: "peridotbuilder",
	Run: mn,
}

var (
	cnf = utils.NewFlagConfig()
)

func init() {
	cnf.DefaultPort = 6665

	dname := "peridot"
	cnf.DatabaseName = &dname
	cnf.Name = "peridotbuilder"

	root.PersistentFlags().String("task-queue", "", "Task queue for builder")
	root.PersistentFlags().String("project-id", "", "Project ID invoking this builder")
	root.PersistentFlags().String("task-id", "", "Task ID invoking this builder")
	root.PersistentFlags().String("parent-task-id", "", "Parent of the task invoking this builder")
	_ = root.MarkFlagRequired("task-queue")
	_ = root.MarkFlagRequired("project-id")
	_ = root.MarkFlagRequired("task-id")

	temporalutils.AddFlags(root.PersistentFlags())
	peridotcommon.AddFlags(root.PersistentFlags())
	utils.AddFlags(root.PersistentFlags(), cnf)
}

func initiatePlugins(plugins models.Plugins) ([]plugin.Plugin, error) {
	var initiatedPlugins []plugin.Plugin
	for _, p := range plugins {
		anyCfg := &anypb.Any{}
		err := protojson.Unmarshal(p.Configuration, anyCfg)
		if err != nil {
			return nil, err
		}

		// No plugins as of now
	}

	return initiatedPlugins, nil
}

func mn(_ *cobra.Command, _ []string) {
	taskQueue := viper.GetString("task-queue")
	projectId := viper.GetString("project-id")

	c, err := temporalutils.NewClient(client.Options{})
	if err != nil {
		log.Fatalf("could not create temporal client: %v", err)
	}

	db := serverconnector.MustAuto()

	var initiatedPlugins []plugin.Plugin
	if projectId != "" {
		plugins, err := db.GetPluginsForProject(projectId)
		if err != nil {
			logrus.Fatalf("could not get plugins: %v", err)
		}
		if plugins != nil {
			initiatedPlugins, err = initiatePlugins(plugins)
			if err != nil {
				logrus.Fatalf("could not initiate plugins: %v", err)
			}
		}
	}

	keykeeperConn, err := grpc.Dial(servicecatalog.KeykeeperGrpc(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("could not connect to keykeeper: %v", err)
	}
	keykeeperClient := keykeeperpb.NewKeykeeperServiceClient(keykeeperConn)
	extraReq := &builderv1.ExtraReq{
		KeykeeperClient: keykeeperClient,
	}

	w, err := builderv1.NewWorker(db, c, taskQueue, extraReq, initiatedPlugins...)
	if err != nil {
		logrus.Fatalf("could not init worker: %v", err)
	}
	defer w.Client.Close()

	// Build
	w.Worker.RegisterWorkflow(w.WorkflowController.BuildWorkflow)
	w.Worker.RegisterActivity(w.WorkflowController.BuildSRPMActivity)
	w.Worker.RegisterActivity(w.WorkflowController.UploadSRPMActivity)
	w.Worker.RegisterActivity(w.WorkflowController.BuildArchActivity)
	w.Worker.RegisterActivity(w.WorkflowController.UploadArchActivity)

	// Import
	w.Worker.RegisterWorkflow(w.WorkflowController.ImportPackageWorkflow)
	w.Worker.RegisterActivity(w.WorkflowController.UpstreamDistGitActivity)
	w.Worker.RegisterActivity(w.WorkflowController.PackageSrcGitActivity)
	w.Worker.RegisterActivity(w.WorkflowController.UpdateDistGitForSrcGitActivity)

	// Catalogsync
	w.Worker.RegisterActivity(w.WorkflowController.SyncCatalogActivity)

	// Worker
	w.Worker.RegisterWorkflow(w.WorkflowController.DestroyWorkerWorkflow)
	w.Worker.RegisterActivity(w.WorkflowController.DeleteK8sPodActivity)

	// RPM Import
	w.Worker.RegisterActivity(w.WorkflowController.RpmImportActivity)

	// Yumrepofs
	w.Worker.RegisterActivity(w.WorkflowController.CreateHashedRepositoriesActivity)

	w.Run()
}

func main() {
	utils.Main()
	if err := root.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
