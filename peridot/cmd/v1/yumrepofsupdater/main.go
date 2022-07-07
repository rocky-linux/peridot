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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	commonpb "peridot.resf.org/common"
	builderv1 "peridot.resf.org/peridot/builder/v1"
	peridotcommon "peridot.resf.org/peridot/common"
	serverconnector "peridot.resf.org/peridot/db/connector"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	"peridot.resf.org/servicecatalog"
	"peridot.resf.org/temporalutils"
	"peridot.resf.org/utils"
	"sync"
)

var root = &cobra.Command{
	Use: "yumrepofs",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

const MainTaskQueue = "yumrepofs"

func init() {
	cnf.DefaultPort = 45102

	dname := "peridot"
	cnf.DatabaseName = &dname
	cnf.Name = "yumrepofsupdater"

	peridotcommon.AddFlags(root.PersistentFlags())
	utils.AddFlags(root.PersistentFlags(), cnf)
}

func mn(_ *cobra.Command, _ []string) {
	c, err := temporalutils.NewClient(client.Options{})
	if err != nil {
		logrus.Fatalf("could not create temporal client: %v", err)
	}

	sess, err := utils.NewAwsSession(&aws.Config{})
	if err != nil {
		logrus.Fatalf("could not create aws session: %v", err)
	}
	ddb := dynamodb.New(sess)

	keykeeperConn, err := grpc.Dial(servicecatalog.KeykeeperGrpc(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("could not connect to keykeeper: %v", err)
	}
	keykeeperClient := keykeeperpb.NewKeykeeperServiceClient(keykeeperConn)

	w, err := builderv1.NewWorker(serverconnector.MustAuto(), c, MainTaskQueue, &builderv1.ExtraReq{
		KeykeeperClient: keykeeperClient,
		DynamoDB:        ddb,
	})
	if err != nil {
		logrus.Fatalf("could not init worker: %v", err)
	}
	defer w.Client.Close()

	// Yumrepofs
	w.Worker.RegisterWorkflow(w.WorkflowController.RepoUpdaterWorkflow)
	w.Worker.RegisterActivity(w.WorkflowController.UpdateRepoActivity)
	w.Worker.RegisterActivity(w.WorkflowController.RequestKeykeeperSignActivity)

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
