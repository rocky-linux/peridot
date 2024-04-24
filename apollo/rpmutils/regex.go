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

package rpmutils

import (
	"regexp"

	"github.com/rocky-linux/srpmproc/pkg/rpmutils"
)

var (
	nvr               *regexp.Regexp
	nvrNoArch         *regexp.Regexp
	nvrUnusualRelease *regexp.Regexp
	epoch             *regexp.Regexp
	module            *regexp.Regexp
	dist              *regexp.Regexp
	moduleDist        *regexp.Regexp
	advisoryId        *regexp.Regexp
)

func NVR() *regexp.Regexp {
	if nvr == nil {
		nvr = regexp.MustCompile("^(\\S+)-([\\w~%.+^]+)-(\\w+(?:\\.[\\w~%+^]+)+?)(?:\\.(\\w+))?(?:\\.rpm)?$")
	}
	return nvr
}

func NVRNoArch() *regexp.Regexp {
	if nvrNoArch == nil {
		nvrNoArch = rpmutils.Nvr
	}
	return nvrNoArch
}

func NVRUnusualRelease() *regexp.Regexp {
	if nvrUnusualRelease == nil {
		nvrUnusualRelease = regexp.MustCompile("^(\\S+)-([\\w~%.+^]+)-(\\w+?)(?:\\.(\\w+))?(?:\\.rpm)?$")
	}
	return nvrUnusualRelease
}

func Epoch() *regexp.Regexp {
	if epoch == nil {
		epoch = regexp.MustCompile("(\\d+):")
	}
	return epoch
}

func Module() *regexp.Regexp {
	if module == nil {
		module = regexp.MustCompile("^(.+)-(.+)-([0-9]{19})\\.((?:.+){8})$")
	}
	return module
}

func Dist() *regexp.Regexp {
	if dist == nil {
		dist = regexp.MustCompile("(\\.el\\d(?:_\\d|))")
	}
	return dist
}

func ModuleDist() *regexp.Regexp {
	if moduleDist == nil {
		moduleDist = regexp.MustCompile("\\.module.+$")
	}
	return moduleDist
}

func AdvisoryId() *regexp.Regexp {
	if advisoryId == nil {
		advisoryId = regexp.MustCompile("^(.+)([SEB]A)-([0-9]{4}):([0-9]+)$")
	}
	return advisoryId
}
