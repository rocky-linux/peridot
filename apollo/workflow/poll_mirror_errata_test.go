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
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rherrata"
	"testing"
)

func getPollRedHatErrataEnv() *testsuite.TestWorkflowEnvironment {
	env := testWfSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(controller.GetAllShortCodesActivity)
	env.RegisterActivity(controller.ProcessRedHatErrataShortCodeActivity)

	return env
}

func TestInstance_ScanRedHatErrata_RHSA20212595_Security_CVE(t *testing.T) {
	resetDb()

	errataMock.Advisories.Response.Docs = []*rherrata.CompactErrata{
		{
			Name:        "RHSA-2021:2595",
			Description: "",
			Synopsis:    "",
			Severity:    "Moderate",
			Type:        "Security",
			AffectedPackages: []string{
				"389-ds-base-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.src.rpm",
				"389-ds-base-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-debuginfo-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-debugsource-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-devel-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-legacy-tools-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-legacy-tools-debuginfo-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-libs-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-libs-debuginfo-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-snmp-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"389-ds-base-snmp-debuginfo-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.x86_64.rpm",
				"python3-lib389-1.4.3.16-16.module+el8.4.0+11446+fc96bc48.noarch.rpm",
			},
			CVEs: []string{
				"CVE-2021-3514",
			},
			Fixes: []string{
				"1952907",
				"1960720",
				"1968588",
				"1970791",
				"1972721",
				"1972738",
			},
			PublicationDate: "2021-06-29T00:00:00Z",
		},
	}

	env := getPollRedHatErrataEnv()
	env.ExecuteWorkflow(controller.PollRedHatErrataWorkflow)
	require.Nil(t, env.GetWorkflowError())

	cves, _ := controller.db.GetAllUnresolvedCVEs()
	require.Len(t, cves, 1)
	require.Equal(t, "CVE-2021-3514", cves[0].ID)
}

func TestInstance_ScanRedHatErrata_BugFix_Erratum(t *testing.T) {
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

	env := getPollRedHatErrataEnv()
	env.ExecuteWorkflow(controller.PollRedHatErrataWorkflow)
	require.Nil(t, env.GetWorkflowError())

	cves := mockDb.Cves
	require.Len(t, cves, 1)
	require.Equal(t, "RHBA-2021:2593", cves[0].ID)

	affectedProducts, _ := controller.db.GetAllAffectedProductsByCVE(cves[0].ID)
	require.Len(t, affectedProducts, 1)
	require.Equal(t, "cmake-3.18.2-11.el8_4", affectedProducts[0].Package)
	require.Equal(t, int(apollopb.AffectedProduct_STATE_FIXED_UPSTREAM), affectedProducts[0].State)
}
