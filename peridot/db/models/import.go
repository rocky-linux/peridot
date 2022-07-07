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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"time"
)

type Import struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	ScmUrl       string               `json:"scmUrl" db:"scm_url"`
	TaskId       string               `json:"taskId" db:"task_id"`
	ProjectId    string               `json:"projectId" db:"project_id"`
	PackageName  string               `json:"packageName" db:"package_name"`
	TaskStatus   peridotpb.TaskStatus `json:"taskStatus" db:"task_status"`
	TaskResponse types.NullJSONText   `json:"taskResponse" db:"task_response"`

	// Only used for select queries
	Total int64 `json:"total" db:"total"`
}

type Imports []Import

func (i *Import) ToProto() (*peridotpb.Import, error) {
	var ir []*peridotpb.ImportRevision = nil

	if i.TaskResponse.Valid && i.TaskStatus == peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED {
		taskMetadata := peridotpb.ImportPackageTask{}
		anyMetadata := anypb.Any{}
		err := protojson.Unmarshal(i.TaskResponse.JSONText, &anyMetadata)
		if err != nil {
			return nil, err
		}
		err = anyMetadata.UnmarshalTo(&taskMetadata)
		if err != nil {
			return nil, err
		}
		ir = taskMetadata.ImportRevisions
	}

	return &peridotpb.Import{
		Id:        i.ID.String(),
		CreatedAt: timestamppb.New(i.CreatedAt),
		Name:      i.PackageName,
		TaskId:    i.TaskId,
		Status:    i.TaskStatus,
		Revisions: ir,
	}, nil
}

func (i Imports) ToProto() ([]*peridotpb.Import, error) {
	var imports []*peridotpb.Import = nil

	for _, v := range i {
		importProto, err := v.ToProto()
		if err != nil {
			return nil, err
		}
		imports = append(imports, importProto)
	}

	return imports, nil
}

type ImportRevision struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	ImportId         string `json:"importId" db:"import_id"`
	ScmHash          string `json:"scmHash" db:"scm_hash"`
	ScmBranchName    string `json:"scmBranchName" db:"scm_branch_name"`
	ScmUrl           string `json:"scmUrl" db:"scm_url"`
	PackageVersionId string `json:"packageVersionId" db:"package_version_id"`
	Modular          bool   `json:"modular" db:"modular"`
	Active           bool   `json:"active" db:"active"`
	Version          string `json:"version" db:"version"`
	Release          string `json:"release" db:"release"`
}

type ImportRevisions []ImportRevision

func (ir *ImportRevision) ToProto() *peridotpb.ImportRevision {
	return &peridotpb.ImportRevision{
		ScmHash:       ir.ScmHash,
		ScmBranchName: ir.ScmBranchName,
		ScmUrl:        ir.ScmUrl,
		Module:        ir.Modular,
		Vre: &peridotpb.VersionRelease{
			Version: wrapperspb.String(ir.Version),
			Release: wrapperspb.String(ir.Release),
		},
		PackageVersionId: ir.PackageVersionId,
	}
}

type ImportBatch struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	Count     int32 `json:"count" db:"count"`
	Pending   int32 `json:"pending" db:"pending"`
	Running   int32 `json:"running" db:"running"`
	Succeeded int32 `json:"succeeded" db:"succeeded"`
	Failed    int32 `json:"failed" db:"failed"`
	Canceled  int32 `json:"canceled" db:"canceled"`

	// Only used for select queries
	Total int64 `json:"total" db:"total"`
}

func (ib *ImportBatch) ToProto() *peridotpb.ImportBatch {
	return &peridotpb.ImportBatch{
		Id:        ib.ID.String(),
		CreatedAt: timestamppb.New(ib.CreatedAt),
		Count:     ib.Count,
		Pending:   ib.Pending,
		Running:   ib.Running,
		Succeeded: ib.Succeeded,
		Failed:    ib.Failed,
		Canceled:  ib.Canceled,
	}
}

type ImportBatches []ImportBatch

func (ib ImportBatches) ToProto() []*peridotpb.ImportBatch {
	var batches []*peridotpb.ImportBatch

	for _, v := range ib {
		batchProto := v.ToProto()
		batches = append(batches, batchProto)
	}

	return batches
}
