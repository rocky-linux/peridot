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
	"errors"
	"fmt"
	"io"
)

const (
	cloneDirectory = "/var/peridot/peridot__rpmbuild_content"
)

type LogPipeFunc func(stdout io.ReadCloser, stderr io.ReadCloser)

var (
	ErrInvalidMode = errors.New("invalid mode")
	CmdDefaultArgs []string
	ExecPath       = "/usr/bin/rpmbuild"

	DefaultEnvs []string
)

func GetDefaultArgs() []string {
	return []string{"--define", fmt.Sprintf("_topdir %s", cloneDirectory)}
}

func GetCmdDefaultArgs() []string {
	return append(CmdDefaultArgs, GetDefaultArgs()...)
}

func ModeToFlags(mode Mode) ([]string, error) {
	var ret []string

	if mode&ModeRebuildSRPM != 0 {
		ret = []string{"--rebuild"}
	} else if mode&ModeBuildAll != 0 {
		ret = []string{"-ba"}
	} else if mode&ModeBuildSRPM != 0 {
		ret = []string{"-bs"}
	} else if mode&ModeBuildBinary != 0 {
		ret = []string{"-bb"}
	} else {
		return nil, ErrInvalidMode
	}

	if mode&ModeNoCheck != 0 {
		ret = append(ret, "--nocheck")
	}
	if mode&ModeNoDeps != 0 {
		ret = append(ret, "--nodeps")
	}

	return ret, nil
}

func GetCloneDirectory() string {
	return cloneDirectory
}
