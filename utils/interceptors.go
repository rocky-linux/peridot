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

package utils

import (
	"context"
	"fmt"
	"github.com/ory/hydra-client-go/v2"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type InterceptorFunc func(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
type ServerInterceptorFunc func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error

type ContextUser struct {
	ID        string `json:"id"`
	AuthToken string `json:"authToken"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

// Finish chains all interceptors
func (a InterceptorFunc) Finish(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return a(ctx, req, usi, handler)
}
func (a ServerInterceptorFunc) Finish(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return a(srv, ss, info, handler)
}

// EndInterceptor should be used in the end of an interceptor chain
func EndInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
func ServerEndInterceptor(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, ss)
}

func checkAuth(ctx context.Context, hydraSDK *client.APIClient, hydraAdmin *client.APIClient) (context.Context, error) {
	// fetch metadata from grpc
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Error(codes.InvalidArgument, "invalid request sent")
	}

	// get authorization header
	authHeader := meta["authorization"]
	if len(authHeader) == 0 {
		return ctx, status.Error(codes.InvalidArgument, "empty authorization header")
	}

	// verify that the authorization header contains a Bearer token
	authToken := strings.SplitN(authHeader[0], " ", 2)
	if len(authToken) != 2 || authToken[0] != "Bearer" {
		return ctx, status.Error(codes.InvalidArgument, "invalid authorization token")
	}

	userInfo, _, err := hydraSDK.OidcAPI.GetOidcUserInfo(context.WithValue(ctx, client.ContextAccessToken, authToken[1])).Execute()
	if err != nil {
		return ctx, err
	}
	if userInfo.GetSub() == "" && hydraAdmin != nil {
		introspect, _, err := hydraAdmin.OAuth2API.IntrospectOAuth2Token(ctx).Token(authToken[1]).Execute()
		if err != nil {
			logrus.Errorf("error introspecting token: %s", err)
			return ctx, status.Errorf(codes.Unauthenticated, "invalid authorization token")
		}

		userInfo.Sub = introspect.ClientId
		userInfo.Name = introspect.Sub
		newEmail := fmt.Sprintf("%s@%s", *introspect.Sub, "serviceaccount.resf.org")
		userInfo.Email = &newEmail
	}
	if userInfo.GetSub() == "" {
		return ctx, status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	// supply subject and token to further requests
	pairs := metadata.Pairs("x-user-id", *userInfo.Sub, "x-user-name", *userInfo.Name, "x-user-email", *userInfo.Email, "x-auth-token", authToken[1])
	ctx = metadata.NewIncomingContext(ctx, metadata.Join(meta, pairs))

	return ctx, nil
}

// AuthInterceptor requires OAuth2 authentication for all routes except listed
func AuthInterceptor(hydraSDK *client.APIClient, hydraAdminSDK *client.APIClient, excludedMethods []string, enforce bool, next InterceptorFunc) InterceptorFunc {
	return func(ctx context.Context, req interface{}, usi *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip authentication for excluded methods
		if !StrContains(usi.FullMethod, excludedMethods) {
			var err error
			if ctx, err = checkAuth(ctx, hydraSDK, hydraAdminSDK); err != nil {
				if enforce {
					return nil, err
				}
			}
		}

		return next(ctx, req, usi, handler)
	}
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}
func ServerAuthInterceptor(hydraSDK *client.APIClient, hydraAdminSDK *client.APIClient, excludedMethods []string, enforce bool, next ServerInterceptorFunc) ServerInterceptorFunc {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newStream := serverStream{
			ServerStream: ss,
			ctx:          ss.Context(),
		}
		// skip authentication for excluded methods
		if !StrContains(info.FullMethod, excludedMethods) {
			var ctx context.Context
			var err error
			if ctx, err = checkAuth(ss.Context(), hydraSDK, hydraAdminSDK); err != nil {
				if enforce {
					return err
				}
			}
			if ctx != nil {
				newStream.ctx = ctx
			}
		}

		return next(srv, &newStream, info, handler)
	}
}

func UserFromContext(ctx context.Context) (*ContextUser, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid request sent")
	}

	uid := meta["x-user-id"]
	if len(uid) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no user id")
	}

	authTokens := meta["x-auth-token"]
	if len(authTokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no auth token")
	}

	var name string
	if names := meta["x-user-name"]; len(names) > 0 {
		name = names[0]
	}
	var email string
	if emails := meta["x-user-email"]; len(emails) > 0 {
		email = emails[0]
	}

	return &ContextUser{
		ID:        uid[0],
		AuthToken: authTokens[0],
		Name:      name,
		Email:     email,
	}, nil
}
