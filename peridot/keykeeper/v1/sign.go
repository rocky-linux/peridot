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

package keykeeperv1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	peridotworkflow "peridot.resf.org/peridot/builder/v1/workflow"
	"peridot.resf.org/peridot/db/models"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

var (
	ErrUnsupportedExtension = errors.New("unsupported extension")
)

func (s *Server) SignArtifactsWorkflow(ctx workflow.Context, artifacts models.TaskArtifacts, buildId string, task *models.Task, keyName string) (*keykeeperpb.SignArtifactsTask, error) {
	taskResponse := &keykeeperpb.SignArtifactsTask{
		SignedArtifacts: []*keykeeperpb.SignedArtifact{},
	}

	err := s.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	defer func() {
		if taskResponse != nil {
			taskResponseAny, err := anypb.New(taskResponse)
			if err != nil {
				s.log.Errorf("could not create anypb for task: %v", err)
				task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED
			} else {
				err = s.db.SetTaskResponse(task.ID.String(), taskResponseAny)
				if err != nil {
					s.log.Errorf("could not set task info: %v", err)
					task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED
				}
			}
		}

		err := s.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			s.log.Errorf("could not set task status: %v", err)
		}
	}()

	if artifacts == nil {
		var err error
		artifacts, err = s.db.GetArtifactsForBuild(buildId)
		if err != nil {
			s.log.Errorf("failed to get artifacts for build %s: %v", buildId, err)
			return nil, status.Error(codes.Internal, "failed to get artifacts")
		}
		if len(artifacts) == 0 {
			return taskResponse, nil
		}
	}

	var futures []peridotworkflow.FutureContext
	for _, artifact := range artifacts {
		signArtifactCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: 10 * time.Hour,
			StartToCloseTimeout:    24 * time.Hour,
			HeartbeatTimeout:       time.Minute,
			TaskQueue:              TaskQueue,
		})
		futures = append(futures, peridotworkflow.FutureContext{
			Ctx:       signArtifactCtx,
			Future:    workflow.ExecuteActivity(signArtifactCtx, s.SignArtifactActivity, artifact.ID.String(), keyName),
			TaskQueue: TaskQueue,
		})
	}

	for _, future := range futures {
		signedArtifact := &keykeeperpb.SignedArtifact{}
		err := future.Future.Get(future.Ctx, signedArtifact)
		if err != nil && !strings.Contains(err.Error(), "unsupported extension") {
			s.log.Errorf("could not get sign artifact: %v", err)
			return nil, err
		}
		if signedArtifact != nil {
			taskResponse.SignedArtifacts = append(taskResponse.SignedArtifacts, signedArtifact)
		}
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return taskResponse, nil
}

