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

package rherrata

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	apollopb "peridot.resf.org/apollo/pb"
	"testing"
)

func newInstance() *MockInstance {
	return NewMock()
}

func TestRHBA20212759(t *testing.T) {
	mock := newInstance()

	htmlFile, err := ioutil.ReadFile("testdata/RHBA-2021-2759.html")
	require.Nil(t, err)

	mock.HTMLResponses["RHBA-2021:2759"] = string(htmlFile[:])

	errata, err := mock.API.GetErrata("RHBA-2021:2759")
	require.Nil(t, err)

	require.Equal(t, "firefox bugfix update", errata.Synopsis)
	require.Equal(t, apollopb.Advisory_TYPE_BUGFIX, errata.Type)
	require.Len(t, errata.Topic, 1)
	require.Equal(t, "An update for firefox is now available for Red Hat Enterprise Linux 8.", errata.Topic[0])
	require.Len(t, errata.Description, 3)
	require.Equal(t, "Mozilla Firefox is an open-source web browser, designed for standards", errata.Description[0])
	require.Equal(t, "compliance, performance, and portability.", errata.Description[1])
	require.Equal(t, "This update upgrades Firefox to version 78.12.0 ESR.", errata.Description[2])
	require.Len(t, errata.AffectedProducts, 12)
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for x86_64 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for x86_64 - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server - AUS 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for IBM z Systems 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for IBM z Systems - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for Power, little endian 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for Power, little endian - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server - TUS 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for ARM 64 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for ARM 64 - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server (for IBM Power LE) - Update Services for SAP Solutions 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server - Update Services for SAP Solutions 8.4"])

	x86 := errata.AffectedProducts["Red Hat Enterprise Linux for x86_64 8"]
	require.Len(t, x86.SRPMs, 1)
	require.Equal(t, "firefox-78.12.0-2.el8_4.src.rpm", x86.SRPMs[0])
	require.Len(t, x86.Packages[ArchX8664], 3)
	require.Equal(t, "firefox-78.12.0-2.el8_4.x86_64.rpm", x86.Packages[ArchX8664][0])
	require.Equal(t, "firefox-debuginfo-78.12.0-2.el8_4.x86_64.rpm", x86.Packages[ArchX8664][1])
	require.Equal(t, "firefox-debugsource-78.12.0-2.el8_4.x86_64.rpm", x86.Packages[ArchX8664][2])
}

func TestRHBA20212743(t *testing.T) {
	mock := newInstance()

	htmlFile, err := ioutil.ReadFile("testdata/RHSA-2021-2743.html")
	require.Nil(t, err)

	mock.HTMLResponses["RHSA-2021:2743"] = string(htmlFile[:])

	errata, err := mock.API.GetErrata("RHSA-2021:2743")
	require.Nil(t, err)

	require.Equal(t, "Important: firefox security update", errata.Synopsis)
	require.Equal(t, apollopb.Advisory_TYPE_SECURITY, errata.Type)
	require.Equal(t, apollopb.Advisory_SEVERITY_IMPORTANT, errata.Severity)
	require.Len(t, errata.Topic, 2)
	require.Equal(t, "An update for firefox is now available for Red Hat Enterprise Linux 8.", errata.Topic[0])
	require.Equal(t, "Red Hat Product Security has rated this update as having a security impact of Important. A Common Vulnerability Scoring System (CVSS) base score, which gives a detailed severity rating, is available for each vulnerability from the CVE link(s) in the References section.", errata.Topic[1])
	require.Len(t, errata.Description, 3)
	require.Equal(t, "Mozilla Firefox is an open-source web browser, designed for standards compliance, performance, and portability.", errata.Description[0])
	require.Equal(t, "This update upgrades Firefox to version 78.12.0 ESR.", errata.Description[1])
	require.Equal(t, "For more details about the security issue(s), including the impact, a CVSS score, acknowledgments, and other related information, refer to the CVE page(s) listed in the References section.", errata.Description[2])
	require.Len(t, errata.AffectedProducts, 12)
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for x86_64 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for x86_64 - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server - AUS 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for IBM z Systems 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for IBM z Systems - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for Power, little endian 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for Power, little endian - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server - TUS 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for ARM 64 8"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux for ARM 64 - Extended Update Support 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server (for IBM Power LE) - Update Services for SAP Solutions 8.4"])
	require.NotNil(t, errata.AffectedProducts["Red Hat Enterprise Linux Server - Update Services for SAP Solutions 8.4"])
	require.Len(t, errata.Fixes, 3)
	require.Equal(t, "1970109", errata.Fixes[0].BugzillaID)
	require.Equal(t, "CVE-2021-30547 chromium-browser: Out of bounds write in ANGLE", errata.Fixes[0].Description)
	require.Equal(t, "1982013", errata.Fixes[1].BugzillaID)
	require.Equal(t, "CVE-2021-29970 Mozilla: Use-after-free in accessibility features of a document", errata.Fixes[1].Description)
	require.Equal(t, "1982014", errata.Fixes[2].BugzillaID)
	require.Equal(t, "CVE-2021-29976 Mozilla: Memory safety bugs fixed in Firefox 90 and Firefox ESR 78.12", errata.Fixes[2].Description)
	require.Len(t, errata.CVEs, 3)
	require.Equal(t, "CVE-2021-29970", errata.CVEs[0])
	require.Equal(t, "CVE-2021-29976", errata.CVEs[1])
	require.Equal(t, "CVE-2021-30547", errata.CVEs[2])
	require.Len(t, errata.References, 1)
	require.Equal(t, "https://access.redhat.com/security/updates/classification/#important", errata.References[0])

	x86 := errata.AffectedProducts["Red Hat Enterprise Linux for x86_64 8"]
	require.Len(t, x86.SRPMs, 1)
	require.Equal(t, "firefox-78.12.0-1.el8_4.src.rpm", x86.SRPMs[0])
	require.Len(t, x86.Packages[ArchX8664], 3)
	require.Equal(t, "firefox-78.12.0-1.el8_4.x86_64.rpm", x86.Packages[ArchX8664][0])
	require.Equal(t, "firefox-debuginfo-78.12.0-1.el8_4.x86_64.rpm", x86.Packages[ArchX8664][1])
	require.Equal(t, "firefox-debugsource-78.12.0-1.el8_4.x86_64.rpm", x86.Packages[ArchX8664][2])
}
