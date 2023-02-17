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
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	commonpb "peridot.resf.org/common"
	peridotadminpb "peridot.resf.org/peridot/admin/pb"
	builderv1 "peridot.resf.org/peridot/builder/v1"
	peridotdb "peridot.resf.org/peridot/db"
	peridotimplv1 "peridot.resf.org/peridot/impl/v1"
	"peridot.resf.org/utils"
)

type Server struct {
	peridotadminpb.UnimplementedPeridotAdminServiceServer

	log            *logrus.Logger
	db             peridotdb.Access
	temporal       client.Client
	temporalWorker *builderv1.Worker
}

var adminUser = &utils.ContextUser{
	ID:        "peridot-errata",
	AuthToken: "",
	Name:      "Peridot Errata",
	Email:     "releng+errata@rockylinux.org",
}

func NewServer(db peridotdb.Access, c client.Client) (*Server, error) {
	temporalWorker, err := builderv1.NewWorker(db, c, peridotimplv1.MainTaskQueue, nil)
	if err != nil {
		return nil, err
	}

	return &Server{
		log:            logrus.New(),
		db:             db,
		temporal:       c,
		temporalWorker: temporalWorker,
	}, nil
}

func (s *Server) interceptor(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	n := utils.EndInterceptor

	return n(ctx, req, usi, handler)
}

func (s *Server) serverInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	n := utils.ServerEndInterceptor

	return n(srv, ss, info, handler)
}

func (s *Server) Run() {
	res := utils.NewGRPCServer(
		&utils.GRPCOptions{
			DialOptions: []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			},
			Interceptor:       s.interceptor,
			ServerInterceptor: s.serverInterceptor,
		},
		func(r *utils.Register) {
			endpoints := []utils.GrpcEndpointRegister{
				commonpb.RegisterHealthCheckServiceHandlerFromEndpoint,
				peridotadminpb.RegisterPeridotAdminServiceHandlerFromEndpoint,
			}

			for _, endpoint := range endpoints {
				err := endpoint(r.Context, r.Mux, r.Endpoint, r.Options)
				if err != nil {
					s.log.Fatalf("could not register handler - %v", err)
				}
			}
		},
		func(r *utils.RegisterServer) {
			commonpb.RegisterHealthCheckServiceServer(r.Server, &utils.HealthServer{})

			peridotadminpb.RegisterPeridotAdminServiceServer(r.Server, s)
		},
	)

	defer res.Cancel()
	res.WaitGroup.Wait()
}
