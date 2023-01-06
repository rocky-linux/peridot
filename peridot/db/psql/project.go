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

package serverpsql

import (
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (a *Access) ListProjects(filters *peridotpb.ProjectFilters) (ret models.Projects, err error) {
	if filters == nil {
		filters = &peridotpb.ProjectFilters{}
	}

	var ids pq.StringArray = nil
	if filters.Ids != nil {
		ids = filters.Ids
	}

	err = a.query.Select(
		&ret,
		`
		select
			id,
			created_at,
			updated_at,
			name,
			major_version,
			dist_tag_override,
			target_gitlab_host,
			target_prefix,
			target_branch_prefix,
			source_git_host,
			source_prefix,
			source_branch_prefix,
			cdn_url,
			strict_mode,
			target_vendor,
			additional_vendor,
			archs,
			follow_import_dist,
			branch_suffix,
			git_make_public,
			vendor_macro,
			packager_macro
		from projects
		where
			($1 :: uuid is null or id = $1 :: uuid)
			and ($2 :: text is null or name ~* $2 :: text)
			and ($3 :: uuid[] is null or id = any($3 :: uuid[]))
		order by created_at desc
		`,
		utils.StringValueP(filters.Id),
		utils.StringValueP(filters.Name),
		ids,
	)

	return ret, err
}

func (a *Access) GetProjectKeys(projectId string) (*models.ProjectKey, error) {
	var ret models.ProjectKey
	err := a.query.Get(
		&ret,
		`
		select
			id,
			created_at,
			project_id,
			gitlab_username,
			gitlab_secret
		from project_keys
		where project_id = $1
		`,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// GetProjectModuleConfiguration returns the module configurations for the given project.
func (a *Access) GetProjectModuleConfiguration(projectId string) (*peridotpb.ModuleConfiguration, error) {
	var ret types.JSONText
	err := a.query.Get(
		&ret,
		`
        select
            proto
        from project_module_configuration
        where
			project_id = $1
			and active = true
        `,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	anyPb := &anypb.Any{}
	err = protojson.Unmarshal(ret, anyPb)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal module configuration (protojson): %v", err)
	}

	pb := &peridotpb.ModuleConfiguration{}
	err = anypb.UnmarshalTo(anyPb, pb, proto.UnmarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal module configuration: %v", err)
	}

	return pb, nil
}

func (a *Access) CreateProjectModuleConfiguration(projectId string, config *peridotpb.ModuleConfiguration) error {
	anyPb, err := anypb.New(config)
	if err != nil {
		return fmt.Errorf("failed to marshal module configuration: %v", err)
	}

	protoJson, err := protojson.Marshal(anyPb)
	if err != nil {
		return fmt.Errorf("failed to marshal module configuration (protojson): %v", err)
	}

	_, err = a.query.Exec(
		`
        insert into project_module_configuration (project_id, proto, active)
        values ($1, $2, true)
		on conflict (project_id) do update
			set proto = $2, active = true
        `,
		projectId,
		protoJson,
	)
	if err != nil {
		return err
	}

	return nil
}

func (a *Access) CreateProject(project *peridotpb.Project) (*models.Project, error) {
	if err := project.ValidateAll(); err != nil {
		return nil, err
	}

	ret := models.Project{
		Name:               project.Name.Value,
		MajorVersion:       int(project.MajorVersion.Value),
		DistTagOverride:    utils.StringValueToNullString(project.DistTag),
		TargetGitlabHost:   project.TargetGitlabHost.Value,
		TargetPrefix:       project.TargetPrefix.Value,
		TargetBranchPrefix: project.TargetBranchPrefix.Value,
		SourceGitHost:      utils.StringValueToNullString(project.SourceGitHost),
		SourcePrefix:       utils.StringValueToNullString(project.SourcePrefix),
		SourceBranchPrefix: utils.StringValueToNullString(project.SourceBranchPrefix),
		CdnUrl:             utils.StringValueToNullString(project.CdnUrl),
		StreamMode:         project.StreamMode,
		TargetVendor:       project.TargetVendor,
		AdditionalVendor:   project.AdditionalVendor.Value,
		Archs:              project.Archs,
		FollowImportDist:   project.FollowImportDist,
		BranchSuffix:       utils.StringValueToNullString(project.BranchSuffix),
		GitMakePublic:      project.GitMakePublic,
		VendorMacro:        utils.StringValueToNullString(project.VendorMacro),
		PackagerMacro:      utils.StringValueToNullString(project.PackagerMacro),
	}

	err := a.query.Get(
		&ret,
		`
		insert into projects
		(name, major_version, dist_tag_override, target_gitlab_host, target_prefix,
		target_branch_prefix, source_git_host, source_prefix, source_branch_prefix, cdn_url,
		stream_mode, target_vendor, additional_vendor, archs, follow_import_dist, branch_suffix, git_make_public, vendor_macro, packager_macro)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		returning id, created_at, updated_at
		`,
		ret.Name,
		ret.MajorVersion,
		ret.DistTagOverride,
		ret.TargetGitlabHost,
		ret.TargetPrefix,
		ret.TargetBranchPrefix,
		ret.SourceGitHost,
		ret.SourcePrefix,
		ret.SourceBranchPrefix,
		ret.CdnUrl,
		ret.StreamMode,
		ret.TargetVendor,
		ret.AdditionalVendor,
		ret.Archs,
		ret.FollowImportDist,
		ret.BranchSuffix,
		ret.GitMakePublic,
		ret.VendorMacro,
		ret.PackagerMacro,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) UpdateProject(id string, project *peridotpb.Project) (*models.Project, error) {
	if err := project.ValidateAll(); err != nil {
		return nil, err
	}

	ret := models.Project{
		Name:               project.Name.Value,
		MajorVersion:       int(project.MajorVersion.Value),
		DistTagOverride:    utils.StringValueToNullString(project.DistTag),
		TargetGitlabHost:   project.TargetGitlabHost.Value,
		TargetPrefix:       project.TargetPrefix.Value,
		TargetBranchPrefix: project.TargetBranchPrefix.Value,
		SourceGitHost:      utils.StringValueToNullString(project.SourceGitHost),
		SourcePrefix:       utils.StringValueToNullString(project.SourcePrefix),
		SourceBranchPrefix: utils.StringValueToNullString(project.SourceBranchPrefix),
		CdnUrl:             utils.StringValueToNullString(project.CdnUrl),
		StreamMode:         project.StreamMode,
		TargetVendor:       project.TargetVendor,
		AdditionalVendor:   project.AdditionalVendor.Value,
		Archs:              project.Archs,
		FollowImportDist:   project.FollowImportDist,
		BranchSuffix:       utils.StringValueToNullString(project.BranchSuffix),
		GitMakePublic:      project.GitMakePublic,
		VendorMacro:        utils.StringValueToNullString(project.VendorMacro),
		PackagerMacro:      utils.StringValueToNullString(project.PackagerMacro),
	}

	err := a.query.Get(
		&ret,
		`
		update projects set
			name = $1,
			major_version = $2,
			dist_tag_override = $3,
			target_gitlab_host = $4,
			target_prefix = $5,
			target_branch_prefix = $6,
			source_git_host = $7,
			source_prefix = $8,
			source_branch_prefix = $9,
			cdn_url = $10,
			stream_mode = $11,
			target_vendor = $12,
			additional_vendor = $13,
			archs = $14,
			follow_import_dist = $15,
			branch_suffix = $16,
			git_make_public = $17,
            vendor_macro = $18,
            packager_macro = $19,
			updated_at = now()
		where id = $20
		returning id, created_at, updated_at
		`,
		ret.Name,
		ret.MajorVersion,
		ret.DistTagOverride,
		ret.TargetGitlabHost,
		ret.TargetPrefix,
		ret.TargetBranchPrefix,
		ret.SourceGitHost,
		ret.SourcePrefix,
		ret.SourceBranchPrefix,
		ret.CdnUrl,
		ret.StreamMode,
		ret.TargetVendor,
		ret.AdditionalVendor,
		ret.Archs,
		ret.FollowImportDist,
		ret.BranchSuffix,
		ret.GitMakePublic,
		ret.VendorMacro,
		ret.PackagerMacro,
		id,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) SetProjectKeys(projectId string, username string, password string) error {
	_, err := a.query.Exec(
		`
		insert into project_keys (project_id, gitlab_username, gitlab_secret)
		values ($1, $2, $3)
		on conflict (project_id) do update
			set gitlab_username = $2, gitlab_secret = $3
		`,
		projectId,
		username,
		password,
	)
	return err
}

func (a *Access) SetBuildRootPackages(projectId string, srpmPackages pq.StringArray, buildPackages pq.StringArray) error {
	if srpmPackages == nil {
		srpmPackages = pq.StringArray{}
	}
	if buildPackages == nil {
		buildPackages = pq.StringArray{}
	}

	_, err := a.query.Exec(
		`
		update projects set srpm_stage_packages = $2, build_stage_packages = $3
    where project_id = $1
		`,
		projectId,
		srpmPackages,
		buildPackages,
	)
	return err
}
