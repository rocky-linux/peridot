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

package obsidianimplv1

import (
	"context"

	"github.com/gogo/status"
	client "github.com/ory/hydra-client-go/v2"
	"peridot.resf.org/utils"

	"google.golang.org/grpc/codes"
	obsidianpb "peridot.resf.org/obsidian/pb"
)

const (
	authError  = "auth_error"
	noUser     = "no_user"
	badConsent = "bad_consent"
)

func (s *Server) ProcessLoginRequest(challenge string) (*obsidianpb.SessionStatusResponse, error) {
	ctx := context.TODO()

	loginReq, _, err := s.hydra.OAuth2API.GetOAuth2LoginRequest(ctx).LoginChallenge(challenge).Execute()
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if loginReq.Challenge != challenge {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if loginReq.Skip {
		return s.AcceptLoginRequest(ctx, challenge, loginReq)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:      true,
		ClientName: *loginReq.Client.ClientName,
		Scopes:     loginReq.RequestedScope,
	}, nil
}

func (s *Server) ProcessConsentRequest(challenge string) (*obsidianpb.SessionStatusResponse, error) {
	ctx := context.TODO()

	consentReq, _, err := s.hydra.OAuth2API.GetOAuth2ConsentRequest(ctx).ConsentChallenge(challenge).Execute()
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if consentReq.Challenge != challenge {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if *consentReq.Skip {
		return s.AcceptConsentRequest(ctx, challenge, consentReq)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:      true,
		ClientName: *consentReq.Client.ClientName,
		Scopes:     consentReq.RequestedScope,
	}, nil
}

func (s *Server) AcceptConsentRequest(ctx context.Context, challenge string, consentReq *client.OAuth2ConsentRequest) (*obsidianpb.SessionStatusResponse, error) {
	user, err := s.db.GetUserByID(*consentReq.Subject)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, noUser)
	}

	consent, _, err := s.hydra.OAuth2API.AcceptOAuth2ConsentRequest(ctx).
		ConsentChallenge(challenge).
		AcceptOAuth2ConsentRequest(client.AcceptOAuth2ConsentRequest{
			Remember:                 utils.Pointer[bool](true),
			GrantScope:               consentReq.RequestedScope,
			GrantAccessTokenAudience: consentReq.RequestedAccessTokenAudience,
			Session: &client.AcceptOAuth2ConsentRequestSession{
				AccessToken: map[string]interface{}{
					"id": user.ID,
				},
				IdToken: map[string]interface{}{
					"id":         user.ID,
					"name":       user.Name.String,
					"email":      user.Email,
					"created_at": user.CreatedAt,
				},
			},
		}).Execute()

	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, badConsent)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:       true,
		RedirectUrl: consent.RedirectTo,
		ClientName:  *consentReq.Client.ClientName,
		Scopes:      consentReq.RequestedScope,
	}, nil
}

func (s *Server) AcceptLoginRequest(ctx context.Context, challenge string, loginReq *client.OAuth2LoginRequest) (*obsidianpb.SessionStatusResponse, error) {
	acceptLogin, _, err := s.hydra.OAuth2API.AcceptOAuth2LoginRequest(ctx).
		LoginChallenge(challenge).
		AcceptOAuth2LoginRequest(client.AcceptOAuth2LoginRequest{
			Context:  ctx,
			Remember: utils.Pointer[bool](true),
			Subject:  loginReq.Subject,
		}).Execute()
	if err != nil {
		return nil, status.Error(codes.Internal, authError)
	}

	user, err := s.db.GetUserByID(loginReq.Subject)
	if err != nil || user == nil || user.ID == "" {
		return nil, status.Error(codes.InvalidArgument, noUser)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:       true,
		RedirectUrl: acceptLogin.RedirectTo,
		ClientName:  *loginReq.Client.ClientName,
		Scopes:      loginReq.RequestedScope,
	}, nil
}
