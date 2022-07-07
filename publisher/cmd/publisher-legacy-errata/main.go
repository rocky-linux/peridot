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
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"peridot.resf.org/publisher/updateinfo/legacy"
	"peridot.resf.org/secparse/db/connector"
	"peridot.resf.org/utils"
)

var root = &cobra.Command{
	Use: "publisher-legacy-errata",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

var (
	repoDir      string
	from         string
	composeName  string
	productName  string
	productShort string
	republish    bool
)

func init() {
	dname := "secparse"
	cnf.DatabaseName = &dname
	cnf.Name = "publisher"

	pflags := root.PersistentFlags()
	pflags.StringVar(&repoDir, "repo-dir", "/mnt/repos-staging/pub/rocky", "Directory with composes")
	pflags.StringVar(&from, "from", "releng@rockylinux.org", "Email address of publisher")
	pflags.StringVar(&composeName, "compose-name", "", "Compose to use")
	pflags.StringVar(&productName, "product-name", "", "Product name")
	pflags.StringVar(&productShort, "product-short", "", "Product name (short)")
	pflags.BoolVar(&republish, "republish", false, "Republish (process published advisories as well, should be used if repodata is out of sync)")
	_ = root.MarkPersistentFlagRequired("compose-name")
	_ = root.MarkPersistentFlagRequired("product-name")
	_ = root.MarkPersistentFlagRequired("product-short")

	utils.AddDBFlagsOnly(pflags, cnf)
	utils.BindOnly(pflags, cnf)
}

func mn(_ *cobra.Command, _ []string) {
	scanner := &legacy.Scanner{
		FS: osfs.New(repoDir),
		DB: connector.MustAuto(),
	}
	err := scanner.ScanAndPublish(from, composeName, productName, productShort, republish)
	if err != nil {
		logrus.Fatalf("Could not scan and publish: %v", err)
	}
}

func main() {
	if err := root.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
