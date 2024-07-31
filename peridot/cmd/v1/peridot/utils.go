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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

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

// getLatestBuildTaskIdForPackageName retrieves the latest build task ID for a given package name within a specified project.
//
// Parameters:
// - projectId: The ID of the project where the package resides.
// - name: The name of the package for which the latest build task ID is to be fetched.
//
// Returns:
// - string: The task ID of the latest build for the specified package name.
// - error: An error if the retrieval process fails.
//
// It accomplishes the following steps:
// 1. Initializes clients for the package and build services.
// 2. Verifies the existence of the package in the specified project.
// 3. Lists the latest builds for the package name with an SUCCEEDED status.
// 4. Converts the total number of results to an integer.
// 5. Checks if the result set is larger than one page, handles pagination if needed.
// 6. Returns the task ID of the latest build if available, or an error if not found.
func getLatestBuildTaskIdForPackageName(projectId string, name string, status string) (string, error) {

	if status == "" {
		status = string(peridotopenapi.SUCCEEDED)
	}

	packageCl := getClient(servicePackage).(peridotopenapi.PackageServiceApi)
	buildCl := getClient(serviceBuild).(peridotopenapi.BuildServiceApi)

	_, _, err := packageCl.GetPackage(getContext(), projectId, "name", name).Execute()
	if err != nil {
		errFatal(err)
	}

	// try to get the latest builds for the package
	listBuildsReq := buildCl.ListBuilds(getContext(), projectId)

	listBuildsReq = listBuildsReq.FiltersPackageName(name)
	listBuildsReq = listBuildsReq.FiltersStatus(status)

	res, _, err := listBuildsReq.Execute()
	if err != nil {
		errFatal(err)
	}

	// TODO(neil): why is Total a string?
	total, err := strconv.Atoi(*res.Total)
	if err != nil {
		errFatal(err)
	}

	// TODO(neil): support pagination?
	if total > int(*res.Size) {
		fmt.Errorf("result set larger than one page")
	}

	if len(*res.Builds) > 0 {
		builds := *res.Builds
		return builds[0].GetTaskId(), nil
	}
	return "", errors.New("unable to determine latest build task for package")
}

// PrettyPrintJSON takes a marshaled JSON byte array and prints it in a pretty format
func PrettyPrintJSON(data []byte) error {
	var prettyJSON map[string]interface{}

	// Unmarshal the JSON data into a map
	err := json.Unmarshal(data, &prettyJSON)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	// Marshal the map back to JSON with indentation for pretty printing
	formattedJSON, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	fmt.Println(string(formattedJSON))
	return nil
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func formatDuration(start, end time.Time) string {
	duration := end.Sub(start)
	return time.Time{}.Add(duration).Format("15:04:05")
}

func getLogLink(subtaskId string) string {
	return fmt.Sprintf("https://%s/api/v1/projects/%s/tasks/%s/logs",
		strings.Replace(endpoint(), "-api", "", 1),
		mustGetProjectID(), subtaskId)
}
