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

package cron

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"peridot.resf.org/koji"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/rherrata"
	"testing"
)

func TestInstance_CheckIfCVEResolvedDownstream_RHBA20212593_NotFixedDownstream(t *testing.T) {
	resetDb()

	errataMock.Advisories.Response.Docs = []*rherrata.CompactErrata{
		{
			Name:        "RHBA-2021:2593",
			Description: "",
			Synopsis:    "",
			Severity:    "None",
			Type:        "Bug Fix",
			AffectedPackages: []string{
				"cmake-3.18.2-11.el8_4.src.rpm",
				"cmake-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-data-3.18.2-11.el8_4.noarch.rpm",
				"cmake-debuginfo-3.18.2-11.el8_4.i686.rpm",
				"cmake-debuginfo-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-debugsource-3.18.2-11.el8_4.i686.rpm",
				"cmake-debugsource-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-doc-3.18.2-11.el8_4.noarch.rpm",
				"cmake-filesystem-3.18.2-11.el8_4.i686.rpm",
				"cmake-filesystem-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-gui-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-gui-debuginfo-3.18.2-11.el8_4.i686.rpm",
				"cmake-gui-debuginfo-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-rpm-macros-3.18.2-11.el8_4.noarch.rpm",
			},
			CVEs:            []string{},
			Fixes:           []string{},
			PublicationDate: "2021-06-29T00:00:00Z",
		},
	}

	cronInstance.ScanRedHatErrata()
	cronInstance.CheckIfCVEResolvedDownstream()

	affectedProducts, _ := cronInstance.db.GetAllAffectedProductsByCVE("RHBA-2021:2593")
	require.Len(t, affectedProducts, 1)
	require.Equal(t, "cmake-3.18.2-11.el8_4", affectedProducts[0].Package)
	require.Equal(t, int(secparseadminpb.AffectedProductState_FixedUpstream), affectedProducts[0].State)
}

func TestInstance_CheckIfCVEResolvedDownstream_RHBA20212593_FixedDownstream(t *testing.T) {
	resetDb()

	htmlFile, err := ioutil.ReadFile("testdata/RHBA-2021-2593.html")
	require.Nil(t, err)

	errataMock.HTMLResponses["RHBA-2021:2593"] = string(htmlFile[:])

	errataMock.Advisories.Response.Docs = []*rherrata.CompactErrata{
		{
			Name:        "RHBA-2021:2593",
			Description: "",
			Synopsis:    "",
			Severity:    "None",
			Type:        "Bug Fix",
			AffectedPackages: []string{
				"cmake-3.18.2-11.el8_4.src.rpm",
				"cmake-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-data-3.18.2-11.el8_4.noarch.rpm",
				"cmake-debuginfo-3.18.2-11.el8_4.i686.rpm",
				"cmake-debuginfo-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-debugsource-3.18.2-11.el8_4.i686.rpm",
				"cmake-debugsource-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-doc-3.18.2-11.el8_4.noarch.rpm",
				"cmake-filesystem-3.18.2-11.el8_4.i686.rpm",
				"cmake-filesystem-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-gui-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-gui-debuginfo-3.18.2-11.el8_4.i686.rpm",
				"cmake-gui-debuginfo-3.18.2-11.el8_4.x86_64.rpm",
				"cmake-rpm-macros-3.18.2-11.el8_4.noarch.rpm",
			},
			CVEs:            []string{},
			Fixes:           []string{},
			PublicationDate: "2021-06-29T00:00:00Z",
		},
	}

	cronInstance.ScanRedHatErrata()

	kojiMock.Tagged = []*koji.Build{
		{
			BuildId:         10,
			CompletionTime:  "",
			CompletionTs:    0,
			CreationEventId: 0,
			CreationTime:    "",
			CreationTs:      0,
			Epoch:           "",
			Extra: &koji.ListBuildsExtra{
				Source: &koji.ListBuildsExtraSource{
					OriginalUrl: "git+https://git.rockylinux.org/staging/rpms/cmake.git?#aa313111d4efd7cc6c36d41cd9fc29874d1e0740",
				},
			},
			Name:        "cmake",
			Nvr:         "cmake-3.18.2-11.el8_4",
			OwnerId:     0,
			OwnerName:   "distrobuild",
			PackageId:   0,
			PackageName: "cmake",
			Release:     "11.el8_4",
			Source:      "",
			StartTime:   "",
			StartTs:     0,
			State:       0,
			TaskId:      0,
			Version:     "3.18.2",
			VolumeId:    0,
			VolumeName:  "",
			TagId:       0,
			TagName:     "",
		},
	}

	kojiMock.RPMs = []*koji.RPM{
		{
			Arch:    "src",
			BuildId: 10,
			Name:    "cmake",
			Nvr:     "cmake-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "x86_64",
			BuildId: 10,
			Name:    "cmake",
			Nvr:     "cmake-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "x86_64",
			BuildId: 10,
			Name:    "cmake-gui",
			Nvr:     "cmake-gui-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "noarch",
			BuildId: 10,
			Name:    "cmake-doc",
			Nvr:     "cmake-doc-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "noarch",
			BuildId: 10,
			Name:    "cmake-rpm-macros",
			Nvr:     "cmake-rpm-macros-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "noarch",
			BuildId: 10,
			Name:    "cmake-data",
			Nvr:     "cmake-data-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "i686",
			BuildId: 10,
			Name:    "cmake-debuginfo",
			Nvr:     "cmake-debuginfo-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "x86_64",
			BuildId: 10,
			Name:    "cmake-debuginfo",
			Nvr:     "cmake-debuginfo-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "i686",
			BuildId: 10,
			Name:    "cmake-debugsource",
			Nvr:     "cmake-debugsource-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "x86_64",
			BuildId: 10,
			Name:    "cmake-debugsource",
			Nvr:     "cmake-debugsource-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "i686",
			BuildId: 10,
			Name:    "cmake-filesystem",
			Nvr:     "cmake-filesystem-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "x86_64",
			BuildId: 10,
			Name:    "cmake-filesystem",
			Nvr:     "cmake-filesystem-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "i686",
			BuildId: 10,
			Name:    "cmake-gui-debuginfo",
			Nvr:     "cmake-gui-debuginfo-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
		{
			Arch:    "x86_64",
			BuildId: 10,
			Name:    "cmake-gui-debuginfo",
			Nvr:     "cmake-gui-debuginfo-3.18.2-11.el8_4",
			Release: "11.el8_4",
			Version: "3.18.2",
		},
	}

	cronInstance.CheckIfCVEResolvedDownstream()

	affectedProducts, _ := cronInstance.db.GetAllAffectedProductsByCVE("RHBA-2021:2593")
	require.Len(t, affectedProducts, 1)
	require.Equal(t, "cmake-3.18.2-11.el8_4", affectedProducts[0].Package)
	require.Equal(t, int(secparseadminpb.AffectedProductState_FixedDownstream), affectedProducts[0].State)

	require.Len(t, mockDb.BuildReferences, 14)
	require.Equal(t, "10", mockDb.BuildReferences[0].KojiID)
}
