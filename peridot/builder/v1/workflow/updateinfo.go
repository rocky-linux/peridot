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
	"bytes"
	"cirello.io/dynamolock"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"net/http"
	"net/url"
	adminpb "peridot.resf.org/peridot/admin/pb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/yummeta"
	"strings"
	"time"
)

func (c *Controller) SetUpdateInfoActivity(ctx context.Context, req *adminpb.AddUpdateInformationRequest) (*adminpb.AddUpdateInformationTask, error) {
	stopChan := makeHeartbeat(ctx, 10*time.Second)
	defer func() { stopChan <- true }()

	lock, err := dynamolock.New(
		c.dynamodb,
		viper.GetString("dynamodb-table"),
		dynamolock.WithLeaseDuration(10*time.Second),
		dynamolock.WithHeartbeatPeriod(3*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamolock: %v", err)
	}
	defer lock.Close()

	var lockedItem *dynamolock.Lock
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		lockedItem, err = lock.AcquireLock(
			req.ProjectId,
		)
		if err != nil {
			c.log.Errorf("failed to acquire lock: %v", err)
			continue
		}
		break
	}
	didRelease := false
	releaseLock := func() error {
		if didRelease {
			return nil
		}
		lockSuccess, err := lock.ReleaseLock(lockedItem)
		if err != nil {
			return fmt.Errorf("error releasing lock: %v", err)
		}
		if !lockSuccess {
			return fmt.Errorf("lost lock before release")
		}
		return nil
	}
	defer releaseLock()

	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("project not found")
	}
	project := projects[0]

	repositories, err := c.db.FindRepositoriesForProject(project.ID.String(), nil, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %v", err)
	}

	apiBase := "https://apollo.build.resf.org/api/v3/updateinfo"

	for _, repo := range repositories {
		if repo.Name == "all" {
			continue
		}

		for _, arch := range project.Archs {
			realProductName := strings.ReplaceAll(req.ProductName, "$arch", arch)
			apiURL := fmt.Sprintf(
				"%s/%s/%s/updateinfo.xml",
				apiBase,
				url.PathEscape(realProductName),
				repo.Name,
			)
			c.log.Infof("Getting updateinfo %s", apiURL)
			resp, err := http.Get(apiURL)
			if err != nil {
				return nil, fmt.Errorf("failed to get updateinfo: %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				if resp.StatusCode == http.StatusNotFound {
					c.log.Warnf("no updateinfo found for %s/%s", realProductName, repo.Name)
					continue
				} else {
					return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				}
			}

			c.log.Infof("Got updateinfo for %s/%s", realProductName, repo.Name)

			xmlBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read updateinfo: %v", err)
			}

			hasher := sha256.New()

			openSize := len(xmlBytes)
			_, err = hasher.Write(xmlBytes)
			if err != nil {
				return nil, fmt.Errorf("could not hash updateinfo: %v", err)
			}
			openChecksum := hex.EncodeToString(hasher.Sum(nil))
			hasher.Reset()

			var gzippedBuf bytes.Buffer
			w := gzip.NewWriter(&gzippedBuf)
			_, err = w.Write(xmlBytes)
			if err != nil {
				return nil, fmt.Errorf("could not gzip encode: %v", err)
			}
			_ = w.Close()

			closedSize := len(gzippedBuf.Bytes())
			_, err = hasher.Write(gzippedBuf.Bytes())
			if err != nil {
				return nil, fmt.Errorf("could not hash gzipped: %v", err)
			}
			closedChecksum := hex.EncodeToString(hasher.Sum(nil))
			hasher.Reset()
			updateInfoB64 := base64.StdEncoding.EncodeToString(gzippedBuf.Bytes())

			timestamp := time.Now().Unix()

			newRevisionID := uuid.New()
			updateInfoPath := fmt.Sprintf("repodata/%s-UPDATEINFO.xml.gz", newRevisionID.String())
			updateInfoEntry := &yummeta.RepoMdData{
				Type: "updateinfo",
				Checksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: closedChecksum,
				},
				OpenChecksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: openChecksum,
				},
				Location: &yummeta.RepoMdDataLocation{
					Href: updateInfoPath,
				},
				Timestamp: timestamp,
				Size:      closedSize,
				OpenSize:  openSize,
			}

			latestRevision, err := c.db.GetLatestActiveRepositoryRevision(repo.ID.String(), arch)
			if err != nil {
				return nil, fmt.Errorf("failed to get latest active repository revision: %v", err)
			}

			// Get the existing repomd.xml
			// If updateinfo already exists, replace it
			// Else append it
			repomdXml, err := base64.StdEncoding.DecodeString(latestRevision.RepomdXml)
			if err != nil {
				return nil, fmt.Errorf("decode repomd xml: %w", err)
			}
			var repomdRoot yummeta.RepoMdRoot
			err = xml.Unmarshal(repomdXml, &repomdRoot)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal repomd.xml: %v", err)
			}

			doesExist := false
			for i, existingEntry := range repomdRoot.Data {
				if existingEntry.Type == "updateinfo" {
					repomdRoot.Data[i] = updateInfoEntry
					doesExist = true
					break
				}
			}
			if !doesExist {
				repomdRoot.Data = append(repomdRoot.Data, updateInfoEntry)
			}

			// Re-marshal the repomd.xml
			repomdRoot.Revision = newRevisionID.String()
			repomdXml, err = xml.Marshal(repomdRoot)
			if err != nil {
				return nil, fmt.Errorf("could not marshal repomd.xml: %v", err)
			}
			repomdXmlB64 := base64.StdEncoding.EncodeToString(repomdXml)

			// Create new revision
			_, err = c.db.CreateRevisionForRepository(
				newRevisionID.String(),
				latestRevision.ProjectRepoId,
				latestRevision.Arch,
				repomdXmlB64,
				latestRevision.PrimaryXml,
				latestRevision.FilelistsXml,
				latestRevision.OtherXml,
				updateInfoB64,
				latestRevision.ModuleDefaultsYaml,
				latestRevision.ModulesYaml,
				latestRevision.GroupsXml,
				latestRevision.UrlMappings.String(),
			)
			if err != nil {
				return nil, fmt.Errorf("error creating new revision: %v", err)
			}
		}
	}

	return &adminpb.AddUpdateInformationTask{}, nil
}

func (c *Controller) UpdateInfoWorkflow(ctx workflow.Context, req *adminpb.AddUpdateInformationRequest, task *models.Task) (*adminpb.AddUpdateInformationTask, error) {
	var ret adminpb.AddUpdateInformationTask

	deferTask, errorDetails, err := c.commonCreateTask(task, &ret)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	stage2Ctx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 25 * time.Hour,
		StartToCloseTimeout:    30 * time.Hour,
		HeartbeatTimeout:       30 * time.Second,
		TaskQueue:              c.mainQueue,
		// Yumrepofs is locking for a short period so let's not wait too long to retry
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.1,
			MaximumInterval:    25 * time.Second,
		},
	})
	err = workflow.ExecuteActivity(stage2Ctx, c.SetUpdateInfoActivity, req).Get(stage2Ctx, &ret)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &ret, nil
}
