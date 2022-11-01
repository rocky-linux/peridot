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
	"database/sql"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ory/hydra-client-go/client/admin"
	hydramodels "github.com/ory/hydra-client-go/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"peridot.resf.org/obsidian/db/models"
	obsidianpb "peridot.resf.org/obsidian/pb"
	"peridot.resf.org/utils"
)

type EmailClaim struct {
	Email string `json:"email"`
}
type NameClaim struct {
	Name string `json:"name"`
}

// callbackForwarder helps create an external callback since usual
// OAuth2 providers doesn't allow callback to localhost.
// Cloudflare Workers is a good option for this.
func callbackForwarder(callbackURL string) string {
	env := os.Getenv("RESF_ENV")
	// this section contained a callback forwarder, but cannot be published
	// todo(mustafa): evaluate other ways to make it easier for dev
	if env == "dev" || env == "" {
		if fwd := os.Getenv("OBSIDIAN_CALLBACK_FORWARDER"); fwd != "" {
			return fmt.Sprintf("%s/%s", fwd, callbackURL)
		}
		return callbackURL
	}
	return callbackURL
}

func (s *Server) GetOAuth2Providers(_ context.Context, _ *obsidianpb.GetOAuth2ProvidersRequest) (*obsidianpb.GetOAuth2ProvidersResponse, error) {
	providers, err := s.db.ListOAuth2Providers()
	if err != nil {
		s.log.Errorf("failed to list OAuth2 providers: %s", err)
		return nil, utils.CouldNotRetrieveObjects
	}

	return &obsidianpb.GetOAuth2ProvidersResponse{
		Providers: providers.ToProto(),
	}, nil
}

func (s *Server) InitiateOAuth2Session(ctx context.Context, req *obsidianpb.InitiateOAuth2SessionRequest) (*obsidianpb.InitiateOAuth2SessionResponse, error) {
	if req.Challenge == "" {
		return nil, status.Error(codes.InvalidArgument, "challenge cannot be empty")
	}
	if req.ProviderId == "" {
		return nil, status.Error(codes.InvalidArgument, "provider_id cannot be empty")
	}

	loginReq, _, conf, err := s.getProviderAndLoginRequest(ctx, req.Challenge, req.ProviderId)
	if err != nil {
		return nil, err
	}

	redirectURL := conf.AuthCodeURL(*loginReq.Payload.Challenge)
	err = grpc.SetHeader(ctx, metadata.Pairs("location", redirectURL))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set redirect url")
	}

	return &obsidianpb.InitiateOAuth2SessionResponse{}, nil
}

