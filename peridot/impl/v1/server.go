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

package peridotimplv1

import (
	"context"
	"fmt"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	hydraclient "github.com/ory/hydra-client-go/client"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"io"
	"net/url"
	commonpb "peridot.resf.org/common"
	builderv1 "peridot.resf.org/peridot/builder/v1"
	peridotdb "peridot.resf.org/peridot/db"
	"peridot.resf.org/peridot/lookaside"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/servicecatalog"
	"peridot.resf.org/utils"
)

var (
	MainTaskQueue = "peridot-main-queue"
)

type (
	PermissionType = string
	ObjectType     = string
)

const (
	PermissionManage PermissionType = "manage"
	PermissionBuild  PermissionType = "build"
	PermissionView   PermissionType = "view"

	ObjectProject ObjectType = "peridot/project"
	ObjectGlobal  ObjectType = "global"

	ObjectIdPeridot string = "peridot"

	SubjectUser ObjectType = "user"
)

type Server struct {
	peridotpb.UnimplementedProjectServiceServer
	peridotpb.UnimplementedBuildServiceServer
	peridotpb.UnimplementedImportServiceServer
	peridotpb.UnimplementedPackageServiceServer
	peridotpb.UnimplementedSearchServiceServer
	peridotpb.UnimplementedTaskServiceServer

	log            *logrus.Logger
	db             peridotdb.Access
	temporal       client.Client
	temporalWorker *builderv1.Worker
	authz          *authzed.Client
	hydra          *hydraclient.OryHydra
	hydraAdmin     *hydraclient.OryHydra
	storage        lookaside.Storage
}

func NewServer(db peridotdb.Access, c client.Client, storage lookaside.Storage) (*Server, error) {
	temporalWorker, err := builderv1.NewWorker(db, c, MainTaskQueue, nil)
	if err != nil {
		return nil, err
	}

	authz, err := authzed.NewClient(servicecatalog.SpiceDB(), servicecatalog.SpiceDBCredentials()...)
	if err != nil {
		return nil, err
	}

	publicURL, err := url.Parse(servicecatalog.HydraPublic())
	if err != nil {
		return nil, fmt.Errorf("could not parse hydra public url, error: %s", err)
	}
	hydraSDK := hydraclient.NewHTTPClientWithConfig(nil, &hydraclient.TransportConfig{
		Schemes:  []string{publicURL.Scheme},
		Host:     publicURL.Host,
		BasePath: publicURL.Path,
	})

	adminURL, err := url.Parse(servicecatalog.HydraAdmin())
	if err != nil {
		return nil, fmt.Errorf("could not parse hydra admin url, error: %s", err)
	}
	hydraAdminSDK := hydraclient.NewHTTPClientWithConfig(nil, &hydraclient.TransportConfig{
		Schemes:  []string{adminURL.Scheme},
		Host:     adminURL.Host,
		BasePath: adminURL.Path,
	})

	return &Server{
		log:            logrus.New(),
		db:             db,
		temporal:       c,
		temporalWorker: temporalWorker,
		authz:          authz,
		hydra:          hydraSDK,
		hydraAdmin:     hydraAdminSDK,
		storage:        storage,
	}, nil
}

func (s *Server) interceptor(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	n := utils.EndInterceptor
	n = utils.AuthInterceptor(s.hydra, s.hydraAdmin, []string{}, false, n)

	return n(ctx, req, usi, handler)
}

func (s *Server) serverInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	n := utils.ServerEndInterceptor
	n = utils.ServerAuthInterceptor(s.hydra, s.hydraAdmin, []string{}, false, n)

	return n(srv, ss, info, handler)
}

