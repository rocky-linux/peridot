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
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"google.golang.org/grpc/codes"
	obsidianpb "peridot.resf.org/obsidian/pb"
)

const (
	authError = "auth_error"
	noUser    = "no_user"
)

func (s *Server) ProcessLoginRequest(challenge string) (*obsidianpb.SessionStatusResponse, error) {
	ctx := context.TODO()

	loginReq, err := s.hydra.Admin.GetLoginRequest(&admin.GetLoginRequestParams{
		LoginChallenge: challenge,
		Context:        ctx,
	})
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if *loginReq.Payload.Challenge != challenge {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if *loginReq.Payload.Skip {
		return s.AcceptLoginRequest(ctx, challenge, loginReq)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:      true,
		ClientName: loginReq.Payload.Client.ClientName,
		Scopes:     loginReq.Payload.RequestedScope,
	}, nil
}

func (s *Server) ProcessConsentRequest(challenge string) (*obsidianpb.SessionStatusResponse, error) {
	ctx := context.TODO()

	consentReq, err := s.hydra.Admin.GetConsentRequest(&admin.GetConsentRequestParams{
		Context:          ctx,
		ConsentChallenge: challenge,
	})
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if *consentReq.Payload.Challenge != challenge {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	if consentReq.Payload.Skip {
		return s.AcceptConsentRequest(ctx, challenge, consentReq)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:      true,
		ClientName: consentReq.Payload.Client.ClientName,
		Scopes:     consentReq.Payload.RequestedScope,
	}, nil
}

func (s *Server) AcceptConsentRequest(ctx context.Context, challenge string, consentReq *admin.GetConsentRequestOK) (*obsidianpb.SessionStatusResponse, error) {
	user, err := s.db.GetUserByID(consentReq.Payload.Subject)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, noUser)
	}

	consent, err := s.hydra.Admin.AcceptConsentRequest(&admin.AcceptConsentRequestParams{
		Context:          ctx,
		ConsentChallenge: challenge,
		Body: &models.AcceptConsentRequest{
			Remember:                 true,
			GrantScope:               consentReq.Payload.RequestedScope,
			GrantAccessTokenAudience: consentReq.Payload.RequestedAccessTokenAudience,
			Session: &models.ConsentRequestSession{
				AccessToken: map[string]interface{}{
					"id": user.ID,
				},
				IDToken: map[string]interface{}{
					"id":         user.ID,
					"name":       user.Name.String,
					"email":      user.Email,
					"created_at": user.CreatedAt,
				},
			},
		},
	})
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, authError)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:       true,
		RedirectUrl: *consent.Payload.RedirectTo,
		ClientName:  consentReq.Payload.Client.ClientName,
		Scopes:      consentReq.Payload.RequestedScope,
	}, nil
}

func (s *Server) AcceptLoginRequest(ctx context.Context, challenge string, loginReq *admin.GetLoginRequestOK) (*obsidianpb.SessionStatusResponse, error) {
	acceptLogin, err := s.hydra.Admin.AcceptLoginRequest(&admin.AcceptLoginRequestParams{
		LoginChallenge: challenge,
		Body: &models.AcceptLoginRequest{
			Subject:  loginReq.Payload.Subject,
			Remember: true,
		},
		Context: ctx,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, authError)
	}

	user, err := s.db.GetUserByID(*loginReq.Payload.Subject)
	if err != nil || user == nil || user.ID == "" {
		return nil, status.Error(codes.InvalidArgument, noUser)
	}

	return &obsidianpb.SessionStatusResponse{
		Valid:       true,
		RedirectUrl: *acceptLogin.Payload.RedirectTo,
		ClientName:  loginReq.Payload.Client.ClientName,
		Scopes:      loginReq.Payload.RequestedScope,
	}, nil
}
