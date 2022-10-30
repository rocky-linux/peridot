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
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"peridot.resf.org/servicecatalog"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// FlagConfig is some extra customization options
// for Cobra flags
type FlagConfig struct {
	Name         string
	DefaultPort  int
	DatabaseName *string
}

// NewFlagConfig returns a default FlagConfig
func NewFlagConfig() *FlagConfig {
	return &FlagConfig{
		DefaultPort: 8080,
	}
}

func defaultEnv(env string, def string) string {
	if e := os.Getenv(env); e != "" {
		return e
	}

	return def
}

func AddDBFlagsOnly(f *pflag.FlagSet, config *FlagConfig) {
	dname := ""
	if config.DatabaseName != nil {
		dname = *config.DatabaseName
	}

	// Default is dev only
	f.String("database.url", fmt.Sprintf("postgresql://postgres:postgres@127.0.0.1:%s/%sdev?sslmode=disable", defaultEnv("POSTGRES_PORT", "5432"), dname), "Database host")
}

// AddFlags add the default flags to the cmd of a grpc service
func AddFlags(f *pflag.FlagSet, config *FlagConfig) {
	defaultPort := config.DefaultPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		envNumPort, err := strconv.Atoi(envPort)
		if err != nil {
			logrus.Fatalf("could not use PORT, err: %s", err)
		}

		defaultPort = envNumPort
	}
	config.DefaultPort = defaultPort

	issuer := "https://hdr.build.resf.org/"
	jwks := fmt.Sprintf("%s/.well-known/jwks.json", servicecatalog.HydraPublic())
	if os.Getenv("RESF_ENV") == "" || (os.Getenv("RESF_ENV") == "dev" && os.Getenv("LOCALSTACK_ENDPOINT") != "") {
		issuer = "https://hdr-dev.internal.pdev.resf.localhost/"
	} else if os.Getenv("RESF_ENV") != "prod" {
		issuer = fmt.Sprintf("https://hdr-%s.internal.build.resf.org/", os.Getenv("RESF_ENV"))
	}

	f.String("oidc.issuer", issuer, "OpenID Connect Issuer for the authentication interceptor")
	f.String("oidc.jwks", jwks, "OpenID Connect JWKs for the authentication interceptor")

	f.String("api.port", strconv.Itoa(defaultPort), "Port to serve the REST service on")
	f.String("grpc.port", strconv.Itoa(defaultPort+1), "Port to serve the gRPC service on")

	AddDBFlagsOnly(f, config)

	f.Bool("production", false, "Enable production mode")

	BindOnly(f, config)
}

// BindOnly only binds the given flags to viper opposed to AddFlags
// which also adds flags
func BindOnly(f *pflag.FlagSet, config *FlagConfig) {
	name := config.Name
	if name == "" {
		name = "PERIDOT"
	}

	viper.SetEnvPrefix(name)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	err := viper.BindPFlags(f)
	if err != nil {
		logrus.Fatalf("could not bind flags to viper - %s", err)
	}
}

// Main is common actions that should be executed in main
func Main() {
	rand.Seed(time.Now().Unix())

	if dbURL := viper.GetString("database.url"); strings.Contains(dbURL, "REPLACEME") {
		escapedPassword := url.QueryEscape(os.Getenv("DATABASE_PASSWORD"))
		viper.Set("database.url", strings.Replace(dbURL, "REPLACEME", escapedPassword, 1))
	}

	if !viper.GetBool("production") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.Infoln("production mode activated")
	}
}
