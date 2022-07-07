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
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"time"
)

type Task struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	CreatedAt  time.Time    `json:"createdAt" db:"created_at"`
	FinishedAt sql.NullTime `json:"finishedAt" db:"finished_at"`

	Arch         string               `json:"arch" db:"arch"`
	Type         peridotpb.TaskType   `json:"type" db:"type"`
	Response     types.NullJSONText   `json:"response" db:"response"`
	Metadata     types.NullJSONText   `json:"metadata" db:"metadata"`
	Status       peridotpb.TaskStatus `json:"status" db:"status"`
	ProjectId    sql.NullString       `json:"projectId" db:"project_id"`
	ParentTaskId sql.NullString       `json:"parentTaskId" db:"parent_task_id"`

	SubmitterId          sql.NullString `json:"submitterId" db:"submitter_id"`
	SubmitterDisplayName sql.NullString `json:"submitterDisplayName" db:"submitter_display_name"`
	SubmitterEmail       sql.NullString `json:"submitterEmail" db:"submitter_email"`

	// Only useful for select queries
	Total int64 `json:"total" db:"total"`
}

type Tasks []Task

func (t *Task) ToProto(disableResponse bool) (*peridotpb.Subtask, error) {
	var infoPb *anypb.Any
	var metadataPb *anypb.Any

	if t.Response.Valid && !disableResponse {
		infoPb = &anypb.Any{}
		err := protojson.Unmarshal(t.Response.JSONText, infoPb)
		if err != nil {
			return nil, err
		}
	}
	if t.Metadata.Valid {
		metadataPb = &anypb.Any{}
		err := protojson.Unmarshal(t.Metadata.JSONText, metadataPb)
		if err != nil {
			return nil, err
		}
	}

	var parentTaskId *wrapperspb.StringValue
	if t.ParentTaskId.Valid {
		parentTaskId = wrapperspb.String(t.ParentTaskId.String)
	}

	var finishedAt *timestamppb.Timestamp
	if t.FinishedAt.Valid {
		finishedAt = timestamppb.New(t.FinishedAt.Time)
	}

	return &peridotpb.Subtask{
		Arch:                 t.Arch,
		Type:                 t.Type,
		Response:             infoPb,
		Metadata:             metadataPb,
		Status:               t.Status,
		ParentTaskId:         parentTaskId,
		Id:                   t.ID.String(),
		SubmitterId:          utils.NullStringValueP(t.SubmitterId),
		SubmitterDisplayName: utils.NullStringValueP(t.SubmitterDisplayName),
		SubmitterEmail:       utils.NullStringValueP(t.SubmitterEmail),
		FinishedAt:           finishedAt,
		CreatedAt:            timestamppb.New(t.CreatedAt),
	}, nil
}

func (t Tasks) ToProto(disableResponses bool) (ret []*peridotpb.Subtask, err error) {
	for _, v := range t {
		vp, err := v.ToProto(disableResponses)
		if err != nil {
			return nil, err
		}
		ret = append(ret, vp)
	}

	return ret, nil
}

type TaskArtifact struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	TaskId     string             `json:"taskId" db:"task_id"`
	Name       string             `json:"name" db:"name"`
	HashSha256 string             `json:"hashSha256" db:"hash_sha256"`
	Arch       string             `json:"arch" db:"arch"`
	Metadata   types.NullJSONText `json:"metadata" db:"metadata"`
	Signatures pq.StringArray     `json:"signatures" db:"signatures"`

	// Only used ephemerally - not in DB
	Forced   bool `json:"forced"`
	Multilib bool `json:"multilib"`
}

type TaskArtifacts []TaskArtifact