func (s *Server) Run() {
	res := utils.NewGRPCServer(
		&utils.GRPCOptions{
			DialOptions: []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			},
			ServerOptions: []grpc.ServerOption{
				grpc.UnaryInterceptor(s.interceptor),
				grpc.StreamInterceptor(s.serverInterceptor),
				grpc.MaxRecvMsgSize(1024 * 1024 * 1024),
			},
		},
		func(r *utils.Register) {
			endpoints := []utils.GrpcEndpointRegister{
				commonpb.RegisterHealthCheckServiceHandlerFromEndpoint,
				peridotpb.RegisterProjectServiceHandlerFromEndpoint,
				peridotpb.RegisterBuildServiceHandlerFromEndpoint,
				peridotpb.RegisterImportServiceHandlerFromEndpoint,
				peridotpb.RegisterPackageServiceHandlerFromEndpoint,
				peridotpb.RegisterSearchServiceHandlerFromEndpoint,
				peridotpb.RegisterTaskServiceHandlerFromEndpoint,
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

			peridotpb.RegisterProjectServiceServer(r.Server, s)
			peridotpb.RegisterBuildServiceServer(r.Server, s)
			peridotpb.RegisterImportServiceServer(r.Server, s)
			peridotpb.RegisterPackageServiceServer(r.Server, s)
			peridotpb.RegisterSearchServiceServer(r.Server, s)
			peridotpb.RegisterTaskServiceServer(r.Server, s)
		},
	)

	defer res.Cancel()
	res.WaitGroup.Wait()
}

func (s *Server) checkPermSubject(ctx context.Context, objectType ObjectType, objectId string, permissionType PermissionType, subject string) error {
	res, err := s.authz.CheckPermission(ctx, &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: objectType,
			ObjectId:   objectId,
		},
		Permission: permissionType,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   subject,
			},
		},
	})
	if err != nil {
		s.log.Errorf("error checking permission - %v", err)
		return utils.InternalError
	}

	if res.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		return nil
	}

	return status.Error(codes.PermissionDenied, "permission denied")
}

func (s *Server) checkPermission(ctx context.Context, objectType ObjectType, objectId string, permissionType PermissionType) error {
	userSubject := "anonymous"
	user, err := utils.UserFromContext(ctx)
	if err == nil && user != nil {
		userSubject = user.ID
	}

	err = s.checkPermSubject(ctx, objectType, objectId, permissionType, userSubject)
	if err == nil {
		return nil
	}

	// todo(mustafa): SpiceDB doesn't currently support wildcard/PUBLIC but it is in the process
	// todo(mustafa): of adding support for it. Until then, we're going to re-check the permission
	if userSubject != "anonymous" && permissionType == PermissionView {
		err = s.checkPermSubject(ctx, objectType, objectId, permissionType, "anonymous")
		if err == nil {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, "permission denied")
}

func (s *Server) lookupResourcesSubject(ctx context.Context, objectType ObjectType, permissionType PermissionType, subject string) ([]string, error) {
	res, err := s.authz.LookupResources(ctx, &v1.LookupResourcesRequest{
		ResourceObjectType: objectType,
		Permission:         permissionType,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   subject,
			},
		},
	})
	if err != nil {
		s.log.Errorf("error checking permission - %v", err)
		return nil, utils.InternalError
	}

	var ret []string
	for {
		resource, err := res.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ret = append(ret, resource.ResourceObjectId)
	}

	return ret, nil
}

func (s *Server) lookupResources(ctx context.Context, objectType ObjectType, permissionType PermissionType) ([]string, error) {
	userSubject := "anonymous"
	user, err := utils.UserFromContext(ctx)
	if err == nil && user != nil {
		userSubject = user.ID
	}

	resources, err := s.lookupResourcesSubject(ctx, objectType, permissionType, userSubject)
	if err != nil {
		return nil, err
	}

	// todo(mustafa): SpiceDB doesn't currently support wildcard/PUBLIC but it is in the process
	// todo(mustafa): of adding support for it. Until then, we're going to re-check the permission
	if userSubject != "anonymous" && permissionType == PermissionView {
		anonymousResources, err := s.lookupResourcesSubject(ctx, objectType, permissionType, "anonymous")
		if err != nil {
			return nil, err
		}

		for _, anonymousResource := range anonymousResources {
			if !utils.StrContains(anonymousResource, resources) {
				resources = append(resources, anonymousResource)
			}
		}
	}

	return resources, nil
}
