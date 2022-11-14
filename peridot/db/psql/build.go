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
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (a *Access) CreateBuild(packageId string, packageVersionId string, taskId string, projectId string) (*models.Build, error) {
	p := models.Build{
		PackageId:        packageId,
		PackageVersionId: packageVersionId,
		TaskId:           taskId,
		ProjectId:        projectId,
	}

	err := a.query.Get(&p, "insert into builds (package_id, package_version_id, task_id, project_id) values ($1, $2, $3, $4) returning id, created_at", packageId, packageVersionId, taskId, projectId)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) GetArtifactsForBuild(buildId string) (ret models.TaskArtifacts, err error) {
	err = a.query.Select(
		&ret,
		`
		select
			distinct(ta.id),
			ta.created_at,
			ta.task_id,
			ta.name,
			ta.hash_sha256,
			ta.arch,
			ta.metadata,
			array_remove(array_agg(tas.gpg_key_id), NULL) as signatures
		from task_artifacts ta
		inner join build_tasks bt on bt.task_id = ta.task_id
		left outer join task_artifact_signatures tas on tas.task_artifact_id = ta.id
		where bt.build_id = $1
		group by ta.id
		order by ta.created_at desc
		`,
		buildId,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetBuildCount() (int64, error) {
	var count int64
	err := a.query.Get(&count, "select count(id) from builds")
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (a *Access) CreateBuildBatch(projectId string) (string, error) {
	var id string
	err := a.query.Get(&id, "insert into build_batches (project_id) values ($1) returning id", projectId)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (a *Access) AttachBuildToBatch(buildId string, batchId string) error {
	_, err := a.query.Exec("insert into build_batch_items (build_batch_id, build_id) values ($1, $2) on conflict do nothing", batchId, buildId)
	if err != nil {
		return err
	}
	return nil
}

func (a *Access) ListBuilds(projectId string, page int32, limit int32) (models.Builds, error) {
	var ret models.Builds
	err := a.query.Select(
		&ret,
		`
		select
			b.id,
			b.created_at,
			b.package_id,
			p.name as package_name,
			b.package_version_id,
			b.task_id,
			b.project_id,
			t.status as task_status,
			t.response as task_response,
			t.metadata as task_metadata,
			count(b.*) over() as total
		from builds b
		inner join tasks t on t.id = b.task_id
		inner join packages p on p.id = b.package_id
		where b.project_id = $1
		order by b.created_at desc
		limit $2 offset $3
		`,
		projectId,
		limit,
		utils.GetOffset(page, limit),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) BuildCountInProject(projectId string) (int64, error) {
	var count int64
	err := a.query.Get(&count, "select count(*) from imports where project_id = $1", projectId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *Access) GetBuild(projectId string, buildId string) (*models.Build, error) {
	var ret models.Build
	err := a.query.Get(
		&ret,
		`
		select
			b.id,
			b.created_at,
			b.package_id,
			p.name as package_name,
			b.package_version_id,
			b.task_id,
			b.project_id,
			t.status as task_status,
			t.response as task_response,
			t.metadata as task_metadata
		from builds b
		inner join tasks t on t.id = b.task_id
		inner join packages p on p.id = b.package_id
		where
			b.project_id = $1
			and b.id = $2
		order by b.created_at desc
		`,
		projectId,
		buildId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetBuildByTaskIdAndPackageId(taskId string, packageId string) (*models.Build, error) {
	var ret models.Build
	err := a.query.Get(
		&ret,
		`
		select
			b.id,
			b.created_at,
			b.package_id,
			p.name as package_name,
			b.package_version_id,
			b.task_id,
			b.project_id,
			t.status as task_status,
			t.response as task_response,
			t.metadata as task_metadata
		from builds b
		inner join tasks t on t.id = b.task_id
		inner join packages p on p.id = b.package_id
		where
			b.task_id = $1
			and b.package_id = $2
		order by b.created_at desc
		`,
		taskId,
		packageId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetBuildBatch(projectId string, batchId string, batchFilter *peridotpb.BatchFilter, page int32, limit int32) (models.Builds, error) {
	if batchFilter == nil {
		batchFilter = &peridotpb.BatchFilter{}
	}

	var ret models.Builds
	err := a.query.Select(
		&ret,
		`
		select
			distinct(b.id),
			b.created_at,
			b.package_id,
			p.name as package_name,
			b.package_version_id,
			b.task_id,
			b.project_id,
			t.status as task_status,
			t.response as task_response,
			t.metadata as task_metadata,
			count(b.*) over() as total
		from build_batches buildb
		inner join build_batch_items buildbi on buildbi.build_batch_id = buildb.id
		inner join builds b on b.id = buildbi.build_id
		inner join tasks t on t.id = b.task_id
		inner join packages p on p.id = b.package_id
		where
			buildb.project_id = $1
			and buildb.id = $2
			and ($3 = 0 or t.status = $3)
		order by b.created_at desc
		limit $4 offset $5
		`,
		projectId,
		batchId,
		batchFilter.Status,
		limit,
		utils.GetOffset(page, limit),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) ListBuildBatches(projectId string, batchId *string, page int32, limit int32) (models.BuildBatches, error) {
	var ret models.BuildBatches
	err := a.query.Select(
		&ret,
		`
		select
			distinct(buildb.id),
			buildb.created_at,
			count(buildbi.*) as count,
			count(t.*) filter (where t.status = 1) as pending,
			count(t.*) filter (where t.status = 2) as running,
			count(t.*) filter (where t.status = 3) as succeeded,
			count(t.*) filter (where t.status = 4) as failed,
			count(t.*) filter (where t.status = 5) as canceled,
			count(buildb.*) over() as total
		from build_batches buildb
		inner join build_batch_items buildbi on buildbi.build_batch_id = buildb.id
		inner join builds b on b.id = buildbi.build_id
		inner join tasks t on t.id = b.task_id
		where
			buildb.project_id = $1
			and ($2 :: uuid is null or buildb.id = $2 :: uuid)
		group by buildb.id
		order by buildb.created_at desc
		limit $3 offset $4
		`,
		projectId,
		batchId,
		utils.UnlimitedLimit(limit),
		utils.GetOffset(page, limit),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) BuildBatchCountInProject(projectId string) (int64, error) {
	var count int64
	err := a.query.Get(&count, "select count(*) from build_batches where project_id = $1", projectId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *Access) BuildsInBatchCount(projectId string, batchId string) (int64, error) {
	var count int64
	err := a.query.Get(
		&count,
		`
		select
			count(b.*)
		from build_batches buildb
		inner join build_batch_items buildbi on buildbi.build_batch_id = buildb.id
		inner join builds b on b.id = buildbi.build_id
		where
			buildb.project_id = $1
			and buildb.id = $2
		`,
		projectId,
		batchId,
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *Access) LockNVRA(nvra string) error {
	_, err := a.query.Exec("insert into nvrs (name) values ($1)", nvra)
	return err
}

func (a *Access) UnlockNVRA(nvra string) error {
	_, err := a.query.Exec("delete from nvrs where name = $1", nvra)
	return err
}

func (a *Access) NVRAExists(nvra string) (bool, error) {
	var count int
	err := a.query.Get(&count, "select count(*) from nvrs where name = $1", nvra)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (a *Access) GetBuildByPackageNameAndVersionAndRelease(name string, version string, release string, projectId string) (*models.Build, error) {
	var ret models.Build
	err := a.query.Get(
		&ret,
		`
		select
			b.id,
			b.created_at,
			b.package_id,
			p.name as package_name,
			b.package_version_id,
			b.task_id,
			b.project_id,
			t.status as task_status,
			t.response as task_response,
			t.metadata as task_metadata
		from builds b
		inner join tasks t on t.id = b.task_id
		inner join packages p on p.id = b.package_id
        inner join package_versions pv on pv.id = b.package_version_id
		where
			b.project_id = $1
            and p.name = $2
			and pv.version = $3
            and pv.release = $4
            and t.status = 3
		order by b.created_at desc
        limit 1
		`,
		projectId,
		name,
		version,
		release,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetLatestBuildIdsByPackageName(name string, projectId string) ([]string, error) {
	var ret []string
	err := a.query.Select(
		&ret,
		`
		select
			b.id
		from builds b
		inner join tasks t on t.id = b.task_id
        inner join packages p on p.id = b.package_id
        inner join project_package_versions ppv on ppv.package_version_id = b.package_version_id
		where
            p.name = $2
            and t.status = 3
            and ppv.active_in_repo = true
            and ppv.project_id = b.project_id
            and b.project_id = $1
		order by b.created_at asc
		`,
		projectId,
		name,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetLatestBuildsByPackageNameAndBranchName(name string, branchName string, projectId string) ([]string, error) {
	var ret []string
	err := a.query.Select(
		&ret,
		`
        select
            b.id
        from builds b
        inner join tasks t on t.id = b.task_id
        inner join packages p on p.id = b.package_id
        inner join project_package_versions ppv on ppv.package_version_id = b.package_version_id
        inner join import_revisions ir on ir.package_version_id = b.package_version_id
        where
            b.project_id = $3
            and p.name = $1
            and ppv.active_in_repo = true
            and ppv.project_id = b.project_id
            and ir.scm_branch_name = $2
            and t.status = 3
        order by b.created_at asc
        `,
		name,
		branchName,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetActiveBuildIdsByTaskArtifactGlob(taskArtifactGlob string, projectId string) ([]string, error) {
	var ret []string
	err := a.query.Select(
		&ret,
		`
		select
			b.id
		from builds b
		inner join tasks t on t.id = b.task_id
        inner join project_package_versions ppv on ppv.package_version_id = b.package_version_id
		where
			task_id in (select distinct(parent_task_id) from tasks where id in (select task_id from task_artifacts where name like $2))
            and t.status = 3
            and ppv.active_in_repo = true
            and ppv.project_id = b.project_id
            and b.project_id = $1
		order by b.created_at asc
		`,
		projectId,
		taskArtifactGlob,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetAllBuildIdsByPackageName(name string, projectId string) ([]string, error) {
	var ret []string
	err := a.query.Select(
		&ret,
		`
		select
			b.id
		from builds b
		inner join tasks t on t.id = b.task_id
        inner join packages p on p.id = b.package_id
        inner join project_package_versions ppv on ppv.package_version_id = b.package_version_id
		where
            p.name = $2
            and t.status = 3
            and ppv.project_id = b.project_id
            and b.project_id = $1
		order by b.created_at asc
		`,
		projectId,
		name,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
