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

package workflow

import (
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"io/ioutil"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rherrata"
	"peridot.resf.org/koji"
	"testing"
)

func getDownstreamCheckEnv() *testsuite.TestWorkflowEnvironment {
	env := getPollRedHatErrataEnv()
	env.RegisterActivity(controller.UpdateCVEStateActivity)
	env.RegisterActivity(controller.DownstreamCVECheckActivity)

	return env
}

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

	env := getDownstreamCheckEnv()
	env.ExecuteWorkflow(controller.PollRedHatErrataWorkflow)
	require.Nil(t, env.GetWorkflowError())

	env = getDownstreamCheckEnv()
	env.ExecuteWorkflow(controller.DownstreamCVECheckWorkflow)
	require.Nil(t, env.GetWorkflowError())

	affectedProducts, _ := controller.db.GetAllAffectedProductsByCVE("RHBA-2021:2593")
	require.Len(t, affectedProducts, 1)
	require.Equal(t, "cmake-3.18.2-11.el8_4", affectedProducts[0].Package)
	require.Equal(t, int(apollopb.AffectedProduct_STATE_FIXED_UPSTREAM), affectedProducts[0].State)
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

	env := getDownstreamCheckEnv()
	env.ExecuteWorkflow(controller.PollRedHatErrataWorkflow)
	require.Nil(t, env.GetWorkflowError())

	env = getDownstreamCheckEnv()
	env.ExecuteWorkflow(controller.UpdateCVEStateWorkflow)
	require.Nil(t, env.GetWorkflowError())

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

	env = getDownstreamCheckEnv()
	env.ExecuteWorkflow(controller.DownstreamCVECheckWorkflow)
	require.Nil(t, env.GetWorkflowError())

	affectedProducts, _ := controller.db.GetAllAffectedProductsByCVE("RHBA-2021:2593")
	require.Len(t, affectedProducts, 1)
	require.Equal(t, "cmake-3.18.2-11.el8_4", affectedProducts[0].Package)
	require.Equal(t, int(apollopb.AffectedProduct_STATE_FIXED_DOWNSTREAM), affectedProducts[0].State)

	require.Len(t, mockDb.BuildReferences, 14)
	require.Equal(t, "10", mockDb.BuildReferences[0].KojiID.String)
}

/*func TestInstance_CheckIfCVEResolvedDownstream_RHSA20221642_FixedDownstream(t *testing.T) {
    resetDb()

    htmlFile, err := ioutil.ReadFile("testdata/RHSA-2022-1642.html")
    require.Nil(t, err)

    errataMock.HTMLResponses["RHSA-2022:1642"] = string(htmlFile[:])

    errataMock.Advisories.Response.Docs = []*rherrata.CompactErrata{
        {
            Name:        "RHSA-2022:1642",
            Description: "",
            Synopsis:    "",
            Severity:    "Important",
            Type:        "Security",
            AffectedPackages: []string{
                "zlib-1.2.11-18.el8_5.src.rpm",
                "zlib-1.2.11-18.el8_5.i686.rpm",
                "zlib-1.2.11-18.el8_5.x86_64.rpm",
                "zlib-debuginfo-1.2.11-18.el8_5.i686.rpm",
                "zlib-debuginfo-1.2.11-18.el8_5.x86_64.rpm",
                "zlib-debugsource-1.2.11-18.el8_5.i686.rpm",
                "zlib-debugsource-1.2.11-18.el8_5.x86_64.rpm",
                "zlib-devel-1.2.11-18.el8_5.i686.rpm",
                "zlib-devel-1.2.11-18.el8_5.x86_64.rpm",
            },
            CVEs:            []string{
                "CVE-2018-25032",
            },
            Fixes:           []string{},
            PublicationDate: "2022-04-28T00:00:00Z",
        },
    }

    env := getDownstreamCheckEnv()
    env.ExecuteWorkflow(controller.PollRedHatErrataWorkflow)
    require.Nil(t, env.GetWorkflowError())

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
                    OriginalUrl: "git+https://git.rockylinux.org/staging/rpms/zlib.git?#cc63be52ed1ba4f25d2015fd014558a3e7e19b08",
                },
            },
            Name:        "zlib",
            Nvr:         "zlib-1.2.11-18.el8_5",
            OwnerId:     0,
            OwnerName:   "distrobuild",
            PackageId:   0,
            PackageName: "zlib",
            Release:     "18.el8_5",
            Source:      "",
            StartTime:   "",
            StartTs:     0,
            State:       0,
            TaskId:      0,
            Version:     "1.2.11",
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
            Name:    "zlib",
            Nvr:     "zlib-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "i686",
            BuildId: 10,
            Name:    "zlib",
            Nvr:     "zlib-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "x86_64",
            BuildId: 10,
            Name:    "zlib",
            Nvr:     "zlib-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "i686",
            BuildId: 10,
            Name:    "zlib-debuginfo",
            Nvr:     "zlib-debuginfo-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "x86_64",
            BuildId: 10,
            Name:    "zlib-debuginfo",
            Nvr:     "zlib-debuginfo-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "i686",
            BuildId: 10,
            Name:    "zlib-debugsource",
            Nvr:     "zlib-debugsource-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "x86_64",
            BuildId: 10,
            Name:    "zlib-debugsource",
            Nvr:     "zlib-debugsource-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "i686",
            BuildId: 10,
            Name:    "zlib-devel",
            Nvr:     "zlib-devel-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
        {
            Arch:    "x86_64",
            BuildId: 10,
            Name:    "zlib-devel",
            Nvr:     "zlib-devel-1.2.11-18.el8_5",
            Release: "18.el8_5",
            Version: "1.2.11",
        },
    }

    env = getDownstreamCheckEnv()
    env.ExecuteWorkflow(controller.DownstreamCVECheckWorkflow)
    require.Nil(t, env.GetWorkflowError())

    affectedProducts, _ := controller.db.GetAllAffectedProductsByCVE("RHSA-2022:1642")
    require.Len(t, affectedProducts, 1)
    require.Equal(t, "zlib-1.2.11-18.el8_5", affectedProducts[0].Package)
    require.Equal(t, int(apollopb.AffectedProduct_STATE_FIXED_DOWNSTREAM), affectedProducts[0].State)

    require.Len(t, mockDb.BuildReferences, 14)
    require.Equal(t, "10", mockDb.BuildReferences[0].KojiID)
}*/