func (s *Server) ConfirmOAuth2Session(ctx context.Context, req *obsidianpb.ConfirmOAuth2SessionRequest) (*obsidianpb.ConfirmOAuth2SessionResponse, error) {
	if req.State == "" {
		return nil, status.Error(codes.InvalidArgument, "state cannot be empty")
	}
	if req.ProviderId == "" {
		return nil, status.Error(codes.InvalidArgument, "provider_id cannot be empty")
	}

	loginReq, provider, conf, err := s.getProviderAndLoginRequest(ctx, req.State, req.ProviderId)
	if err != nil {
		return nil, err
	}

	tok, err := conf.Exchange(ctx, req.Code)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to exchange code: %s", err)
	}

	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "id_token not found")
	}

	var verifier *oidc.IDTokenVerifier
	switch provider.Provider {
	case "google":
		p, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create provider")
		}
		verifier = p.Verifier(&oidc.Config{ClientID: provider.ClientId})
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported provider")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to verify id_token: %s", err)
	}

	beginTx, err := s.db.Begin()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to begin transaction")
	}
	tx := s.db.UseTransaction(beginTx)

	committed := false
	defer func() {
		if !committed {
			_ = beginTx.Rollback()
		}
	}()

	// Check if the user is already associated with provider
	existingUser, err := s.db.GetUserByOAuth2ProviderExternalID(provider.ID.String(), idToken.Subject)
	if err != nil {
		if err != sql.ErrNoRows {
			s.log.Errorf("failed to get user by oauth2 provider external id: %s", err)
			return nil, utils.InternalError
		} else {
			var name *string
			var email string

			nameClaim := NameClaim{}
			emailClaim := EmailClaim{}

			// Get email and potentially name from id_token
			if err := idToken.Claims(&nameClaim); err == nil {
				name = &nameClaim.Name
			}
			if err := idToken.Claims(&emailClaim); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "failed to parse email claim: %s", err)
			}
			email = emailClaim.Email

			// Check if the user already exists (and is connected to another provider)
			_, err = s.db.GetUserByEmail(email)
			if err == nil {
				// The user has to link with this provider first
				alreadyExistsErr := status.Errorf(codes.AlreadyExists, "user with email %s already exists, you need to sign in with an already established provider to link a new one", email)
				rejectRes, err := s.hydra.Admin.RejectLoginRequest(&admin.RejectLoginRequestParams{
					Body: &hydramodels.RejectRequest{
						StatusCode:       int64(codes.AlreadyExists),
						ErrorDescription: "User already exists",
						ErrorHint:        "Sign in to your account, link this provider and try again",
						Error:            "user_already_exists",
					},
					LoginChallenge: req.State,
					Context:        nil,
					HTTPClient:     nil,
				})
				if err != nil {
					return nil, alreadyExistsErr
				}

				// Redirect to Hydra location
				err = grpc.SetHeader(ctx, metadata.Pairs("location", *rejectRes.Payload.RedirectTo))
				if err != nil {
					return nil, alreadyExistsErr
				}

				return &obsidianpb.ConfirmOAuth2SessionResponse{}, nil
			}
			if err != sql.ErrNoRows {
				s.log.Errorf("failed to get user by email: %s", err)
				return nil, status.Error(codes.Internal, "failed to check if user exists")
			}

			// User doesn't exist so create it
			newUser, err := tx.CreateUser(name, email)
			if err != nil {
				s.log.Errorf("failed to create user: %s", err)
				return nil, status.Error(codes.Internal, "failed to create user")
			}
			// Link the user to the provider
			err = tx.LinkUserToOAuth2Provider(newUser.ID, provider.ID.String(), idToken.Subject)
			if err != nil {
				s.log.Errorf("failed to link user to oauth2 provider: %s", err)
				return nil, status.Error(codes.Internal, "failed to link user to oauth2 provider")
			}
			existingUser = newUser
		}
	}
	err = beginTx.Commit()
	if err != nil {
		s.log.Errorf("failed to commit transaction: %s", err)
		return nil, status.Error(codes.Internal, "could not save user")
	}
	committed = true

	// Set user ID and accept the login request
	loginReq.Payload.Subject = &existingUser.ID
	res, err := s.AcceptLoginRequest(ctx, req.State, loginReq)
	if err != nil {
		return nil, err
	}

	// Redirect to Hydra location
	err = grpc.SetHeader(ctx, metadata.Pairs("location", res.RedirectUrl))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set header")
	}

	return &obsidianpb.ConfirmOAuth2SessionResponse{}, nil
}

func (s *Server) getProviderAndLoginRequest(ctx context.Context, challenge string, providerId string) (*admin.GetLoginRequestOK, *models.OAuth2Provider, *oauth2.Config, error) {
	loginReq, err := s.hydra.Admin.GetLoginRequest(&admin.GetLoginRequestParams{
		LoginChallenge: challenge,
		Context:        ctx,
	})
	if err != nil || loginReq == nil {
		return nil, nil, nil, status.Error(codes.NotFound, "login request not found")
	}

	provider, err := s.db.GetOAuth2ProviderByID(providerId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil, status.Error(codes.NotFound, "provider not found")
		}
		s.log.Errorf("failed to get OAuth2 provider: %s", err)
		return nil, nil, nil, utils.InternalError
	}

	conf := oauth2.Config{}
	switch provider.Provider {
	case "google":
		conf = oauth2.Config{
			ClientID:     provider.ClientId,
			ClientSecret: provider.ClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  callbackForwarder(fmt.Sprintf("%s/v1/oauth2/providers/%s/callback", os.Getenv("OBSIDIAN_HTTP_PUBLIC_URL"), provider.ID.String())),
			Scopes:       []string{"openid", "email", "profile"},
		}
	default:
		return nil, nil, nil, status.Error(codes.InvalidArgument, "provider not supported")
	}

	return loginReq, provider, &conf, nil
}
