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
	"peridot.resf.org/secparse/cron"
	"peridot.resf.org/secparse/db/connector"
	"peridot.resf.org/utils"
	"sync"
	"time"
)

var root = &cobra.Command{
	Use: "secparsecron",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	dname := "secparse"
	cnf.DatabaseName = &dname
	cnf.Name = "secparse"

	pflags := root.PersistentFlags()
	pflags.String("koji-endpoint", "https://koji.rockylinux.org/kojihub", "Koji endpoint to check for downstream fix")
	pflags.String("koji-compose", "dist-rocky8-compose", "Tag to source compose packages from")
	pflags.String("koji-module-compose", "dist-rocky8-module-compose", "Tag to source compose modules from")

	utils.AddFlags(pflags, cnf)
}

func mn(_ *cobra.Command, _ []string) {
	cronInstance, err := cron.New(connector.MustAuto())
	if err != nil {
		logrus.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		// Poll Red Hat for new advisories every two hours
		for {
			cronInstance.ScanRedHatErrata()
			cronInstance.PollRedHatForNewCVEs()
			time.Sleep(2 * time.Hour)
		}
	}()

	go func() {
		// Poll unresolved CVE status and update every hour
		for {
			cronInstance.UpdateCVEState()
			time.Sleep(time.Hour)
		}
	}()

	go func() {
		// Auto detect downstream builds when CVEs are fixed upstream (check every 10 minutes)
		for {
			cronInstance.CheckIfCVEResolvedDownstream()
			time.Sleep(10 * time.Minute)
		}
	}()

	go func() {
		// Create advisory for fixed CVEs (check every 10 minutes)
		for {
			cronInstance.CreateAdvisoryForFixedCVEs()
			time.Sleep(10 * time.Minute)
		}
	}()

	wg.Wait()
}

func main() {
	utils.Main()
	if err := root.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
