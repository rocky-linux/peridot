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

func (a *Access) CreateImport(scmUrl string, taskId string, packageId string, projectId string) (*models.Import, error) {
	p := models.Import{
		ScmUrl:    scmUrl,
		TaskId:    taskId,
		ProjectId: projectId,
	}

	err := a.query.Get(&p, "insert into imports (scm_url, task_id, package_id, project_id) values ($1, $2, $3, $4) returning id, created_at", scmUrl, taskId, packageId, projectId)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) CreateImportRevision(importId string, scmHash string, scmBranchName string, scmUrl string, packageVersionId string, modular bool) (*models.ImportRevision, error) {
	p := models.ImportRevision{
		ImportId:         importId,
		ScmHash:          scmHash,
		ScmBranchName:    scmBranchName,
		ScmUrl:           scmUrl,
		PackageVersionId: packageVersionId,
		Modular:          modular,
		Active:           true,
	}

	err := a.query.Get(&p, "insert into import_revisions (import_id, scm_hash, scm_branch_name, scm_url, package_version_id, modular) values ($1, $2, $3, $4, $5, $6) returning id, created_at", importId, scmHash, scmBranchName, scmUrl, packageVersionId, modular)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) GetLatestImportRevisionsForPackageInProject(packageName string, projectId string) (models.ImportRevisions, error) {
	var p models.ImportRevisions

	err := a.query.Select(
		&p,
		`
		select
			distinct(ir.id),
			ir.created_at,
			ir.import_id,
			ir.scm_hash,
			ir.scm_branch_name,
			ir.scm_url,
			ir.package_version_id,
			ir.modular,
			ir.active,
	        pv.version,
			pv.release
		from
			import_revisions ir
		inner join project_package_versions ppv on ppv.package_version_id = ir.package_version_id
		inner join package_versions pv on pv.id = ir.package_version_id
		inner join packages p on p.id = ppv.package_id
		inner join projects proj on proj.id = ppv.project_id
		where
			p.name = $1
			and ppv.project_id = $2
			and ppv.active = true
			and ir.active = true
		order by ir.created_at desc
		`,
		packageName,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (a *Access) GetImportRevisionByScmHash(scmHash string) (*models.ImportRevision, error) {
	var p models.ImportRevision

	err := a.query.Get(
		&p,
		`
		select
			distinct(ir.id),
			ir.created_at,
			ir.import_id,
			ir.scm_hash,
			ir.scm_branch_name,
			ir.scm_url,
			ir.package_version_id,
			ir.modular,
			ir.active,
	        pv.version,
			pv.release
		from import_revisions ir
		inner join package_versions pv on pv.id = ir.package_version_id
		where ir.scm_hash = $1
	    `,
		scmHash,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) DeactivateImportRevisionsByPackageVersionId(packageVersionId string) error {
	_, err := a.query.Exec(
		`
        update import_revisions
        set active = false
        where package_version_id = $1
        `,
		packageVersionId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (a *Access) CreateImportBatch(projectId string) (string, error) {
	var id string
	err := a.query.Get(&id, "insert into import_batches (project_id) values ($1) returning id", projectId)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (a *Access) AttachImportToBatch(importId string, batchId string) error {
	_, err := a.query.Exec("insert into import_batch_items (import_batch_id, import_id) values ($1, $2) on conflict do nothing", batchId, importId)
	if err != nil {
		return err
	}
	return nil
}

func (a *Access) ListImports(projectId string, page int32, limit int32) (models.Imports, error) {
	var ret models.Imports
	err := a.query.Select(
		&ret,
		`
		select
			i.id,
			i.created_at,
			i.scm_url,
			i.task_id,
			i.project_id,
			p.name as package_name,
			t.status as task_status,
			t.response as task_response,
			count(i.*) over() as total
		from imports i
		inner join tasks t on t.id = i.task_id
		inner join packages p on p.id = i.package_id
		where i.project_id = $1
		order by i.created_at desc
		limit $2 offset $3
		`,
		projectId,
		utils.UnlimitedLimit(limit),
		utils.GetOffset(page, limit),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) ImportCountInProject(projectId string) (int64, error) {
	var count int64
	err := a.query.Get(&count, "select count(*) from imports where project_id = $1", projectId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *Access) GetImport(projectId string, importId string) (*models.Import, error) {
	var ret models.Import
	err := a.query.Get(
		&ret,
		`
		select
			i.id,
			i.created_at,
			i.scm_url,
			i.task_id,
			i.project_id,
			p.name as package_name,
			t.status as task_status,
			t.response as task_response,
			count(i.*) over() as total
		from imports i
		inner join tasks t on t.id = i.task_id
		inner join packages p on p.id = i.package_id
		where
			i.project_id = $1
			and i.id = $2
		order by i.created_at desc
		`,
		projectId,
		importId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetImportBatch(projectId string, batchId string, batchFilter *peridotpb.BatchFilter, page int32, limit int32) (models.Imports, error) {
	if batchFilter == nil {
		batchFilter = &peridotpb.BatchFilter{}
	}

	var ret models.Imports
	err := a.query.Select(
		&ret,
		`
		select
			i.id,
			i.created_at,
			i.scm_url,
			i.task_id,
			i.project_id,
			p.name as package_name,
			t.status as task_status,
			t.response as task_response,
			count(i.*) over() as total
		from import_batches ib
		inner join import_batch_items ibi on ibi.import_batch_id = ib.id
		inner join imports i on i.id = ibi.import_id
		inner join tasks t on t.id = i.task_id
		inner join packages p on p.id = i.package_id
		where
			ib.project_id = $1
			and ib.id = $2
			and ($3 = 0 or t.status = $3)
		order by i.created_at desc
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

func (a *Access) ListImportBatches(projectId string, batchId *string, page int32, limit int32) (models.ImportBatches, error) {
	var ret models.ImportBatches
	err := a.query.Select(
		&ret,
		`
		select
			distinct(ib.id),
			ib.created_at,
			count(ibi.*) as count,
			count(t.*) filter (where t.status = 1) as pending,
			count(t.*) filter (where t.status = 2) as running,
			count(t.*) filter (where t.status = 3) as succeeded,
			count(t.*) filter (where t.status = 4) as failed,
			count(t.*) filter (where t.status = 5) as canceled,
			count(ib.*) over() as total
		from import_batches ib
		inner join import_batch_items ibi on ibi.import_batch_id = ib.id
		inner join imports i on i.id = ibi.import_id
		inner join tasks t on t.id = i.task_id
		where
			ib.project_id = $1
			and ($2 :: uuid is null or ib.id = $2 :: uuid)
		group by ib.id
		order by ib.created_at desc
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

func (a *Access) ImportBatchCountInProject(projectId string) (int64, error) {
	var count int64
	err := a.query.Get(&count, "select count(*) from import_batches where project_id = $1", projectId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *Access) ImportsInBatchCount(projectId string, batchId string) (int64, error) {
	var count int64
	err := a.query.Get(
		&count,
		`
		select
			count(i.*)
		from import_batches ib
		inner join import_batch_items ibi on ibi.import_batch_id = ib.id
		inner join imports i on i.id = ibi.import_id
		where
			ib.project_id = $1
			and ib.id = $2
		`,
		projectId,
		batchId,
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}
