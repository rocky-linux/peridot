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
	"peridot.resf.org/utils"
	"strings"

	"github.com/ory/hydra-client-go/v2"
	"github.com/sirupsen/logrus"
	"peridot.resf.org/servicecatalog"
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

	hydraSDKConfiguration := client.NewConfiguration()
	hydraSDKConfiguration.Servers[0].URL = adminURL.String()
	hydraSDKConfiguration.Host = adminURL.Host
	hydraSDKConfiguration.Scheme = adminURL.Scheme
	hydraSDK := client.NewAPIClient(hydraSDKConfiguration)

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
	clientModel := client.OAuth2Client{
		ClientName:   &visibleName,
		ClientId:     &serviceName,
		Scope:        utils.Pointer[string](strings.Join(req.Scopes, " ")),
		ClientSecret: utils.Pointer[string](secret()),
	}
	if req.Frontend {
		clientModel.RedirectUris = redirectUri(req)
	}

	ret := &AutoSignupResponse{
		ClientID: serviceName,
		Secret:   secret(),
	}

	_, _, err = hydraSDK.OAuth2API.GetOAuth2Client(ctx, serviceName).Execute()
	if err != nil {
		logrus.Error(err)
		_, _, err := hydraSDK.OAuth2API.CreateOAuth2Client(ctx).OAuth2Client(clientModel).Execute()
		if err != nil {
			logrus.Fatalf("could not create hydra client: %v", err)
		}
		logrus.Infof("created hydra client %s", serviceName)
	} else {
		_, _, err := hydraSDK.OAuth2API.SetOAuth2Client(ctx, serviceName).OAuth2Client(clientModel).Execute()
		if err != nil {
			logrus.Fatalf("could not update hydra client: %v", err)
		}
		logrus.Infof("updated hydra client %s", serviceName)
	}

	return ret
}
