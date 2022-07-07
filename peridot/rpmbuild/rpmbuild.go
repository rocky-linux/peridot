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

package rpmbuild

import (
	"bufio"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type Mode int

const (
	ModeBuildSRPM Mode = 1 << iota
	ModeRebuildSRPM
	ModeBuildAll
	ModeBuildBinary
	ModeNoCheck
	ModeNoDeps
	ModePrivileged
)

type Options struct {
	With    []string
	Without []string
}

type Access interface {
	// Exec executes a rpmbuild command
	// Returns (stdout, error)
	Exec(mode Mode, arch string, filePath string, options Options) error
}

type Instance struct {
	FS billy.Filesystem
}

func New(fs billy.Filesystem) *Instance {
	return &Instance{
		FS: fs,
	}
}

func (i *Instance) Exec(mode Mode, arch string, filePath string, options Options) error {
	mFlags, err := ModeToFlags(mode)
	if err != nil {
		return err
	}

	if arch != "" {
		mFlags = append(mFlags, "--target="+arch)
	}
	if options.With != nil {
		for _, v := range options.With {
			mFlags = append(mFlags, "--with="+v)
		}
	}
	if options.Without != nil {
		for _, v := range options.Without {
			mFlags = append(mFlags, "--without="+v)
		}
	}
	args := append(GetCmdDefaultArgs(), append(mFlags, filePath)...)

	shCommand := fmt.Sprintf("%s %s", ExecPath, strings.Join(args, " "))
	fmt.Printf("[+] %s\n", shCommand)
	cmd := exec.Command(ExecPath, args...)
	cmd.Env = DefaultEnvs

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(stdout io.ReadCloser) {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}(stdoutPipe)
	go func(stderr io.ReadCloser) {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}(stderrPipe)

	wg.Wait()

	return cmd.Wait()
}
