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

package yummeta

import (
	"encoding/xml"
	"strings"
)

type PrimaryPackageVersion struct {
	Epoch string `xml:"epoch,attr,omitempty"`
	Ver   string `xml:"ver,attr,omitempty"`
	Rel   string `xml:"rel,attr,omitempty"`
}

type PrimaryPackageChecksum struct {
	Type  string `xml:"type,attr,omitempty"`
	PkgId string `xml:"pkgid,attr,omitempty"`
	Value string `xml:",chardata"`
}

type PrimaryPackageTime struct {
	File  string `xml:"file,attr,omitempty"`
	Build string `xml:"build,attr,omitempty"`
}

type PrimaryPackageSize struct {
	Package   string `xml:"package,attr,omitempty"`
	Installed string `xml:"installed,attr,omitempty"`
	Archive   string `xml:"archive,attr,omitempty"`
}

type PrimaryPackageLocation struct {
	Href string `xml:"href,attr,omitempty"`
}

type PrimaryRpmHeaderRange struct {
	Start string `xml:"start,attr,omitempty"`
	End   string `xml:"end,attr,omitempty"`
}

type PrimaryRpmEntry struct {
	Name  string `xml:"name,attr,omitempty"`
	Flags string `xml:"flags,attr,omitempty"`
	Epoch string `xml:"epoch,attr,omitempty"`
	Ver   string `xml:"ver,attr,omitempty"`
	Rel   string `xml:"rel,attr,omitempty"`
}

type PrimaryRpmEntries struct {
	RpmEntries []*PrimaryRpmEntry `xml:"rpm_entry,omitempty"`
}

type PrimaryRpmFile struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

type PrimaryPackageFormat struct {
	RpmLicense     string                 `xml:"rpm_license,omitempty"`
	RpmVendor      string                 `xml:"rpm_vendor"`
	RpmGroup       string                 `xml:"rpm_group,omitempty"`
	RpmBuildHost   string                 `xml:"rpm_buildhost,omitempty"`
	RpmSourceRpm   string                 `xml:"rpm_sourcerpm,omitempty"`
	RpmHeaderRange *PrimaryRpmHeaderRange `xml:"rpm_header-range,omitempty"`
	RpmProvides    *PrimaryRpmEntries     `xml:"rpm_provides,omitempty"`
	RpmRequires    *PrimaryRpmEntries     `xml:"rpm_requires,omitempty"`
	RpmObsoletes   *PrimaryRpmEntries     `xml:"rpm_obsoletes,omitempty"`
	RpmConflicts   *PrimaryRpmEntries     `xml:"rpm_conflicts,omitempty"`
	RpmRecommends  *PrimaryRpmEntries     `xml:"rpm_recommends,omitempty"`
	RpmSuggests    *PrimaryRpmEntries     `xml:"rpm_suggests,omitempty"`
	RpmSupplements *PrimaryRpmEntries     `xml:"rpm_supplements,omitempty"`
	RpmEnhances    *PrimaryRpmEntries     `xml:"rpm_enhances,omitempty"`
	File           []*PrimaryRpmFile      `xml:"file,omitempty"`
}

type PrimaryPackage struct {
	Type        string                  `xml:"type,attr,omitempty"`
	Name        string                  `xml:"name,omitempty"`
	Arch        string                  `xml:"arch,omitempty"`
	Version     *PrimaryPackageVersion  `xml:"version,omitempty"`
	Checksum    *PrimaryPackageChecksum `xml:"checksum,omitempty"`
	Summary     string                  `xml:"summary,omitempty"`
	Description string                  `xml:"description,omitempty"`
	Packager    string                  `xml:"packager,omitempty"`
	Url         string                  `xml:"url,omitempty"`
	Time        *PrimaryPackageTime     `xml:"time,omitempty"`
	Size        *PrimaryPackageSize     `xml:"size,omitempty"`
	Location    *PrimaryPackageLocation `xml:"location,omitempty"`
	Format      *PrimaryPackageFormat   `xml:"format,omitempty"`
}

type PrimaryRoot struct {
	XMLName      xml.Name          `xml:"metadata,omitempty"`
	Xmlns        string            `xml:"xmlns,attr,omitempty"`
	XmlnsRpm     string            `xml:"xmlns:rpm,attr,omitempty"`
	Rpm          string            `xml:"rpm,attr,omitempty,omitempty"`
	PackageCount int               `xml:"packages,attr,omitempty"`
	Packages     []*PrimaryPackage `xml:"package,omitempty"`
}

func MarshalPrimary(primary *PrimaryRoot) ([]byte, error) {
	ret, err := xml.Marshal(primary)
	if err != nil {
		return nil, err
	}

	return []byte(strings.ReplaceAll(string(ret), "rpm_", "rpm:")), nil
}

func UnmarshalPrimary(data []byte, primary *PrimaryRoot) error {
	return xml.Unmarshal([]byte(strings.ReplaceAll(string(data), "rpm:", "rpm_")), primary)
}
