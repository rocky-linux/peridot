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
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"time"
)

type Project struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updatedAt" db:"updated_at"`

	Name            string         `json:"name" db:"name"`
	MajorVersion    int            `json:"majorVersion" db:"major_version"`
	DistTagOverride sql.NullString `json:"distTagOverride" db:"dist_tag_override"`

	TargetGitlabHost   string         `json:"targetGitlabHost" db:"target_gitlab_host"`
	TargetPrefix       string         `json:"targetPrefix" db:"target_prefix"`
	TargetBranchPrefix string         `json:"targetBranchPrefix" db:"target_branch_prefix"`
	SourceGitHost      sql.NullString `json:"sourceGitHost" db:"source_git_host"`
	SourcePrefix       sql.NullString `json:"sourcePrefix" db:"source_prefix"`
	SourceBranchPrefix sql.NullString `json:"sourceBranchPrefix" db:"source_branch_prefix"`
	CdnUrl             sql.NullString `json:"cdnUrl" db:"cdn_url"`
	StrictMode         bool           `json:"strictMode" db:"strict_mode"`
	StreamMode         bool           `json:"'streamMode'" db:"stream_mode"`
	TargetVendor       string         `json:"targetVendor" db:"target_vendor"`
	AdditionalVendor   string         `json:"additionalVendor" db:"additional_vendor"`

	Archs            pq.StringArray `json:"archs" db:"archs"`
	BuildPoolType    sql.NullString `json:"buildPoolType" db:"build_pool_type"`
	FollowImportDist bool           `json:"followImportDist" db:"follow_import_dist"`
	BranchSuffix     sql.NullString `json:"branchSuffix" db:"branch_suffix"`
	GitMakePublic    bool           `json:"gitMakePublic" db:"git_make_public"`

	VendorMacro   sql.NullString `json:"vendorMacro" db:"vendor_macro"`
	PackagerMacro sql.NullString `json:"packagerMacro" db:"packager_macro"`

	SrpmStagePackages  pq.StringArray `json:"srpmStagePackages" db:"srpm_stage_packages"`
	BuildStagePackages pq.StringArray `json:"buildStagePackages" db:"build_stage_packages"`
}

type Projects []Project

type ProjectKey struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	ProjectId      uuid.UUID `json:"projectId" db:"project_id"`
	GitlabUsername string    `json:"gitlabUsername" db:"gitlab_username"`
	GitlabSecret   string    `json:"gitlabSecret" db:"gitlab_secret"`
}

func (p *Project) ToProto() *peridotpb.Project {
	distTag := fmt.Sprintf("el%d", p.MajorVersion)
	if p.DistTagOverride.Valid {
		distTag = p.DistTagOverride.String
	}

	return &peridotpb.Project{
		Id:                 p.ID.String(),
		CreatedAt:          timestamppb.New(p.CreatedAt),
		UpdatedAt:          utils.NullTimeToTimestamppb(p.UpdatedAt),
		Name:               wrapperspb.String(p.Name),
		MajorVersion:       wrapperspb.Int32(int32(p.MajorVersion)),
		Archs:              p.Archs,
		DistTag:            wrapperspb.String(distTag),
		TargetGitlabHost:   wrapperspb.String(p.TargetGitlabHost),
		TargetPrefix:       wrapperspb.String(p.TargetPrefix),
		TargetBranchPrefix: wrapperspb.String(p.TargetBranchPrefix),
		SourceGitHost:      utils.NullStringValueP(p.SourceGitHost),
		SourcePrefix:       utils.NullStringValueP(p.SourcePrefix),
		SourceBranchPrefix: utils.NullStringValueP(p.SourceBranchPrefix),
		CdnUrl:             utils.NullStringValueP(p.CdnUrl),
		StreamMode:         p.StreamMode,
		TargetVendor:       p.TargetVendor,
		AdditionalVendor:   wrapperspb.String(p.AdditionalVendor),
		FollowImportDist:   p.FollowImportDist,
		BranchSuffix:       utils.NullStringValueP(p.BranchSuffix),
		GitMakePublic:      p.GitMakePublic,
		VendorMacro:        utils.NullStringValueP(p.VendorMacro),
		PackagerMacro:      utils.NullStringValueP(p.PackagerMacro),
	}
}

func (p Projects) ToProto() (ret []*peridotpb.Project) {
	for _, v := range p {
		ret = append(ret, v.ToProto())
	}

	return ret
}
