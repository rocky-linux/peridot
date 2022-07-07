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
	peridotpb "peridot.resf.org/peridot/pb"
	"time"
)

type Build struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	PackageId        string               `json:"packageId" db:"package_id"`
	PackageName      string               `json:"packageName" db:"package_name"`
	PackageVersionId string               `json:"packageVersionId" db:"package_version_id"`
	TaskId           string               `json:"taskId" db:"task_id"`
	ProjectId        string               `json:"projectId" db:"project_id"`
	TaskStatus       peridotpb.TaskStatus `json:"taskStatus" db:"task_status"`
	TaskResponse     types.NullJSONText   `json:"taskResponse" db:"task_response"`
	TaskMetadata     types.NullJSONText   `json:"taskMetadata" db:"task_metadata"`

	// Only used for select queries
	Total int64 `json:"total" db:"total"`
}

type Builds []Build

func (b *Build) ToProto() (*peridotpb.Build, error) {
	var ir []*peridotpb.ImportRevision

	if b.TaskResponse.Valid && b.TaskStatus == peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED {
		anyResponse := anypb.Any{}
		err := protojson.Unmarshal(b.TaskResponse.JSONText, &anyResponse)
		if err != nil {
			return nil, err
		}

		if anyResponse.TypeUrl == "type.googleapis.com/resf.peridot.v1.ModuleBuildTask" {
			taskResponse := peridotpb.ModuleBuildTask{}

			err = anyResponse.UnmarshalTo(&taskResponse)
			if err != nil {
				return nil, err
			}

			for _, stream := range taskResponse.Streams {
				ir = append(ir, stream.ImportRevision)
			}
		} else if anyResponse.TypeUrl == "type.googleapis.com/resf.peridot.v1.SubmitBuildTask" {
			taskResponse := peridotpb.SubmitBuildTask{}

			err = anyResponse.UnmarshalTo(&taskResponse)
			if err != nil {
				return nil, err
			}
			ir = append(ir, taskResponse.ImportRevision)
		}
	}

	return &peridotpb.Build{
		Id:              b.ID.String(),
		CreatedAt:       timestamppb.New(b.CreatedAt),
		Name:            b.PackageName,
		ImportRevisions: ir,
		TaskId:          b.TaskId,
		Status:          b.TaskStatus,
	}, nil
}

func (b Builds) ToProto() ([]*peridotpb.Build, error) {
	var builds []*peridotpb.Build
	for _, build := range b {
		protoBuild, err := build.ToProto()
		if err != nil {
			return nil, err
		}
		builds = append(builds, protoBuild)
	}
	return builds, nil
}

type BuildBatch struct {
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

func (ib *BuildBatch) ToProto() *peridotpb.BuildBatch {
	return &peridotpb.BuildBatch{
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

type BuildBatches []BuildBatch

func (ib BuildBatches) ToProto() []*peridotpb.BuildBatch {
	var batches []*peridotpb.BuildBatch

	for _, v := range ib {
		batchProto := v.ToProto()
		batches = append(batches, batchProto)
	}

	return batches
}
