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

package updateinfo

import "encoding/xml"

const (
	TimeFormat = "2006-01-02 15:04:05"
)

type UpdatesRoot struct {
	XMLName xml.Name  `xml:"updates"`
	Updates []*Update `xml:"update"`
}

type UpdateDate struct {
	Date string `xml:"date,attr"`
}

type UpdateReference struct {
	Href  string `xml:"href,attr"`
	ID    string `xml:"id,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
}

type UpdateReferenceRoot struct {
	References []*UpdateReference `xml:"reference"`
}

type UpdatePackageSum struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type UpdatePackage struct {
	Name            string              `xml:"name,attr"`
	Version         string              `xml:"version,attr"`
	Release         string              `xml:"release,attr"`
	Epoch           string              `xml:"epoch,attr"`
	Arch            string              `xml:"arch,attr"`
	Src             string              `xml:"src,attr"`
	Filename        string              `xml:"filename"`
	RebootSuggested string              `xml:"reboot_suggested"`
	Sum             []*UpdatePackageSum `xml:"sum"`
}

type UpdateCollection struct {
	Short    string           `xml:"short,attr"`
	Name     string           `xml:"name"`
	Packages []*UpdatePackage `xml:"package"`
}

type UpdateCollectionRoot struct {
	Collections []*UpdateCollection `xml:"collection"`
}

type Update struct {
	From        string                `xml:"from,attr"`
	Status      string                `xml:"status,attr"`
	Type        string                `xml:"type,attr"`
	Version     string                `xml:"version,attr"`
	ID          string                `xml:"id"`
	Title       string                `xml:"title"`
	Issued      *UpdateDate           `xml:"issued"`
	Updated     *UpdateDate           `xml:"updated"`
	Rights      string                `xml:"rights"`
	Release     string                `xml:"release"`
	PushCount   string                `xml:"pushcount"`
	Severity    string                `xml:"severity"`
	Summary     string                `xml:"summary"`
	Description string                `xml:"description"`
	References  *UpdateReferenceRoot  `xml:"references"`
	PkgList     *UpdateCollectionRoot `xml:"pkglist"`
}
