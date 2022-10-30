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
	apolloconnector "peridot.resf.org/apollo/db/connector"
	apolloimpl "peridot.resf.org/apollo/impl/v1"
	"peridot.resf.org/utils"
)

var root = &cobra.Command{
	Use: "apollo",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	cnf.DefaultPort = 9100

	cnf.DatabaseName = utils.Pointer[string]("apollo")
	cnf.Name = *cnf.DatabaseName

	root.PersistentFlags().String("homepage", "https://errata.build.resf.org", "Frontend root URL")
	utils.AddFlags(root.PersistentFlags(), cnf)
}

func mn(_ *cobra.Command, _ []string) {
	apolloimpl.NewServer(apolloconnector.MustAuto()).Run()
}

func main() {
	utils.Main()
	if err := root.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
