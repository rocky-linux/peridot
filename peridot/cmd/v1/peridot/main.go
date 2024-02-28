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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var root = &cobra.Command{
	Use: "peridot",
}

func init() {
	root.PersistentFlags().String("endpoint", "peridot-api.build.resf.org", "Peridot API endpoint")
	root.PersistentFlags().String("hdr-endpoint", "hdr.build.resf.org", "RESF OIDC endpoint")
	root.PersistentFlags().Bool("skip-ca-verify", false, "Whether to accept self-signed certificates")
	root.PersistentFlags().String("client-id", "", "Client ID for authentication")
	root.PersistentFlags().String("client-secret", "", "Client secret for authentication")
	root.PersistentFlags().String("project-id", "", "Peridot project ID")
	root.PersistentFlags().Bool("debug", false, "Debug mode")

	root.AddCommand(lookaside)
	lookaside.AddCommand(lookasideUpload)

	root.AddCommand(build)
	build.AddCommand(buildRpmImport)
	build.AddCommand(buildPackage)

	root.AddCommand(project)
	project.AddCommand(projectInfo)
	project.AddCommand(projectList)
	project.AddCommand(projectCreateHashedRepos)
	project.AddCommand(projectCatalogSync)

	root.AddCommand(impCmd)

	viper.SetEnvPrefix("PERIDOT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	err := viper.BindPFlags(root.PersistentFlags())
	if err != nil {
		log.Fatalf("could not bind pflags to viper - %s", err)
	}
}

func main() {
	if err := root.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func endpoint() string {
	return viper.GetString("endpoint")
}

func hdrEndpoint() string {
	return viper.GetString("hdr-endpoint")
}

func skipCaVerify() bool {
	return viper.GetBool("skip-ca-verify")
}

func getClientId() string {
	return viper.GetString("client-id")
}

func getClientSecret() string {
	return viper.GetString("client-secret")
}

func mustGetProjectID() string {
	ret := viper.GetString("project-id")
	if ret == "" {
		logrus.Fatal("project-id is required")
	}
	return ret
}

func debug() bool {
	return viper.GetBool("debug")
}
