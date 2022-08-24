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

package impl

import (
	"context"
	hydraclient "github.com/ory/hydra-client-go/client"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net/url"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"peridot.resf.org/servicecatalog"
	"peridot.resf.org/utils"
)

type Server struct {
	secparseadminpb.UnimplementedSecparseAdminServer

	log   *logrus.Logger
	db    db.Access
	hydra *hydraclient.OryHydra
}

func NewServer(db db.Access) *Server {
	publicURL, err := url.Parse(servicecatalog.HydraPublic())
	if err != nil {
		logrus.Fatalf("failed to parse hydra public url: %s", err)
	}

	hydraSDK := hydraclient.NewHTTPClientWithConfig(nil, &hydraclient.TransportConfig{
		Schemes:  []string{publicURL.Scheme},
		Host:     publicURL.Host,
		BasePath: publicURL.Path,
	})

	return &Server{
		log:   logrus.New(),
		db:    db,
		hydra: hydraSDK,
	}
}

func (s *Server) interceptor(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	n := utils.EndInterceptor
	n = utils.AuthInterceptor(s.hydra, nil, []string{}, true, n)

	return n(ctx, req, usi, handler)
}

func (s *Server) Run() {
	res := utils.NewGRPCServer(
		&utils.GRPCOptions{
			ServerOptions: []grpc.ServerOption{
				grpc.UnaryInterceptor(s.interceptor),
			},
		},
		func(r *utils.Register) {
			err := secparseadminpb.RegisterSecparseAdminHandlerFromEndpoint(
				r.Context,
				r.Mux,
				r.Endpoint,
				r.Options,
			)
			if err != nil {
				s.log.Fatalf("could not register handler - %s", err)
			}
		},
		func(r *utils.RegisterServer) {
			secparseadminpb.RegisterSecparseAdminServer(r.Server, s)
		},
	)

	defer res.Cancel()
	res.WaitGroup.Wait()
}