func (s *Server) SignArtifactActivity(ctx context.Context, artifactId string, keyName string) (*keykeeperpb.SignedArtifact, error) {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(4 * time.Second)
		}
	}()

	artifact, err := s.db.GetTaskArtifactById(artifactId)
	if err != nil {
		s.log.Errorf("could not get artifact: %v", err)
		return nil, status.Errorf(codes.Internal, "could not get artifact")
	}

	key, err := s.EnsureGPGKey(keyName)
	if err != nil {
		s.log.Errorf("failed to load key %s: %v", keyName, err)
		return nil, status.Error(codes.Internal, "failed to load key")
	}

	newObjectKey := fmt.Sprintf("%s/%s/%s", filepath.Dir(artifact.Name), key.gpgId, filepath.Base(artifact.Name))
	existingHash, err := s.db.GetTaskArtifactSignatureHash(artifact.Name, key.keyUuid.String())
	if err != nil && err != sql.ErrNoRows {
		s.log.Errorf("failed to get existing hash: %v", err)
		return nil, status.Error(codes.Internal, "failed to get existing hash")
	}
	if err == nil && existingHash != "" {
		return &keykeeperpb.SignedArtifact{
			Path:       newObjectKey,
			HashSha256: existingHash,
		}, nil
	}

	ranUuid := uuid.New()
	localPath := fmt.Sprintf("/keykeeper/artifacts/%s-%s", ranUuid.String(), filepath.Base(artifact.Name))
	err = s.storage.DownloadObject(artifact.Name, localPath)
	if err != nil {
		s.log.Errorf("failed to download artifact %s: %v", artifact.Name, err)
		return nil, fmt.Errorf("failed to download artifact %s: %v", artifact.Name, err)
	}
	defer func() {
		_ = os.Remove(localPath)
	}()
	ext := filepath.Ext(artifact.Name)

	switch ext {
	case ".rpm":
		rpmSign := func() (*keykeeperpb.SignedArtifact, error) {
			var outBuf bytes.Buffer
			opts := []string{
				"--define", "_gpg_name " + keyName,
				"--define", "_peridot_keykeeper_key " + key.keyUuid.String(),
				"--addsign", localPath,
			}
			cmd := gpgCmdEnv(exec.Command("rpm", opts...))
			cmd.Stdout = &outBuf
			cmd.Stderr = &outBuf
			err := cmd.Run()
			if err != nil {
				s.log.Errorf("failed to sign artifact %s: %v", artifact.Name, err)
				statusErr := status.New(codes.Internal, "failed to sign artifact")
				statusErr, err2 := statusErr.WithDetails(&errdetails.ErrorInfo{
					Reason: "rpmsign-failed",
					Domain: "keykeeper.peridot.resf.org",
					Metadata: map[string]string{
						"logs": outBuf.String(),
						"err":  err.Error(),
					},
				})
				if err2 != nil {
					s.log.Errorf("failed to add error details to status: %v", err2)
				}
				return nil, statusErr.Err()
			}
			_, err = s.storage.PutObject(newObjectKey, localPath)
			if err != nil {
				s.log.Errorf("failed to upload artifact %s: %v", newObjectKey, err)
				return nil, fmt.Errorf("failed to upload artifact %s: %v", newObjectKey, err)
			}

			f, err := os.Open(localPath)
			if err != nil {
				return nil, err
			}

			hasher := sha256.New()
			_, err = io.Copy(hasher, f)
			if err != nil {
				return nil, err
			}
			hash := hex.EncodeToString(hasher.Sum(nil))

			err = s.db.CreateTaskArtifactSignature(artifact.ID.String(), key.keyUuid.String(), hash)
			if err != nil {
				s.log.Errorf("failed to create task artifact signature: %v", err)
				return nil, fmt.Errorf("failed to create task artifact signature: %v", err)
			}

			return &keykeeperpb.SignedArtifact{
				Path:       newObjectKey,
				HashSha256: hash,
			}, nil
		}
		verifySig := func() error {
			opts := []string{
				"--define", "_gpg_name " + keyName,
				"--define", "_peridot_keykeeper_key " + key.keyUuid.String(),
				"--checksig", localPath,
			}
			cmd := gpgCmdEnv(exec.Command("rpm", opts...))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				s.log.Errorf("failed to verify artifact %s: %v", artifact.Name, err)
				return fmt.Errorf("failed to verify artifact %s: %v", artifact.Name, err)
			}
			return nil
		}
		var tries int
		for {
			res, _ := rpmSign()
			err := verifySig()
			if err == nil {
				return res, nil
			}
			if err != nil && tries > 3 {
				return nil, err
			}
			tries++
		}
	default:
		s.log.Infof("skipping artifact %s, extension %s not supported", artifact.Name, ext)
		return nil, ErrUnsupportedExtension
	}
}

