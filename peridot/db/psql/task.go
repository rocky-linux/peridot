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
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (a *Access) CreateTask(user *utils.ContextUser, arch string, taskType peridotpb.TaskType, projectId *string, parentTaskId *string) (*models.Task, error) {
	task := models.Task{
		Arch:     arch,
		Type:     taskType,
		Response: types.NullJSONText{},
		Status:   peridotpb.TaskStatus_TASK_STATUS_PENDING,
	}
	if projectId != nil {
		task.ProjectId = sql.NullString{
			String: *projectId,
			Valid:  true,
		}
	}
	if parentTaskId != nil {
		task.ParentTaskId = sql.NullString{
			String: *parentTaskId,
			Valid:  true,
		}
	}

	var userId *string
	var userName *string
	var userEmail *string
	if user != nil {
		userId = &user.ID
		userName = &user.Name
		userEmail = &user.Email
	}

	err := a.query.Get(
		&task,
		"insert into tasks (arch, type, project_id, parent_task_id, submitter_id, submitter_display_name, submitter_email) values ($1, $2, $3, $4, $5, $6, $7) returning id, created_at",
		arch,
		taskType,
		projectId,
		parentTaskId,
		userId,
		userName,
		userEmail,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (a *Access) SetTaskStatus(id string, status peridotpb.TaskStatus) error {
	extraSql := ""
	switch status {
	case peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED,
		peridotpb.TaskStatus_TASK_STATUS_FAILED,
		peridotpb.TaskStatus_TASK_STATUS_CANCELED:
		extraSql = ", finished_at = now()"
		break
	}

	_, err := a.query.Exec(fmt.Sprintf("update tasks set status = $1%s where id = $2", extraSql), status, id)
	return err
}

func (a *Access) SetTaskResponse(id string, response *anypb.Any) error {
	bts, err := protojson.Marshal(response)
	if err != nil {
		return err
	}

	_, err = a.query.Exec("update tasks set response = $1::jsonb where id = $2", bts, id)
	return err
}

func (a *Access) SetTaskMetadata(id string, metadata *anypb.Any) error {
	bts, err := protojson.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = a.query.Exec("update tasks set metadata = $1::jsonb where id = $2", bts, id)
	return err
}

func (a *Access) GetTask(id string, projectId *string) (ret models.Tasks, err error) {
	err = a.query.Select(
		&ret,
		`
		with recursive task_query as (
			select * from tasks
			where
				id = $1
				and ($2 :: uuid is null or project_id = $2 :: uuid)
			union all
			select t.* from tasks t
			join task_query tq on tq.id = t.parent_task_id
		)
		select
			id,
			created_at,
			finished_at,
			arch,
			type,
			response,
			metadata,
			status,
			project_id,
			parent_task_id,
			submitter_id,
			submitter_display_name,
			submitter_email
		from task_query
		order by created_at asc
		`,
		id,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) ListTasks(projectId *string, page int32, limit int32) (ret models.Tasks, err error) {
	err = a.query.Select(
		&ret,
		`
		select
			id,
			created_at,
			finished_at,
			arch,
			type,
			response,
			metadata,
			status,
			project_id,
			parent_task_id,
			submitter_id,
			submitter_display_name,
			submitter_email,
			count(*) over() as total
		from tasks
		where
			parent_task_id is null
			and ($1 :: uuid is null or project_id = $1 :: uuid)
		order by created_at desc
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

func (a *Access) AttachTaskToBuild(buildId string, taskId string) error {
	_, err := a.query.Exec("insert into build_tasks (build_id, task_id) values ($1, $2) on conflict do nothing", buildId, taskId)
	return err
}

func (a *Access) AttachArtifactToTask(objectName string, hashSha256 string, arch string, metadata *anypb.Any, taskId string) error {
	var bts []byte = nil
	var err error

	if metadata != nil {
		bts, err = protojson.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	_, err = a.query.Exec("insert into task_artifacts (task_id, name, hash_sha256, arch, metadata) values ($1, $2, $3, $4, $5) on conflict do nothing", taskId, objectName, hashSha256, arch, bts)
	return err
}

// GetTaskByBuildId returns the task of a build (only parent task)
func (a *Access) GetTaskByBuildId(buildId string) (*models.Task, error) {
	var ret models.Task
	err := a.query.Get(
		&ret,
		`
		select
			id,
			created_at,
			finished_at,
			arch,
			type,
			response,
			metadata,
			status,
			project_id,
			parent_task_id,
			submitter_id,
			submitter_display_name,
			submitter_email
		from tasks
		where id = (select task_id from builds where id = $1)
		`,
		buildId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) CreateTaskArtifactSignature(taskArtifactId string, keyId string, hashSha256 string) error {
	_, err := a.query.Exec("insert into task_artifact_signatures (task_artifact_id, gpg_key_id, hash_sha256) values ($1, $2, $3) on conflict (task_artifact_id, gpg_key_id) do update set hash_sha256 = $3", taskArtifactId, keyId, hashSha256)
	return err
}

func (a *Access) GetTaskArtifactSignatureHash(name string, gpgKeyId string) (string, error) {
	var ret string
	err := a.query.Get(
		&ret,
		`
        select tas.hash_sha256 from task_artifact_signatures tas
        inner join task_artifacts ta on ta.id = tas.task_artifact_id
        where
            ta.name = $1
            and tas.gpg_key_id = $2
        `,
		name,
		gpgKeyId,
	)
	if err != nil {
		return "", err
	}

	return ret, nil
}

func (a *Access) TaskCountInProject(projectId string) (int64, error) {
	var count int64
	err := a.query.Get(&count, "select count(*) from tasks where project_id = $1", projectId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *Access) GetTaskArtifactById(taskArtifactId string) (*models.TaskArtifact, error) {
	var ret models.TaskArtifact
	err := a.query.Get(
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
		left outer join task_artifact_signatures tas on tas.task_artifact_id = ta.id
		where ta.id = $1
		group by ta.id
		order by ta.created_at desc
		`,
		taskArtifactId,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetTaskStatus(taskId string) (peridotpb.TaskStatus, error) {
	var status peridotpb.TaskStatus
	err := a.query.Get(&status, "select status from tasks where id = $1", taskId)
	if err != nil {
		return peridotpb.TaskStatus_TASK_STATUS_UNSPECIFIED, err
	}
	return status, nil
}
