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

import "encoding/xml"

const (
	TimeFormat = "2006-01-02 15:04:05"
)

type UpdatesRoot struct {
	XMLName xml.Name  `xml:"updates,omitempty"`
	Updates []*Update `xml:"update,omitempty"`
}

type UpdateDate struct {
	Date string `xml:"date,attr,omitempty"`
}

type UpdateReference struct {
	Href  string `xml:"href,attr,omitempty"`
	ID    string `xml:"id,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`
	Title string `xml:"title,attr,omitempty"`
}

type UpdateReferenceRoot struct {
	References []*UpdateReference `xml:"reference,omitempty"`
}

type UpdatePackageSum struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

type UpdatePackage struct {
	Name     string              `xml:"name,attr,omitempty"`
	Version  string              `xml:"version,attr,omitempty"`
	Release  string              `xml:"release,attr,omitempty"`
	Epoch    string              `xml:"epoch,attr,omitempty"`
	Arch     string              `xml:"arch,attr,omitempty"`
	Src      string              `xml:"src,attr,omitempty"`
	Filename string              `xml:"filename,omitempty"`
	Sum      []*UpdatePackageSum `xml:"sum,omitempty"`
}

type UpdateCollection struct {
	Short    string           `xml:"short,attr,omitempty"`
	Name     string           `xml:"name,omitempty"`
	Packages []*UpdatePackage `xml:"package,omitempty"`
}

type UpdateCollectionRoot struct {
	Collections []*UpdateCollection `xml:"collection,omitempty"`
}

type Update struct {
	From        string                `xml:"from,attr,omitempty"`
	Status      string                `xml:"status,attr,omitempty"`
	Type        string                `xml:"type,attr,omitempty"`
	Version     string                `xml:"version,attr,omitempty"`
	ID          string                `xml:"id,omitempty"`
	Title       string                `xml:"title,omitempty"`
	Issued      *UpdateDate           `xml:"issued,omitempty"`
	Updated     *UpdateDate           `xml:"updated,omitempty"`
	Rights      string                `xml:"rights,omitempty"`
	Release     string                `xml:"release,omitempty"`
	PushCount   string                `xml:"pushcount,omitempty"`
	Severity    string                `xml:"severity,omitempty"`
	Summary     string                `xml:"summary,omitempty"`
	Description string                `xml:"description,omitempty"`
	References  *UpdateReferenceRoot  `xml:"references,omitempty"`
	PkgList     *UpdateCollectionRoot `xml:"pkglist,omitempty"`
}
