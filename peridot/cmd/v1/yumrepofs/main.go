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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	peridotcommon "peridot.resf.org/peridot/common"
	serverconnector "peridot.resf.org/peridot/db/connector"
	yumrepofsv1 "peridot.resf.org/peridot/yumrepofs/v1"
	"peridot.resf.org/utils"
)

var root = &cobra.Command{
	Use: "yumrepofs",
	Run: mn,
}

var cnf = utils.NewFlagConfig()

func init() {
	cnf.DefaultPort = 45002

	dname := "peridot"
	cnf.DatabaseName = &dname
	cnf.Name = "yumrepofs"

	peridotcommon.AddFlags(root.PersistentFlags())
	root.PersistentFlags().String("s3-assume-role", "", "S3 assume role")
	utils.AddFlags(root.PersistentFlags(), cnf)
}

func mn(_ *cobra.Command, _ []string) {
	sess, err := utils.NewAwsSessionNoLocalStack(&aws.Config{})
	if err != nil {
		logrus.Fatal(err)
	}

	s, err := yumrepofsv1.NewServer(serverconnector.MustAuto(), sess)
	if err != nil {
		logrus.Fatal(err)
	}
	s.Run()
}

func main() {
	utils.Main()
	if err := root.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
