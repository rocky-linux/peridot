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
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"peridot.resf.org/peridot/db/models"
)

func (a *Access) GetExternalRepositoriesForProject(projectId string) (ret models.ExternalRepositories, err error) {
	err = a.query.Select(&ret, "select id, created_at, project_id, url, priority, module_hotfixes from external_repositories where project_id = $1 order by created_at desc", projectId)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetExternalRepository(projectId string, repoId string) (*models.ExternalRepository, error) {
	var r models.ExternalRepository
	err := a.query.Select(&r, "select id, created_at, project_id, url, priority, module_hotfixes from external_repositories where project_id = $1 and id = $2 order by created_at desc", projectId, repoId)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (a *Access) DeleteExternalRepositoryForProject(projectId string, id string) error {
	_, err := a.query.Exec("delete from external_repositories where project_id = $1 and id = $2", projectId, id)
	return err
}

func (a *Access) CreateExternalRepositoryForProject(projectId string, repoURL string, priority *int32, moduleHotfixes bool) (*models.ExternalRepository, error) {
	var ret models.ExternalRepository
	err := a.query.Get(
		&ret,
		`
        insert into external_repositories (project_id, url, priority)
        values ($1, $2, $3, $4)
        returning id, created_at, project_id, url, priority, module_hotfixes
        `,
		projectId,
		repoURL,
		priority,
		moduleHotfixes,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) FindRepositoriesForPackage(projectId string, pkg string, internalOnly bool) (ret models.Repositories, err error) {
	err = a.query.Select(
		&ret,
		`
		select
			id,
			created_at,
			name,
			project_id,
			packages,
			exclude_filter,
			include_filter,
			additional_multilib,
			exclude_multilib_filter,
			multilib,
			glob_include_filter
		from project_repos
		where
			project_id = $1
			and ($2 = any(packages) or packages = '{}')
			and internal_only = $3
		order by created_at desc
		`,
		projectId,
		pkg,
		internalOnly,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) FindRepositoriesForProject(projectId string, id *string, internalOnly bool) (ret models.Repositories, err error) {
	err = a.query.Select(
		&ret,
		`
		select
			id,
			created_at,
			name,
			project_id,
			packages,
			exclude_filter,
			include_filter,
			additional_multilib,
			exclude_multilib_filter,
			multilib,
			glob_include_filter
		from project_repos
		where
			project_id = $1
			and ($2 :: uuid is null or id = $2 :: uuid)
			and internal_only = $3
		order by created_at desc
		`,
		projectId,
		id,
		internalOnly,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetRepositoryRevision(revisionId string) (*models.RepositoryRevision, error) {
	var ret models.RepositoryRevision
	err := a.query.Get(
		&ret,
		`
		select
			id,
			created_at,
			project_repo_id,
			arch,
			repomd_xml,
			primary_xml,
			filelists_xml,
			other_xml,
			updateinfo_xml,
			module_defaults_yaml,
			modules_yaml,
            groups_xml,
			url_mappings
		from project_repo_revisions
		where
			id = $1
		`,
		revisionId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetLatestActiveRepositoryRevision(repoId string, arch string) (*models.RepositoryRevision, error) {
	var ret models.RepositoryRevision
	err := a.query.Get(
		&ret,
		`
		select
			id,
			created_at,
			project_repo_id,
			arch,
			repomd_xml,
			primary_xml,
			filelists_xml,
			other_xml,
			updateinfo_xml,
			module_defaults_yaml,
			modules_yaml,
			groups_xml,
			url_mappings
		from project_repo_revisions
		where
			project_repo_id = $1
			and arch = $2
		order by created_at desc
		limit 1
		`,
		repoId,
		arch,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(projectId string, name string, arch string) (*models.RepositoryRevision, error) {
	var ret models.RepositoryRevision
	err := a.query.Get(
		&ret,
		`
		select
			prr.id,
			prr.created_at,
			prr.project_repo_id,
			prr.arch,
			prr.repomd_xml,
			prr.primary_xml,
			prr.filelists_xml,
			prr.other_xml,
			prr.updateinfo_xml,
			prr.module_defaults_yaml,
			prr.modules_yaml,
			prr.groups_xml,
			prr.url_mappings
		from project_repo_revisions prr
		inner join project_repos pr on pr.id = prr.project_repo_id
		where
			pr.project_id = $1
			and pr.name = $2
			and prr.arch = $3
		order by prr.created_at desc
        limit 1
		`,
		projectId,
		name,
		arch,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) CreateRevisionForRepository(id string, repoId string, arch string, repomdXml string, primaryXml string, filelistsXml string, otherXml string, updateInfoXml string, moduleDefaultsYaml string, modulesYaml string, groupsXml string, urlMappings string) (*models.RepositoryRevision, error) {
	revision := models.RepositoryRevision{
		ProjectRepoId:      repoId,
		RepomdXml:          repomdXml,
		PrimaryXml:         primaryXml,
		FilelistsXml:       filelistsXml,
		OtherXml:           otherXml,
		UpdateinfoXml:      updateInfoXml,
		ModuleDefaultsYaml: moduleDefaultsYaml,
		ModulesYaml:        modulesYaml,
		GroupsXml:          groupsXml,
		UrlMappings:        types.JSONText(urlMappings),
	}

	err := a.query.Get(
		&revision,
		`
		insert into project_repo_revisions (id, project_repo_id, arch, repomd_xml, primary_xml, filelists_xml, other_xml, updateinfo_xml, module_defaults_yaml, modules_yaml, groups_xml, url_mappings)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		returning id, created_at
		`,
		id,
		repoId,
		arch,
		repomdXml,
		primaryXml,
		filelistsXml,
		otherXml,
		updateInfoXml,
		moduleDefaultsYaml,
		modulesYaml,
		groupsXml,
		urlMappings,
	)
	if err != nil {
		return nil, err
	}

	return &revision, nil
}

func (a *Access) CreateRepositoryWithPackages(name string, projectId string, internalOnly bool, packages pq.StringArray) (*models.Repository, error) {
	repository := models.Repository{
		Name:         name,
		ProjectId:    projectId,
		InternalOnly: internalOnly,
		Packages:     packages,
	}

	err := a.query.Get(
		&repository,
		`
		insert into project_repos (name, project_id, internal_only, packages)
		values ($1, $2, $3, $4)
		returning id, created_at
		`,
		name,
		projectId,
		internalOnly,
		packages,
	)
	if err != nil {
		return nil, err
	}

	return &repository, nil
}

func (a *Access) GetRepository(id *string, name *string, projectId *string) (*models.Repository, error) {
	var r models.Repository
	err := a.query.Get(
		&r,
		`
		select
			id,
			created_at,
			name,
			project_id,
			internal_only,
			packages,
			exclude_filter,
			include_filter,
			additional_multilib,
			exclude_multilib_filter,
			multilib,
			glob_include_filter
		from project_repos
		where
			(($2 :: text is null and id = $1 :: uuid) or name = $2 :: text)
			and ($3 :: uuid is null or project_id = $3 :: uuid)
		`,
		id,
		name,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (a *Access) SetRepositoryOptions(id string, packages pq.StringArray, excludeFilter pq.StringArray, includeFilter pq.StringArray, additionalMultilib pq.StringArray, excludeMultilibFilter pq.StringArray, multilib pq.StringArray, globIncludeFilter pq.StringArray) error {
	_, err := a.query.Exec(
		`
		update project_repos
		set
			packages = $1,
			exclude_filter = $2,
			include_filter = $3,
			additional_multilib = $4,
			exclude_multilib_filter = $5,
			multilib = $6,
			glob_include_filter = $7
		where id = $8
		`,
		packages,
		excludeFilter,
		includeFilter,
		additionalMultilib,
		excludeMultilibFilter,
		multilib,
		globIncludeFilter,
		id,
	)
	return err
}
