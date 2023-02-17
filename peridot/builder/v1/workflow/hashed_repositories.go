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

package workflow

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"io/ioutil"
	"os"
	"path/filepath"
	"peridot.resf.org/peridot/db/models"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/yummeta"
	"strings"
	"time"
)

func (c *Controller) CreateHashedRepositoriesWorkflow(ctx workflow.Context, req *peridotpb.CreateHashedRepositoriesRequest, task *models.Task) (*peridotpb.CreateHashedRepositoriesTask, error) {
	ret := peridotpb.CreateHashedRepositoriesTask{}

	deferTask, errorDetails, err := c.commonCreateTask(task, &ret)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	// Let's provision an ephemeral worker
	taskQueue, cleanupWorker, err := c.provisionWorker(ctx, &ProvisionWorkerRequest{
		TaskId:       task.ID.String(),
		ParentTaskId: task.ParentTaskId,
		Purpose:      "sync",
		Arch:         "noarch",
		ProjectId:    req.ProjectId.Value,
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	defer cleanupWorker()

	createRepositoriesCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 25 * time.Minute,
		StartToCloseTimeout:    15 * time.Minute,
		TaskQueue:              taskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err = workflow.ExecuteActivity(createRepositoriesCtx, c.CreateHashedRepositoriesActivity, req).Get(ctx, &ret)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &ret, nil
}

func (c *Controller) CreateHashedRepositoriesActivity(req *peridotpb.CreateHashedRepositoriesRequest) (*peridotpb.CreateHashedRepositoriesTask, error) {
	ret := peridotpb.CreateHashedRepositoriesTask{}

	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: req.ProjectId,
	})
	if err != nil {
		return nil, err
	}
	if len(projects) != 1 {
		return nil, fmt.Errorf("project not found")
	}
	project := projects[0]

	beginTx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	tx := c.db.UseTransaction(beginTx)

	key, err := tx.GetDefaultKeyForProject(project.ID.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("could not get default key for project %s: %v", project.ID.String(), err)
	}

	for _, repoName := range req.Repositories {
		arches := append(project.Archs, "src")
		for _, arch := range project.Archs {
			arches = append(arches, arch+"-debug")
		}
		for _, arch := range arches {
			tmpDir, err := os.MkdirTemp("", "")
			if err != nil {
				return nil, fmt.Errorf("could not create temporary directory: %v", err)
			}

			//goland:noinspection GoDeferInLoop
			defer func(path string) {
				err := os.RemoveAll(path)
				if err != nil {
					c.log.Errorf("could not remove temporary directory: %v", err)
				}
			}(tmpDir)

			repodataDir := filepath.Join(tmpDir, "repodata")
			err = os.Mkdir(repodataDir, 0755)
			if err != nil {
				return nil, fmt.Errorf("could not create temporary repodata directory: %v", err)
			}

			_, err = c.db.GetRepository(nil, &repoName, &req.ProjectId.Value)
			if err != nil {
				return nil, fmt.Errorf("get repository: %w", err)
			}

			revision, err := c.db.GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(project.ID.String(), repoName, arch)
			if err != nil {
				if err != sql.ErrNoRows {
					return nil, fmt.Errorf("could not get latest active revision: %v", err)
				}
				continue
			}

			filelistsGz, err := base64.StdEncoding.DecodeString(revision.FilelistsXml)
			if err != nil {
				return nil, fmt.Errorf("decode filelists xml: %w", err)
			}
			otherGz, err := base64.StdEncoding.DecodeString(revision.OtherXml)
			if err != nil {
				return nil, fmt.Errorf("decode other xml: %w", err)
			}
			groupsGz, err := base64.StdEncoding.DecodeString(revision.GroupsXml)
			if err != nil {
				return nil, fmt.Errorf("decode groups xml: %w", err)
			}
			var groupsXml []byte
			err = decompressWithGz(groupsGz, &groupsXml)
			if err != nil {
				return nil, fmt.Errorf("decompress groups xml: %w", err)
			}

			// Decode primary and re-write urls
			primaryXmlGz, err := base64.StdEncoding.DecodeString(revision.PrimaryXml)
			if err != nil {
				return nil, fmt.Errorf("decode primary xml: %w", err)
			}
			var primaryXml []byte
			err = decompressWithGz(primaryXmlGz, &primaryXml)
			if err != nil {
				return nil, fmt.Errorf("decompress primary xml: %w", err)
			}
			primaryXml = []byte(strings.ReplaceAll(string(primaryXml), "rpm:", "rpm_"))
			if err != nil {
				return nil, fmt.Errorf("could not decompress primary.xml: %v", err)
			}
			var primaryRoot yummeta.PrimaryRoot
			err = xml.Unmarshal(primaryXml, &primaryRoot)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal primary.xml: %v", err)
			}
			urlMappings := map[string]string{}
			for _, pkg := range primaryRoot.Packages {
				rpmFileName := filepath.Base(pkg.Location.Href)
				firstLetterLower := strings.ToLower(string(rpmFileName[0]))
				newHref := fmt.Sprintf("%s/%s", firstLetterLower, rpmFileName)
				urlMappings[newHref] = strings.TrimPrefix(pkg.Location.Href, "Packages/")
				pkg.Location.Href = fmt.Sprintf("Packages/%s", newHref)
			}

			// Re-set namespaces because of stupid Go quirks with XML namespaces
			primaryRoot.Rpm = "http://linux.duke.edu/metadata/rpm"
			primaryRoot.XmlnsRpm = "http://linux.duke.edu/metadata/rpm"

			// Remarshal primary and recalculate checksum
			primaryXml, err = xml.Marshal(primaryRoot)
			if err != nil {
				return nil, fmt.Errorf("could not marshal primary.xml: %v", err)
			}
			primaryXml = []byte(strings.ReplaceAll(string(primaryXml), "rpm_", "rpm:"))
			var newPrimaryXmlGz []byte
			err = compressWithGz(primaryXml, &newPrimaryXmlGz)
			if err != nil {
				return nil, fmt.Errorf("could not compress primary.xml: %v", err)
			}
			newChecksums, err := getChecksums(primaryXml, newPrimaryXmlGz)
			if err != nil {
				return nil, fmt.Errorf("could not calculate checksums: %v", err)
			}

			// Fetch repomd and set primary checksum
			newRevisionId := uuid.New()
			repomdXml, err := base64.StdEncoding.DecodeString(revision.RepomdXml)
			if err != nil {
				return nil, fmt.Errorf("decode repomd xml: %w", err)
			}
			var repomdRoot yummeta.RepoMdRoot
			err = xml.Unmarshal(repomdXml, &repomdRoot)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal repomd.xml: %v", err)
			}

			primaryXmlName := fmt.Sprintf("%s-PRIMARY.xml.gz", newRevisionId)
			var filelistsXmlName string
			var otherXmlName string
			var groupsXmlName string
			for _, data := range repomdRoot.Data {
				if data.Type == "primary" {
					data.OpenChecksum.Value = newChecksums[0]
					data.Checksum.Value = newChecksums[1]
					data.OpenSize = len(primaryXml)
					data.Size = len(newPrimaryXmlGz)
					data.Location.Href = fmt.Sprintf("repodata/%s", primaryXmlName)
				} else if data.Type == "filelists" {
					filelistsXmlName = filepath.Base(data.Location.Href)
				} else if data.Type == "other" {
					otherXmlName = filepath.Base(data.Location.Href)
				} else if data.Type == "group" {
					groupsXmlName = filepath.Base(data.Location.Href)
				}
			}
			repomdRoot.Revision = newRevisionId.String()
			repomdXml, err = xml.Marshal(repomdRoot)
			if err != nil {
				return nil, fmt.Errorf("could not marshal repomd.xml: %v", err)
			}

			// Marshal url mappings to json
			urlMappingsJson, err := json.Marshal(urlMappings)
			if err != nil {
				return nil, fmt.Errorf("could not marshal url mappings: %v", err)
			}
			urlMappingsJsonStr := string(urlMappingsJson)

			// Create new hashed repo and revision
			primaryXmlB64 := base64.StdEncoding.EncodeToString(newPrimaryXmlGz)

			// First check if hashed repo already exists
			hashedRepoName := fmt.Sprintf("hashed-%s", repoName)
			hashedRepo, err := c.db.GetRepository(nil, &hashedRepoName, &req.ProjectId.Value)
			if err != nil {
				if err != sql.ErrNoRows {
					return nil, fmt.Errorf("get repository: %w", err)
				}
				hashedRepo, err = c.db.CreateRepositoryWithPackages(hashedRepoName, req.ProjectId.Value, true, []string{})
				if err != nil {
					return nil, fmt.Errorf("create repository: %w", err)
				}
			}

			// Write repomd.xml, primary.xml.gz, filelists.xml.gz, other.xml.gz, groups.xml and groups.xml.gz to tmpDir to run sqliterepo_c
			err = multiErrorCheck(
				ioutil.WriteFile(filepath.Join(repodataDir, "repomd.xml"), repomdXml, 0644),
				ioutil.WriteFile(filepath.Join(repodataDir, primaryXmlName), newPrimaryXmlGz, 0644),
				ioutil.WriteFile(filepath.Join(repodataDir, filelistsXmlName), filelistsGz, 0644),
				ioutil.WriteFile(filepath.Join(repodataDir, otherXmlName), otherGz, 0644),
			)
			if err != nil {
				return nil, fmt.Errorf("write repodata files: %w", err)
			}
			if groupsXmlName != "" {
				err = multiErrorCheck(
					ioutil.WriteFile(filepath.Join(repodataDir, groupsXmlName), groupsXml, 0644),
					ioutil.WriteFile(filepath.Join(repodataDir, groupsXmlName+".gz"), groupsGz, 0644),
				)
				if err != nil {
					return nil, fmt.Errorf("write repodata (groups) files: %w", err)
				}
			}
			err = runCmd("sqliterepo_c", "-f", "--compress-type=gzip", tmpDir)
			if err != nil {
				return nil, fmt.Errorf("run sqliterepo_c: %w", err)
			}

			// Walk repodataDir and upload files ending with .sqlite.gz to S3
			err = filepath.Walk(repodataDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				if !strings.HasSuffix(path, ".sqlite.gz") {
					return nil
				}
				_, err = c.storage.PutObject(filepath.Join("sqlite-files", filepath.Base(path)), path)
				if err != nil {
					return fmt.Errorf("put object: %w", err)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walk repodata dir: %w", err)
			}

			repomdXml, err = ioutil.ReadFile(filepath.Join(repodataDir, "repomd.xml"))
			if err != nil {
				return nil, fmt.Errorf("read repomd.xml: %w", err)
			}
			repomdXmlB64 := base64.StdEncoding.EncodeToString(repomdXml)

			if key != nil {
				ctx := context.TODO()
				signRes, err := c.keykeeper.SignText(ctx, &keykeeperpb.SignTextRequest{
					Text:    string(repomdXml),
					KeyName: key.Name,
				})
				if err != nil {
					return nil, fmt.Errorf("sign text: %w", err)
				}
				_, err = c.storage.PutObjectBytes(filepath.Join("repo-signatures", fmt.Sprintf("%s.xml.asc", newRevisionId.String())), []byte(signRes.Signature))
				if err != nil {
					return nil, fmt.Errorf("put object (repo signature): %w", err)
				}
			}

			newRevision, err := tx.CreateRevisionForRepository(newRevisionId.String(), hashedRepo.ID.String(), revision.Arch, repomdXmlB64, primaryXmlB64, revision.FilelistsXml, revision.OtherXml, revision.UpdateinfoXml, revision.ModuleDefaultsYaml, revision.ModulesYaml, revision.GroupsXml, urlMappingsJsonStr)
			if err != nil {
				return nil, fmt.Errorf("error creating new revision: %v", err)
			}
			ret.RepoRevisions = append(ret.RepoRevisions, newRevision.ID.String())
		}
	}

	err = beginTx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &ret, nil
}
