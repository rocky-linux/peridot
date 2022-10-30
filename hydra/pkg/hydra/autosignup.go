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

package hydra

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"peridot.resf.org/servicecatalog"

	"github.com/ory/hydra-client-go/client"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/ory/hydra-client-go/models"
	"github.com/sirupsen/logrus"
)

type AutoSignupRequest struct {
	Name     string
	Client   string
	Scopes   []string
	Frontend bool
	URI      string
}

type AutoSignupResponse struct {
	ClientID string
	Secret   string
}

func redirectUri(req *AutoSignupRequest) []string {
	uri := req.URI

	return []string{uri}
}

func secret() string {
	env := os.Getenv("RESF_ENV")
	if env == "" {
		return "dev-123-secret"
	}

	if scr := os.Getenv("HYDRA_SECRET"); scr != "" {
		return scr
	} else {
		logrus.Fatal("could not find hydra secret")
		// included to make compiler happy
		return ""
	}
}

func AutoSignup(req *AutoSignupRequest) *AutoSignupResponse {
	adminURL, err := url.Parse(servicecatalog.HydraAdmin())
	if err != nil {
		logrus.Fatalf("invalid hydra url: %v", err)
	}

	hydraSDK := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Schemes:  []string{adminURL.Scheme},
		Host:     adminURL.Host,
		BasePath: adminURL.Path,
	})

	ctx := context.TODO()

	ns := os.Getenv("RESF_NS")
	if ns == "" {
		ns = "dev"
	}
	name := fmt.Sprintf("%s-%s", req.Client, ns)
	visibleName := name
	if req.Name != "" {
		visibleName = req.Name
	}
	serviceName := fmt.Sprintf("autos-%s", name)
	clientModel := &models.OAuth2Client{
		ClientName:   visibleName,
		ClientID:     serviceName,
		Scope:        strings.Join(req.Scopes, " "),
		ClientSecret: secret(),
	}
	if req.Frontend {
		clientModel.RedirectUris = redirectUri(req)
	}

	ret := &AutoSignupResponse{
		ClientID: serviceName,
		Secret:   secret(),
	}

	_, err = hydraSDK.Admin.GetOAuth2Client(&admin.GetOAuth2ClientParams{
		Context: ctx,
		ID:      serviceName,
	})
	if err != nil {
		logrus.Error(err)
		_, err := hydraSDK.Admin.CreateOAuth2Client(&admin.CreateOAuth2ClientParams{
			Context: ctx,
			Body:    clientModel,
		})
		if err != nil {
			logrus.Fatalf("could not create hydra client: %v", err)
		}
		logrus.Infof("created hydra client %s", serviceName)
	} else {
		_, err := hydraSDK.Admin.UpdateOAuth2Client(&admin.UpdateOAuth2ClientParams{
			Context: ctx,
			Body:    clientModel,
			ID:      serviceName,
		})
		if err != nil {
			logrus.Fatalf("could not update hydra client: %v", err)
		}
		logrus.Infof("updated hydra client %s", serviceName)
	}

	return ret
}
