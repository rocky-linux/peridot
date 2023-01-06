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
	"github.com/lib/pq"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (a *Access) GetPackagesInProject(filters *peridotpb.PackageFilters, projectId string, page int32, limit int32) (models.Packages, error) {
	if filters == nil {
		filters = &peridotpb.PackageFilters{}
	}

	var p models.Packages
	err := a.query.Select(
		&p,
		`
		select
			p.id,
			p.created_at,
			p.updated_at,
			p.name,
			p.package_type,
			proj_p.package_type_override,
			(
				select
					ir.created_at
				from import_revisions ir
				where package_version_id = (
					select
						proj_pv.package_version_id
					from project_package_versions proj_pv
					where
						proj_pv.package_id = p.id
                        and proj_pv.project_id = proj_p.project_id
						and proj_pv.active = true
					order by proj_pv.created_at desc limit 1
				)
				order by ir.created_at desc limit 1
			) as last_import_at,
			(
				select
					b.created_at
				from builds b
				inner join tasks t on t.id = b.task_id
				where
					t.status = 3
					and package_version_id = (
						select package_version_id
						from project_package_versions proj_pv
						where
							proj_pv.package_id = p.id
                            and proj_pv.project_id = proj_p.project_id
							and proj_pv.active = true
						order by proj_pv.created_at desc limit 1
					)
				order by b.created_at desc limit 1
			) as last_build_at,
			count(p.*) over() as total
		from packages p
		inner join project_packages proj_p on proj_p.package_id = p.id
		where
			($1 :: uuid is null or p.id = $1 :: uuid)
			and ($2 :: text is null or p.name ~* $2 :: text)
			and ($3 :: bool is null or p.package_type in (6))
			and ($4 :: text is null or p.name = $4 :: text)
			and ($5 :: bool is null or (
				select
					ir.created_at
				from import_revisions ir
				where package_version_id = (
					select
						proj_pv.package_version_id
					from project_package_versions proj_pv
					where
						proj_pv.package_id = p.id
                        and proj_pv.project_id = proj_p.project_id
						and proj_pv.active = true
					order by proj_pv.created_at desc limit 1
				)
				order by ir.created_at desc limit 1
			) is null)
			and ($6 :: bool is null or (
				select
					b.created_at
				from builds b
				inner join tasks t on t.id = b.task_id
				where
					t.status = 3
					and package_version_id = (
						select package_version_id
						from project_package_versions proj_pv
						where
							proj_pv.package_id = p.id
                            and proj_pv.project_id = proj_p.project_id
							and proj_pv.active = true
						order by proj_pv.created_at desc limit 1
					)
				order by b.created_at desc limit 1
			) is null)
			and proj_p.project_id = $7 :: uuid
		order by created_at desc, id
		limit $8 offset $9
		`,
		utils.StringValueP(filters.Id),
		utils.StringValueP(filters.Name),
		utils.BoolValueP(filters.Modular),
		utils.StringValueP(filters.NameExact),
		utils.BoolValueP(filters.NoImports),
		utils.BoolValueP(filters.NoBuilds),
		projectId,
		utils.UnlimitedLimit(limit),
		utils.GetOffset(page, limit),
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (a *Access) PackageCountInProject(projectId string) (int64, error) {
	var ret int64
	err := a.query.Get(&ret, "select count(*) from project_packages where project_id = $1", projectId)
	return ret, err
}

func (a *Access) GetPackageVersion(packageVersionId string) (*models.PackageVersion, error) {
	var ret models.PackageVersion
	err := a.query.Get(
		&ret,
		`
		select
			pv.id,
			pv.created_at,
			pv.package_id,
			pv.version,
			pv.release
		from package_versions pv
		where pv.id = $1
		`,
		packageVersionId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetPackageVersionId(packageId string, version string, release string) (string, error) {
	var id string
	err := a.query.Get(&id, "select id from package_versions where package_id = $1 and version = $2 and release = $3", packageId, version, release)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (a *Access) CreatePackageVersion(packageId string, version string, release string) (string, error) {
	var id string
	err := a.query.Get(
		&id,
		`
		insert into package_versions (package_id, version, release)
		values ($1, $2, $3)
		returning id
		`,
		packageId,
		version,
		release,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (a *Access) AttachPackageVersion(projectId string, packageId string, packageVersionId string, active bool) error {
	_, err := a.query.Exec("insert into project_package_versions (project_id, package_id, package_version_id, active) values ($1, $2, $3, $4) on conflict do nothing", projectId, packageId, packageVersionId, active)
	return err
}

func (a *Access) GetProjectPackageVersionFromPackageVersionId(packageVersionId string, projectId string) (string, error) {
	var id string
	err := a.query.Get(
		&id,
		`
		select ppv.id
		from project_package_versions ppv
		inner join package_versions pv on pv.id = ppv.package_version_id
		where
			pv.id = $1
			and ppv.project_id = $2
		`,
		packageVersionId,
		projectId,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

// DeactivateProjectPackageVersionByPackageIdAndProjectId deactivates all package versions for a project
func (a *Access) DeactivateProjectPackageVersionByPackageIdAndProjectId(packageId string, projectId string) error {
	_, err := a.query.Exec("update project_package_versions set active = false where project_id = $1 and package_id = $2", projectId, packageId)
	return err
}

func (a *Access) MakeActiveInRepoForPackageVersion(packageVersionId string, packageId string, projectId string) error {
	tx, err := a.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("update project_package_versions set active_in_repo = false where package_id = $1 and project_id = $2", packageId, projectId)
	if err != nil {
		return err
	}

	_, err = tx.Exec("update project_package_versions set active_in_repo = true where package_version_id = $1 and project_id = $2", packageVersionId, projectId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (a *Access) CreatePackage(name string, packageType peridotpb.PackageType) (*models.Package, error) {
	ret := models.Package{
		Name:        name,
		PackageType: packageType,
	}

	err := a.query.Get(
		&ret,
		`
		insert into packages (name, package_type)
		values ($1, $2)
		returning id, created_at, updated_at
		`,
		name,
		packageType,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) AddPackageToProject(projectId string, packageId string, packageTypeOverride peridotpb.PackageType) error {
	_, err := a.query.Exec(
		`
		insert into project_packages (project_id, package_id, package_type_override)
		values ($1, $2, $3)
		on conflict on constraint project_packages_unique_entry do
			update set package_type_override = $3
		`,
		projectId,
		packageId,
		packageTypeOverride,
	)
	return err
}

func (a *Access) GetPackageID(name string) (string, error) {
	var id string
	err := a.query.Get(&id, "select id from packages where name = $1", name)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (a *Access) SetExtraOptionsForPackage(projectId string, packageName string, withFlags pq.StringArray, withoutFlags pq.StringArray) error {
	if withFlags == nil {
		withFlags = pq.StringArray{}
	}
	if withoutFlags == nil {
		withoutFlags = pq.StringArray{}
	}
	_, err := a.query.Exec(
		`
        insert into extra_package_options (project_id, package_name, with_flags, without_flags)
        values ($1, $2, $3, $4)
        on conflict on constraint extra_package_options_uniq do
            update set with_flags = $3, without_flags = $4, updated_at = now()
        `,
		projectId,
		packageName,
		withFlags,
		withoutFlags,
	)
	return err
}

func (a *Access) GetExtraOptionsForPackage(projectId string, packageName string) (*models.ExtraOptions, error) {
	var ret models.ExtraOptions
	err := a.query.Get(
		&ret,
		"select id, created_at, updated_at, project_id, package_name, with_flags, without_flags, depends_on, enable_module, disable_module from extra_package_options where project_id = $1 and package_name = $2",
		projectId,
		packageName,
	)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (a *Access) SetGroupInstallOptionsForPackage(projectId string, packageName string, dependsOn pq.StringArray, enableModule pq.StringArray, disableModule pq.StringArray) error {
	//NOTE(nhanlon) - 2022-12-19 there is probably a better way to default these?
	if dependsOn == nil {
		dependsOn = pq.StringArray{}
	}
	if enableModule == nil {
		enableModule = pq.StringArray{}
	}
	if disableModule == nil {
		disableModule = pq.StringArray{}
	}

	_, err := a.query.Exec(
		`
        insert into extra_package_options (project_id, package_name, depends_on, enable_module, disable_module)
        values ($1, $2, $3, $4, $5)
        on conflict on constraint extra_package_options_uniq do
            update set depends_on = $3, enable_module = $4, disable_module = $5, updated_at = now()
        `,
		projectId,
		packageName,
		dependsOn,
		enableModule,
		disableModule,
	)
	return err
}

func (a *Access) SetPackageType(projectId string, packageName string, packageType peridotpb.PackageType) error {
	_, err := a.query.Exec(
		`
        update project_packages set package_type_override = $3
        where project_id = $1 and package_id = (select id from packages where name = $2)
        `,
		projectId,
		packageName,
		packageType,
	)
	return err
}
