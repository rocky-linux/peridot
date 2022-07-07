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
	bazelbuild "bazel.build/protobuf"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func callBazel(args ...string) []byte {
	cmd := exec.Command(
		"bazel",
		args...,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		logrus.Error(errOut.String())
		logrus.Fatal(err)
	}

	return out.Bytes()
}

func main() {
	// get bazel workspace
	// exit if not invoked with bazel
	searchDirectory := os.Getenv("BUILD_WORKING_DIRECTORY")
	if searchDirectory == "" {
		logrus.Fatal("error: BUILD_WORKING_DIRECTORY not found")
	}

	// change directory to bazel workspace
	err := os.Chdir(searchDirectory)
	if err != nil {
		logrus.Fatal(err)
	}

	goModContent, err := ioutil.ReadFile("go.mod")
	if err != nil {
		logrus.Fatalf("could not read go.mod: %v", err)
	}

	queryProto := callBazel("query", "kind(go_proto_library, //... except //third_party/...)", "--output", "proto")

	var query bazelbuild.QueryResult
	err = proto.Unmarshal(queryProto, &query)
	if err != nil {
		logrus.Fatal(err)
	}

	var replaceList []string
	for _, rule := range query.Target {
		if *rule.Rule.RuleClass != "go_proto_library" {
			continue
		}

		fullTarget := *rule.Rule.Name
		ruleName := strings.Split(fullTarget, ":")[1]

		for _, nstring := range rule.Rule.Attribute {
			if *nstring.Name == "importpath" {
				origImportPath := *nstring.StringValue
				importpath := origImportPath

				buildLocation := strings.Replace(*rule.Rule.Location, searchDirectory+"/", "", 1)
				buildLocation = filepath.Dir(strings.Split(buildLocation, ":")[0])
				buildLocation = fmt.Sprintf("bazel-bin/%s/%s_", buildLocation, ruleName)

				modDir := filepath.Join(buildLocation, importpath)
				if err := os.MkdirAll(modDir, 0755); err != nil && !os.IsExist(err) {
					logrus.Fatalf("could not generate directory for importpath %s", importpath)
				}

				if strings.HasSuffix(importpath, "/v1") {
					importpath = strings.TrimSuffix(importpath, "/v1")
					modDir = filepath.Join(modDir, "..")
				}

				modContent := []byte(fmt.Sprintf("module %s", importpath))
				if err := ioutil.WriteFile(filepath.Join(modDir, "go.mod"), modContent, 0644); err != nil {
					logrus.Fatalf("could not write go.mod file: %v", err)
				}

				/*dummyDir := modDir
				  if strings.HasSuffix(origImportPath, "/v1") {
				      dummyDir = filepath.Join(dummyDir, "v1")
				  }*/

				/*dummyContent := []byte("// this file is generated for mock purposes. do not check in please\npackage dummy")
				  if err := ioutil.WriteFile(filepath.Join(dummyDir, "dummy.go"), dummyContent, 0644); err != nil {
				      logrus.Fatalf("could not write dummy.go file: %v", err)
				  }*/

				replaceListElem := fmt.Sprintf("\t%s => ./%s", importpath, modDir)
				replaceList = append(replaceList, replaceListElem)
			}
		}
	}

	stringGoMod := string(goModContent)
	newContent := []string{"//gen:comment:this file is generated with nofussvendor. DO NOT EDIT"}
	inSyncReplaceStart := false

	for _, line := range strings.Split(stringGoMod, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "// sync-replace-end" {
			newContent = append(newContent, strings.Join(replaceList, "\n"))
			newContent = append(newContent, ")")
			inSyncReplaceStart = false
		}
		if !inSyncReplaceStart && !strings.HasPrefix(trimmedLine, "//gen:comment") {
			newContent = append(newContent, line)
		}
		if trimmedLine == "// sync-replace-start" {
			inSyncReplaceStart = true
			newContent = append(newContent, "replace (")
		}
	}

	if err := ioutil.WriteFile(filepath.Join(searchDirectory, "go.mod"), []byte(strings.Join(newContent, "\n")), 0644); err != nil {
		logrus.Fatalf("could not write end file go.mod: %v", err)
	}

	logrus.Infof("Added %d replace directives pointing to mock modules", len(replaceList))
}
