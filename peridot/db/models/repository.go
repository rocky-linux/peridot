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
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"time"
)

type ExternalRepository struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	ProjectId      string `json:"projectId" db:"project_id"`
	Url            string `json:"url" db:"url"`
	Priority       int    `json:"priority" db:"priority"`
	ModuleHotfixes bool   `json:"moduleHotfixes" db:"module_hotfixes"`
}

type ExternalRepositories []ExternalRepository

type Repository struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	Name                  string         `json:"name" db:"name"`
	ProjectId             string         `json:"projectId" db:"project_id"`
	InternalOnly          bool           `json:"internalOnly" db:"internal_only"`
	Packages              pq.StringArray `json:"packages" db:"packages"`
	ExcludeFilter         pq.StringArray `json:"excludeFilter" db:"exclude_filter"`
	IncludeFilter         pq.StringArray `json:"includeFilter" db:"include_filter"`
	AdditionalMultilib    pq.StringArray `json:"additionalMultilib" db:"additional_multilib"`
	ExcludeMultilibFilter pq.StringArray `json:"excludeMultilibFilter" db:"exclude_multilib_filter"`
	Multilib              pq.StringArray `json:"multilib" db:"multilib"`
	GlobIncludeFilter     pq.StringArray `json:"globExcludeFilter" db:"glob_include_filter"`
}

func (r *Repository) ToProto() *peridotpb.Repository {
	return &peridotpb.Repository{
		Id:                    r.ID.String(),
		CreatedAt:             timestamppb.New(r.CreatedAt),
		Name:                  wrapperspb.String(r.Name),
		ProjectId:             wrapperspb.String(r.ProjectId),
		Packages:              r.Packages,
		ExcludeFilter:         r.ExcludeFilter,
		IncludeList:           r.IncludeFilter,
		AdditionalMultilib:    r.AdditionalMultilib,
		ExcludeMultilibFilter: r.ExcludeMultilibFilter,
		Multilib:              r.Multilib,
		GlobIncludeFilter:     r.GlobIncludeFilter,
	}
}

type Repositories []Repository

func (rs Repositories) ToProto() []*peridotpb.Repository {
	var result []*peridotpb.Repository
	for _, r := range rs {
		result = append(result, r.ToProto())
	}
	return result
}

type RepositoryRevision struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	ProjectRepoId string `json:"projectRepoId" db:"project_repo_id"`
	Arch          string `json:"arch" db:"arch"`

	RepomdXml          string         `json:"repomdXml" db:"repomd_xml"`
	PrimaryXml         string         `json:"primaryXml" db:"primary_xml"`
	FilelistsXml       string         `json:"filelistsXml" db:"filelists_xml"`
	OtherXml           string         `json:"otherXml" db:"other_xml"`
	UpdateinfoXml      string         `json:"updateinfoXml" db:"updateinfo_xml"`
	ModuleDefaultsYaml string         `json:"moduleDefaultsYaml" db:"module_defaults_yaml"`
	ModulesYaml        string         `json:"modulesYaml" db:"modules_yaml"`
	GroupsXml          string         `json:"groupsXml" db:"groups_xml"`
	UrlMappings        types.JSONText `json:"urlMappings" db:"url_mappings"`
}
