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

package models

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"time"
)

type Package struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updatedAt" db:"updated_at"`

	Name                string                `json:"name" db:"name"`
	PackageType         peridotpb.PackageType `json:"packageType" db:"package_type"`
	PackageTypeOverride sql.NullInt32         `json:"packageTypeOverride" db:"package_type_override"`
	LastImportAt        sql.NullTime          `json:"lastImportAt" db:"last_import_at"`
	LastBuildAt         sql.NullTime          `json:"lastBuildAt" db:"last_build_at"`

	// Only useful for select queries
	Total int64 `json:"total" db:"total"`
}

type Packages []Package

type PackageVersion struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	PackageId string `json:"packageId" db:"package_id"`
	Version   string `json:"version" db:"version"`
	Release   string `json:"release" db:"release"`
}

type PackageVersions []PackageVersion

type ExtraOptions struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updatedAt" db:"updated_at"`

	ProjectId     string         `json:"projectId" db:"project_id"`
	PackageName   string         `json:"packageName" db:"package_name"`
	WithFlags     pq.StringArray `json:"withFlags" db:"with_flags"`
	WithoutFlags  pq.StringArray `json:"withoutFlags" db:"without_flags"`
	DependsOn     pq.StringArray `json:"dependsOn" db:"depends_on"`
	EnableModule  pq.StringArray `json:"enableModule" db:"enable_module"`
	DisableModule pq.StringArray `json:"disableModule" db:"disable_module"`
}

func (p *Package) ToProto() *peridotpb.Package {
	pkg := &peridotpb.Package{
		Id:           p.ID.String(),
		Name:         p.Name,
		Type:         p.PackageType,
		LastImportAt: utils.NullTimeToTimestamppb(p.LastImportAt),
		LastBuildAt:  utils.NullTimeToTimestamppb(p.LastBuildAt),
	}
	if p.PackageTypeOverride.Valid {
		pkg.Type = peridotpb.PackageType(p.PackageTypeOverride.Int32)
	}

	return pkg
}

func (p Packages) ToProto() (ret []*peridotpb.Package) {
	for _, v := range p {
		ret = append(ret, v.ToProto())
	}

	return ret
}

func (pv *PackageVersion) ToProto() *peridotpb.VersionRelease {
	return &peridotpb.VersionRelease{
		Version: wrapperspb.String(pv.Version),
		Release: wrapperspb.String(pv.Release),
	}
}
