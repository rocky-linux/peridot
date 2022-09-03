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
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"log"
	"net/http"

	"openapi.peridot.resf.org/peridotopenapi"
)

type service string

const (
	serviceProject = service("project")
	serviceImport  = service("import")
	servicePackage = service("package")
	serviceBuild   = service("build")
	serviceTask    = service("task")
)

var (
	doNotUseDirectlyClient map[service]interface{}
	doNotUseDirectlyCtx    context.Context
)

func getClient(svc service) interface{} {
	if doNotUseDirectlyClient != nil {
		return doNotUseDirectlyClient[svc]
	}
	doNotUseDirectlyClient = make(map[service]interface{})

	tlsConfig := &tls.Config{
		// We should allow users to configure this.
		InsecureSkipVerify: skipCaVerify(), //nolint:gosec
	}

	apiCfg := &peridotopenapi.Configuration{
		Debug:     debug(),
		Host:      endpoint(),
		Scheme:    "https",
		UserAgent: "peridot/0.1",
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
		DefaultHeader: map[string]string{},
		Servers: peridotopenapi.ServerConfigurations{
			{
				URL: "https://" + endpoint(),
			},
		},
		OperationServers: map[string]peridotopenapi.ServerConfigurations{},
	}

	apiClient := peridotopenapi.NewAPIClient(apiCfg)

	doNotUseDirectlyClient[serviceProject] = apiClient.ProjectServiceApi
	doNotUseDirectlyClient[serviceImport] = apiClient.ImportServiceApi
	doNotUseDirectlyClient[servicePackage] = apiClient.PackageServiceApi
	doNotUseDirectlyClient[serviceBuild] = apiClient.BuildServiceApi
	doNotUseDirectlyClient[serviceTask] = apiClient.TaskServiceApi

	return doNotUseDirectlyClient[svc]
}

func getContext() context.Context {
	if doNotUseDirectlyCtx == nil {
		doNotUseDirectlyCtx = context.TODO()

		oauth2Config := &clientcredentials.Config{
			ClientID:     getClientId(),
			ClientSecret: getClientSecret(),
			// We don't currently support scopes, but authorize based on SpiceDB.
			// Lack of scopes does not indicate that client has full access, but
			// that we're managing access server sides and scopes doesn't affect that.
			Scopes:    []string{},
			TokenURL:  fmt.Sprintf("https://%s/oauth2/token", hdrEndpoint()),
			AuthStyle: oauth2.AuthStyleInHeader,
		}

		tokenSource := oauth2Config.TokenSource(doNotUseDirectlyCtx)
		doNotUseDirectlyCtx = context.WithValue(doNotUseDirectlyCtx, peridotopenapi.ContextOAuth2, tokenSource)
	}
	return doNotUseDirectlyCtx
}

func errFatal(err error) {
	if err != nil {
		log.Fatalf("an error occurred: %s", err.Error())
	}
}