// SignArtifacts signs artifacts belonging to a build with the given key.
// The artifacts are loaded from the shared Peridot bucket and is
// then uploaded back.
// Since an artifact can have multiple signers, the resulting
// artifacts are uploaded to `{taskId}/{keyId}/{artifactName}`.
// Essentially we only add the keyId to the artifact key.
// Each artifact should in theory only be signed once,
// but signing it multiple times does not cause any harm.
// We currently fetch the artifact from the shared bucket,
// then sign it within the current container, then upload it back.
// todo(mustafa): Look into a way to avoid fetching the artifact from the shared bucket.
// todo(mustafa): We should still probably avoid signing the same
// todo(mustafa): artifact (with a key it's already signed with)
func (s *Server) SignArtifacts(_ context.Context, req *keykeeperpb.SignArtifactsRequest) (*peridotpb.AsyncTask, error) {
	build, err := s.db.GetTaskByBuildId(req.BuildId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "build not found")
		}
		s.log.Errorf("failed to get build %s: %v", req.BuildId, err)
		return nil, status.Error(codes.Internal, "failed to get build")
	}
	artifacts, err := s.db.GetArtifactsForBuild(req.BuildId)
	if err != nil {
		s.log.Errorf("failed to get artifacts for build %s: %v", req.BuildId, err)
		return nil, status.Error(codes.Internal, "failed to get artifacts")
	}
	if len(artifacts) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no artifacts to sign")
	}

	rollback := true
	beginTx, err := s.db.Begin()
	if err != nil {
		s.log.Error(err)
		return nil, utils.InternalError
	}
	defer func() {
		if rollback {
			_ = beginTx.Rollback()
		}
	}()
	tx := s.db.UseTransaction(beginTx)
	buildTaskId := build.ID.String()

	task, err := tx.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_KEYKEEPER_SIGN_ARTIFACT, &build.ProjectId.String, &buildTaskId)
	if err != nil {
		s.log.Errorf("could not create build task in SubmitBuild: %v", err)
		return nil, status.Error(codes.InvalidArgument, "could not create import task")
	}

	taskProto, err := task.ToProto(false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not marshal task: %v", err)
	}

	rollback = false
	err = beginTx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, "could not save, try again")
	}

	_, err = s.temporal.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        task.ID.String(),
			TaskQueue: TaskQueue,
		},
		s.SignArtifactsWorkflow,
		nil,
		req.BuildId,
		task,
		req.KeyName,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not start workflow: %v", err)
	}

	return &peridotpb.AsyncTask{
		TaskId:   task.ID.String(),
		Subtasks: []*peridotpb.Subtask{taskProto},
		Done:     false,
	}, nil
}

// SignText signs given text with the given key.
// This method only returns the signature part of the gpg clearsign
func (s *Server) SignText(_ context.Context, req *keykeeperpb.SignTextRequest) (*keykeeperpb.SignTextResponse, error) {
	key, err := s.EnsureGPGKey(req.KeyName)
	if err != nil {
		s.log.Errorf("failed to load key %s: %v", req.KeyName, err)
		return nil, status.Error(codes.Internal, "failed to load key")
	}

	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		s.log.Errorf("failed to create temp file: %v", err)
		return nil, status.Error(codes.Internal, "failed to create temp file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	err = ioutil.WriteFile(tmpFile.Name(), []byte(req.Text), 0644)
	if err != nil {
		s.log.Errorf("failed to write to temp file: %v", err)
		return nil, status.Error(codes.Internal, "failed to write to temp file")
	}

	cmdArgs := []string{
		"--batch",
		"--no-verbose",
		"--armor",
		"--pinentry-mode=loopback",
		"--no-secmem-warning",
		"--detach-sign",
		"--passphrase",
		key.keyUuid.String(),
		"-u",
		req.KeyName,
		"-o",
		tmpFile.Name() + ".asc",
		tmpFile.Name(),
	}
	cmd := gpgCmdEnv(exec.Command("gpg", cmdArgs...))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		s.log.Errorf("failed to sign text: %v", err)
		return nil, status.Error(codes.Internal, "failed to sign text")
	}
	defer os.Remove(tmpFile.Name() + ".asc")
	signedText, err := ioutil.ReadFile(tmpFile.Name() + ".asc")
	if err != nil {
		s.log.Errorf("failed to read signed text: %v", err)
		return nil, status.Error(codes.Internal, "failed to read signed text")
	}

	return &keykeeperpb.SignTextResponse{
		Signature: string(signedText),
	}, nil
}
