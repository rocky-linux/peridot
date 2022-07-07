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
	"peridot.resf.org/secparse/rhsecurity"
	"testing"
)

func TestInstance_PollRedHatForNewCVEs_AddNewCVE(t *testing.T) {
	resetDb()

	securityMock.Cves = []*rhsecurity.CVE{
		{
			CVE:                 "CVE-2021-3602",
			Severity:            "moderate",
			PublicDate:          "2021-07-15T14:00:00Z",
			Advisories:          []string{},
			Bugzilla:            "1969264",
			BugzillaDescription: "CVE-2021-3602 buildah: Host environment variables leaked in build container when using chroot isolation",
			CvssScore:           nil,
			CvssScoringVector:   nil,
			CWE:                 "CWE-200",
			AffectedPackages:    nil,
			ResourceUrl:         "https://access.redhat.com/hydra/rest/securitydata/cve/CVE-2021-3602.json",
			Cvss3ScoringVector:  "CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:C/C:H/I:N/A:N",
			Cvss3Score:          "5.6",
		},
	}

	cronInstance.PollRedHatForNewCVEs()

	cves, _ := cronInstance.db.GetAllUnresolvedCVEs()
	require.Len(t, cves, 1)
	require.Equal(t, "CVE-2021-3602", cves[0].ID)
}

func TestPollRedHatForNewCVEs_SkipExistingCVE(t *testing.T) {
	resetDb()

	securityMock.Cves = []*rhsecurity.CVE{
		{
			CVE:                 "CVE-2021-3602",
			Severity:            "moderate",
			PublicDate:          "2021-07-15T14:00:00Z",
			Advisories:          []string{},
			Bugzilla:            "1969264",
			BugzillaDescription: "CVE-2021-3602 buildah: Host environment variables leaked in build container when using chroot isolation",
			CvssScore:           nil,
			CvssScoringVector:   nil,
			CWE:                 "CWE-200",
			AffectedPackages:    nil,
			ResourceUrl:         "https://access.redhat.com/hydra/rest/securitydata/cve/CVE-2021-3602.json",
			Cvss3ScoringVector:  "CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:C/C:H/I:N/A:N",
			Cvss3Score:          "5.6",
		},
	}

	cronInstance.PollRedHatForNewCVEs()

	cves, _ := cronInstance.db.GetAllUnresolvedCVEs()
	require.Len(t, cves, 1)
	require.Equal(t, "CVE-2021-3602", cves[0].ID)

	cronInstance.PollRedHatForNewCVEs()

	cves, _ = cronInstance.db.GetAllUnresolvedCVEs()
	require.Len(t, cves, 1)
	require.Equal(t, "CVE-2021-3602", cves[0].ID)
}
