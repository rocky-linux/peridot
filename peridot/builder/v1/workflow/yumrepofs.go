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
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"github.com/gobwas/glob"
	"github.com/google/uuid"
	"github.com/rocky-linux/srpmproc/modulemd"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"path/filepath"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/composetools"
	peridotdb "peridot.resf.org/peridot/db"
	"peridot.resf.org/peridot/db/models"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/yummeta"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"regexp"
	"strings"
	"time"
)

var (
	globFilterArchRegex = regexp.MustCompile(`^\[(.+)\](.+)$`)
)

type UpdateRepoRequest struct {
	ProjectID   string   `json:"projectId"`
	BuildIDs    []string `json:"buildId"`
	TaskID      *string  `json:"taskId"`
	ForceRepoId string   `json:"forceRepoId"`
	// todo(mustafa): Add support for deleting packages
	Delete           bool `json:"delete"`
	ForceNonModular  bool `json:"forceNonModular"`
	DisableSigning   bool `json:"disableSigning"`
	DisableSetActive bool `json:"disableSetActive"`
	NoDeletePrevious bool `json:"noDeletePrevious"`
}

type CompiledGlobFilter struct {
	Glob glob.Glob
	Arch string
}

type CachedRepo struct {
	Arch           string
	Repo           *models.Repository
	PrimaryRoot    *yummeta.PrimaryRoot
	FilelistsRoot  *yummeta.FilelistsRoot
	OtherRoot      *yummeta.OtherRoot
	Modulemd       []*modulemd.ModuleMd
	GroupsXml      string
	DefaultsYaml   []byte
	ModuleDefaults []*modulemd.Defaults
}

type Cache struct {
	GlobFilters map[string]*CompiledGlobFilter
	Repos       map[string]*CachedRepo
}

// Chain multiple errors and stop processing if any error is returned
func multiErrorCheck(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// Compress the given data using gzip
func compressWithGz(content []byte, out *[]byte) error {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(content)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	*out = buf.Bytes()
	return nil
}

// Decompress the given data using gzip
func decompressWithGz(content []byte, out *[]byte) error {
	var buf bytes.Buffer
	buf.Write(content)
	r, err := gzip.NewReader(&buf)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	*out = b
	return nil
}

// Utility function to marshal XML and chain errors using multiErrorCheck
func xmlMarshal(x interface{}, out *[]byte) error {
	b, err := xml.Marshal(x)
	if err != nil {
		return err
	}

	*out = b
	return nil
}

// Get checksums for multiple files
func getChecksums(content ...[]byte) ([]string, error) {
	var ret []string

	for _, c := range content {
		h := sha256.New()
		_, err := h.Write(c)
		if err != nil {
			return nil, err
		}

		ret = append(ret, hex.EncodeToString(h.Sum(nil)))
	}

	return ret, nil
}

// Utility function to base64 decode and chain errors using multiErrorCheck
func b64Decode(src string, dst *[]byte) error {
	b, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return err
	}

	*dst = b
	return nil
}

func GenerateArchMapForArtifacts(artifacts models.TaskArtifacts, project *models.Project, repo *models.Repository) (map[string][]*models.TaskArtifact, error) {
	var compiledIncludeGlobs []CompiledGlobFilter
	for _, includeGlob := range repo.GlobIncludeFilter {
		var arch string
		var globVal string
		if globFilterArchRegex.MatchString(includeGlob) {
			match := globFilterArchRegex.FindStringSubmatch(includeGlob)
			arch = match[1]
			globVal = match[2]
		} else {
			globVal = includeGlob
		}
		g, err := glob.Compile(globVal)
		if err != nil {
			return nil, fmt.Errorf("failed to compile glob: %v", err)
		}
		compiledIncludeGlobs = append(compiledIncludeGlobs, CompiledGlobFilter{
			Arch: arch,
			Glob: g,
		})
	}

	artifactArchMap := map[string][]*models.TaskArtifact{}
	allArches := project.Archs
	for _, arch := range project.Archs {
		allArches = append(allArches, arch+"-debug")
	}
	for _, arch := range allArches {
		artifactArchMap[arch] = []*models.TaskArtifact{}
	}

	for i, artifact := range artifacts {
		var name string
		var arch string
		base := strings.TrimSuffix(filepath.Base(artifact.Name), ".rpm")
		if rpmutils.NVRUnusualRelease().MatchString(base) {
			nvr := rpmutils.NVRUnusualRelease().FindStringSubmatch(base)
			name = nvr[1]
			arch = nvr[4]
		} else if rpmutils.NVR().MatchString(base) {
			nvr := rpmutils.NVR().FindStringSubmatch(base)
			name = nvr[1]
			arch = nvr[4]
		}

		var anyMetadata anypb.Any
		err := protojson.Unmarshal(artifact.Metadata.JSONText, &anyMetadata)
		if err != nil {
			return nil, err
		}

		var rpmMetadata peridotpb.RpmArtifactMetadata
		err = anypb.UnmarshalTo(&anyMetadata, &rpmMetadata, proto.UnmarshalOptions{})
		if err != nil {
			return nil, err
		}

		var pkgPrimary yummeta.PrimaryRoot
		var pkgFilelists yummeta.FilelistsRoot
		err = multiErrorCheck(
			yummeta.UnmarshalPrimary(rpmMetadata.Primary, &pkgPrimary),
			xml.Unmarshal(rpmMetadata.Filelists, &pkgFilelists),
		)
		if err != nil {
			return nil, err
		}

		// Currently only apply skip rules for non-module components
		// Modules can use filter_rpms
		// todo(mustafa): Revisit
		if !strings.Contains(artifact.Name, ".module+") {
			// Get multilib arches and check if RPM should be multilib
			// But first let's verify that multilib has been enabled for arch
			for _, multilibArch := range repo.Multilib {
				multilibArches, err := composetools.GetMultilibArches(multilibArch)
				if err != nil {
					return nil, fmt.Errorf("could not get multilib arches for %s: %v", artifact.Arch, err)
				}
				// Artifact is not multilib for multilib enabled architecture, skip
				if !utils.StrContains(artifact.Arch, multilibArches) {
					continue
				}

				// Let's check if it's a devel or a runtime package
				isDevel, err := composetools.DevelMultilib(pkgPrimary.Packages[0], pkgFilelists.Packages[0].Files, repo.ExcludeMultilibFilter, repo.AdditionalMultilib)
				if err != nil {
					return nil, fmt.Errorf("could not determine if %s is a devel multilib package: %v", artifact.Name, err)
				}
				isRuntime, err := composetools.RuntimeMultilib(pkgPrimary.Packages[0], pkgFilelists.Packages[0].Files, repo.ExcludeMultilibFilter, repo.AdditionalMultilib)
				if err != nil {
					return nil, fmt.Errorf("could not determine if %s is a runtime multilib package: %v", artifact.Name, err)
				}
				// Not applicable, skip multilib
				if !isDevel && !isRuntime {
					// If it exists in prepopulate, we should probably add it to the list
					if !utils.StrContains(fmt.Sprintf("%s.%s", name, arch), repo.IncludeFilter) {
						continue
					}
				}

				// A multilib package, let's add it to the appropriate arch
				if artifactArchMap[multilibArch] == nil {
					artifactArchMap[multilibArch] = []*models.TaskArtifact{}
				}
				newArtifact := artifacts[i]
				newArtifact.Multilib = true
				artifactArchMap[multilibArch] = append(artifactArchMap[multilibArch], &newArtifact)
			}
		}

		newArtifact := artifacts[i]

		// Let's process the additional packages list, also called the glob include filter in Peridot
		// Let's just mark those artifacts as forced
		for _, includeFilter := range compiledIncludeGlobs {
			if includeFilter.Arch != "" && includeFilter.Arch != newArtifact.Arch {
				continue
			}
			if includeFilter.Glob.Match(name) {
				newArtifact.Forced = true
			}
		}

		if artifact.Arch == "noarch" {
			// Noarch packages should be present in all arch repositories
			for _, arch := range project.Archs {
				// Skip noarch variant
				if arch == "noarch" {
					continue
				}
				if artifactArchMap[arch] == nil {
					artifactArchMap[arch] = []*models.TaskArtifact{}
				}
				artifactArchMap[arch] = append(artifactArchMap[arch], &newArtifact)
			}
		} else {
			if composetools.IsDebugPackage(name) {
				arch = arch + "-debug"
			}
			artifactArchMap[arch] = append(artifactArchMap[arch], &newArtifact)
		}
	}

	return artifactArchMap, nil
}

func (c *Controller) generateIndexedModuleDefaults(projectId string) (map[string]*modulemd.Defaults, error) {
	conf, err := c.db.GetProjectModuleConfiguration(projectId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("could not get project module configuration: %v", err)
	}
	if conf == nil || conf.Default == nil || len(conf.Default) == 0 {
		return nil, nil
	}

	index := map[string]*modulemd.Defaults{}
	for _, moduleDefault := range conf.Default {
		if index[moduleDefault.Name] != nil {
			return nil, fmt.Errorf("duplicate default module %s", moduleDefault.Name)
		}

		document := &modulemd.Defaults{
			Document: "modulemd-defaults",
			Version:  1,
			Data: &modulemd.DefaultsData{
				Module:   moduleDefault.Name,
				Stream:   moduleDefault.Stream,
				Profiles: map[string][]string{},
			},
		}
		if moduleDefault.CommonProfile != nil && len(moduleDefault.CommonProfile) > 0 {
			for _, commonProfile := range moduleDefault.CommonProfile {
				document.Data.Profiles[commonProfile] = []string{"common"}
			}
		}
		if moduleDefault.Profile != nil && len(moduleDefault.Profile) > 0 {
			for _, profile := range moduleDefault.Profile {
				if document.Data.Profiles[profile.Stream] == nil {
					document.Data.Profiles[profile.Stream] = []string{}
				}
				for _, name := range profile.Name {
					document.Data.Profiles[profile.Stream] = append(document.Data.Profiles[profile.Stream], name)
				}
			}
		}

		index[moduleDefault.Name] = document
	}

	return index, nil
}

func (c *Controller) RepoUpdaterWorkflow(ctx workflow.Context, req *UpdateRepoRequest) (*yumrepofspb.UpdateRepoTask, error) {
	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectID),
	})
	if err != nil {
		return nil, fmt.Errorf("could not list projects: %v", err)
	}
	project := projects[0]

	key, err := c.db.GetDefaultKeyForProject(project.ID.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("could not get default key for project %s: %v", project.ID.String(), err)
	}

	var gpgId *string

	var signArtifactsTaskIds []string
	if key != nil && !req.DisableSigning {
		var signFutures []workflow.Future
		signArtifactsCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: 10 * time.Hour,
			StartToCloseTimeout:    24 * time.Hour,
			HeartbeatTimeout:       15 * time.Second,
			TaskQueue:              c.mainQueue,
		})
		for _, buildID := range req.BuildIDs {
			signFutures = append(signFutures, workflow.ExecuteActivity(signArtifactsCtx, c.RequestKeykeeperSignActivity, buildID, key.Name))
		}
		for _, future := range signFutures {
			var signTaskId string
			if err := future.Get(signArtifactsCtx, &signTaskId); err != nil {
				return nil, fmt.Errorf("could not sign artifacts: %v", err)
			}
			if signTaskId != "" {
				signArtifactsTaskIds = append(signArtifactsTaskIds, signTaskId)
			}
		}
		gpgId = &key.GpgId
	}

	var task models.Task
	taskSideEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_YUMREPOFS_UPDATE, &req.ProjectID, req.TaskID)
		if err != nil {
			return &models.Task{}
		}

		return task
	})
	err = taskSideEffect.Get(&task)
	if err != nil {
		return nil, err
	}
	if !task.ProjectId.Valid {
		return nil, fmt.Errorf("could not create task")
	}

	updateRepoCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
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
	updateTask := &yumrepofspb.UpdateRepoTask{}
	err = workflow.ExecuteActivity(updateRepoCtx, c.UpdateRepoActivity, req, task, &gpgId, signArtifactsTaskIds).Get(updateRepoCtx, updateTask)
	if err != nil {
		return nil, err
	}

	return updateTask, nil
}

func (c *Controller) RequestKeykeeperSignActivity(ctx context.Context, buildId string, keyName string) (string, error) {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(4 * time.Second)
		}
	}()

	task, err := c.keykeeper.SignArtifacts(ctx, &keykeeperpb.SignArtifactsRequest{
		BuildId: buildId,
		KeyName: keyName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "no artifacts to sign") {
			return "", nil
		}
		return "", err
	}

	return task.TaskId, nil
}

func (c *Controller) UpdateRepoActivity(ctx context.Context, req *UpdateRepoRequest, task *models.Task, gpgId *string, signTaskIds []string) (*yumrepofspb.UpdateRepoTask, error) {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(4 * time.Second)
		}
	}()

	signArtifactsTasks := &keykeeperpb.BatchSignArtifactsTask{
		Tasks: []*keykeeperpb.SignArtifactsTask{},
	}

	for _, signTaskId := range signTaskIds {
		taskResponse := &keykeeperpb.SignArtifactsTask{}
		wr := c.temporal.GetWorkflow(ctx, signTaskId, "")
		err := wr.Get(ctx, taskResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to get sign artifacts task: %v", err)
		}
		signArtifactsTasks.Tasks = append(signArtifactsTasks.Tasks, taskResponse)
	}

	updateRepoTask := &yumrepofspb.UpdateRepoTask{
		Changes: []*yumrepofspb.RepositoryChange{},
	}

	deferTask, errorDetails, err := c.commonCreateTask(task, updateRepoTask)
	defer deferTask()
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

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
			req.ProjectID,
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

	beginTx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %v", err)
	}
	tx := c.db.UseTransaction(beginTx)

	cache := &Cache{
		GlobFilters: map[string]*CompiledGlobFilter{},
		Repos:       map[string]*CachedRepo{},
	}

	for _, buildID := range req.BuildIDs {
		buildTask, err := c.db.GetTaskByBuildId(buildID)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, fmt.Errorf("failed to get build task: %v", err)
		}

		buildTaskPb, err := buildTask.ToProto(false)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, err
		}

		packageOperationMetadata := &peridotpb.PackageOperationMetadata{}
		err = buildTaskPb.Metadata.UnmarshalTo(packageOperationMetadata)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, err
		}

		if packageOperationMetadata.Modular && !req.ForceNonModular {
			buildResponse := &peridotpb.ModuleBuildTask{}
			err = buildTaskPb.Response.UnmarshalTo(buildResponse)
			if err != nil {
				setInternalError(errorDetails, err)
				return nil, err
			}

			// We need to process the streams in two parts
			// First non-devel stream documents then devel counterparts.
			// Currently we're adding the devel stream documents to the
			// non-devel stream document list with a "-devel" suffix.
			// Let's just take care of the sorting here.
			// nonDevelStreams is just the same stream object but
			// the stream documents for streams ending with "-devel"
			// is removed from the list.
			var nonDevelStreams []*peridotpb.ModuleStream
			var develStreams []*peridotpb.ModuleStream
			for _, stream := range buildResponse.Streams {
				newNonDevelStream := proto.Clone(stream).(*peridotpb.ModuleStream)
				newNonDevelStream.ModuleStreamDocuments = map[string]*peridotpb.ModuleStreamDocument{}
				for arch, streams := range stream.ModuleStreamDocuments {
					if newNonDevelStream.ModuleStreamDocuments[arch] == nil {
						newNonDevelStream.ModuleStreamDocuments[arch] = &peridotpb.ModuleStreamDocument{}
					}
					for docStream, doc := range streams.Streams {
						if !strings.HasSuffix(docStream, "-devel") {
							newNonDevelStream.ModuleStreamDocuments[arch].Streams = map[string][]byte{
								docStream: doc,
							}
						}
					}
				}
				nonDevelStreams = append(nonDevelStreams, newNonDevelStream)

				newDevelStream := proto.Clone(stream).(*peridotpb.ModuleStream)
				newDevelStream.Name = newDevelStream.Name + "-devel"
				newDevelStream.ModuleStreamDocuments = map[string]*peridotpb.ModuleStreamDocument{}
				for arch, streams := range stream.ModuleStreamDocuments {
					if newDevelStream.ModuleStreamDocuments[arch] == nil {
						newDevelStream.ModuleStreamDocuments[arch] = &peridotpb.ModuleStreamDocument{}
					}
					for docStream, doc := range streams.Streams {
						if strings.HasSuffix(docStream, "-devel") {
							newDevelStream.ModuleStreamDocuments[arch].Streams = map[string][]byte{
								strings.TrimSuffix(docStream, "-devel"): doc,
							}
						}
					}
				}
				develStreams = append(develStreams, newDevelStream)
			}

			var combinedStreams []*peridotpb.ModuleStream
			for _, stream := range nonDevelStreams {
				combinedStreams = append(combinedStreams, stream)
			}
			for _, stream := range develStreams {
				combinedStreams = append(combinedStreams, stream)
			}

			var totalChanges []*yumrepofspb.RepositoryChange
			for _, stream := range combinedStreams {
				packageName := fmt.Sprintf("module:%s:%s", stream.Name, stream.Stream)
				taskRes, err := c.makeRepoChanges(tx, req, errorDetails, packageName, buildID, &*stream, gpgId, signArtifactsTasks, cache)
				if err != nil {
					return nil, err
				}
				totalChanges = append(totalChanges, taskRes.Changes...)
			}
			updateRepoTask.Changes = append(updateRepoTask.Changes, totalChanges...)
		} else {
			taskRes, err := c.makeRepoChanges(tx, req, errorDetails, packageOperationMetadata.PackageName, buildID, nil, gpgId, signArtifactsTasks, cache)
			if err != nil {
				return nil, err
			}
			updateRepoTask.Changes = append(updateRepoTask.Changes, taskRes.Changes...)
		}
	}

	// Sort changes by repo name and reduce them together
	repoSortedChanges := map[string]*yumrepofspb.RepositoryChange{}
	for _, change := range updateRepoTask.Changes {
		if repoSortedChanges[change.Name] == nil {
			repoSortedChanges[change.Name] = &yumrepofspb.RepositoryChange{
				Name: change.Name,
			}
		}
		repoSortedChanges[change.Name].AddedPackages = append(repoSortedChanges[change.Name].AddedPackages, change.AddedPackages...)
		repoSortedChanges[change.Name].ModifiedPackages = append(repoSortedChanges[change.Name].ModifiedPackages, change.ModifiedPackages...)
		repoSortedChanges[change.Name].RemovedPackages = append(repoSortedChanges[change.Name].RemovedPackages, change.RemovedPackages...)
		repoSortedChanges[change.Name].AddedModules = append(repoSortedChanges[change.Name].AddedModules, change.AddedModules...)
		repoSortedChanges[change.Name].ModifiedModules = append(repoSortedChanges[change.Name].ModifiedModules, change.ModifiedModules...)
		repoSortedChanges[change.Name].RemovedModules = append(repoSortedChanges[change.Name].RemovedModules, change.RemovedModules...)
	}
	var reducedChanges []*yumrepofspb.RepositoryChange
	for _, changes := range repoSortedChanges {
		reducedChanges = append(reducedChanges, changes)
	}
	updateRepoTask.Changes = reducedChanges

	for _, repo := range cache.Repos {
		c.log.Infof("processing repo %s - %s", repo.Repo.Name, repo.Repo.ID.String())
		primaryRoot := repo.PrimaryRoot
		filelistsRoot := repo.FilelistsRoot
		otherRoot := repo.OtherRoot
		modulesRoot := repo.Modulemd

		var newPrimary []byte
		var newFilelists []byte
		var newOther []byte
		var newModules []byte
		var newGroups []byte
		err = multiErrorCheck(
			xmlMarshal(primaryRoot, &newPrimary),
			xmlMarshal(filelistsRoot, &newFilelists),
			xmlMarshal(otherRoot, &newOther),
		)
		if err != nil {
			return nil, err
		}
		if repo.GroupsXml != "" {
			groupsXml, err := base64.StdEncoding.DecodeString(repo.GroupsXml)
			if err != nil {
				newGroups = []byte{}
			} else {
				err = decompressWithGz(groupsXml, &newGroups)
				if err != nil {
					// Ignore groups if we can't decompress them
					newGroups = []byte{}
				}
			}
		}

		if len(modulesRoot) > 0 {
			var buf bytes.Buffer
			_, _ = buf.WriteString("---\n")
			yamlEncoder := yaml.NewEncoder(&buf)
			for _, def := range repo.ModuleDefaults {
				err := yamlEncoder.Encode(def)
				if err != nil {
					return nil, fmt.Errorf("could not encode module default: %s", err)
				}
				_, _ = buf.WriteString("...\n")
			}
			for _, doc := range modulesRoot {
				err := yamlEncoder.Encode(doc)
				if err != nil {
					return nil, fmt.Errorf("could not encode module document: %s", err)
				}
				_, _ = buf.WriteString("...\n")
			}
			err = yamlEncoder.Close()
			if err != nil {
				return nil, fmt.Errorf("could not close yaml encoder: %v", err)
			}
			newModules = buf.Bytes()
		}

		newPrimary = []byte(strings.ReplaceAll(string(newPrimary), "rpm_", "rpm:"))
		newChecksums, err := getChecksums(newPrimary, newFilelists, newOther, newModules, newGroups)
		if err != nil {
			return nil, err
		}

		var newPrimaryGz []byte
		var newFilelistsGz []byte
		var newOtherGz []byte
		var newModulesGz []byte
		var newGroupsGz []byte
		err = multiErrorCheck(
			compressWithGz(newPrimary, &newPrimaryGz),
			compressWithGz(newFilelists, &newFilelistsGz),
			compressWithGz(newOther, &newOtherGz),
			compressWithGz(newModules, &newModulesGz),
			compressWithGz(newGroups, &newGroupsGz),
		)
		if err != nil {
			return nil, err
		}
		newGzChecksums, err := getChecksums(newPrimaryGz, newFilelistsGz, newOtherGz, newModulesGz, newGroupsGz)
		if err != nil {
			return nil, err
		}

		newRevision := uuid.New()
		repomdRoot := yummeta.RepoMdRoot{
			Rpm:      "http://linux.duke.edu/metadata/rpm",
			XmlnsRpm: "http://linux.duke.edu/metadata/rpm",
			Xmlns:    "http://linux.duke.edu/metadata/repo",
			Revision: repo.Repo.ID.String(),
		}

		blobHref := func(blob string) string {
			ext := "xml"
			if blob == "MODULES" {
				ext = "yaml"
			}
			return fmt.Sprintf("repodata/%s-%s.%s.gz", newRevision.String(), blob, ext)
		}

		now := time.Now()

		// Add primary
		repomdRoot.Data = append(repomdRoot.Data, &yummeta.RepoMdData{
			Type: "primary",
			Checksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: newGzChecksums[0],
			},
			OpenChecksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: newChecksums[0],
			},
			Location: &yummeta.RepoMdDataLocation{
				Href: blobHref("PRIMARY"),
			},
			Timestamp: now.Unix(),
			Size:      len(newPrimaryGz),
			OpenSize:  len(newPrimary),
		})

		// Add filelists
		repomdRoot.Data = append(repomdRoot.Data, &yummeta.RepoMdData{
			Type: "filelists",
			Checksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: newGzChecksums[1],
			},
			OpenChecksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: newChecksums[1],
			},
			Location: &yummeta.RepoMdDataLocation{
				Href: blobHref("FILELISTS"),
			},
			Timestamp: now.Unix(),
			Size:      len(newFilelistsGz),
			OpenSize:  len(newFilelists),
		})

		// Add other
		repomdRoot.Data = append(repomdRoot.Data, &yummeta.RepoMdData{
			Type: "other",
			Checksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: newGzChecksums[2],
			},
			OpenChecksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: newChecksums[2],
			},
			Location: &yummeta.RepoMdDataLocation{
				Href: blobHref("OTHER"),
			},
			Timestamp: now.Unix(),
			Size:      len(newOtherGz),
			OpenSize:  len(newOther),
		})

		// Add modules if any entries
		if len(modulesRoot) > 0 {
			repomdRoot.Data = append(repomdRoot.Data, &yummeta.RepoMdData{
				Type: "modules",
				Checksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: newGzChecksums[3],
				},
				OpenChecksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: newChecksums[3],
				},
				Location: &yummeta.RepoMdDataLocation{
					Href: blobHref("MODULES"),
				},
				Timestamp: now.Unix(),
				Size:      len(newModulesGz),
				OpenSize:  len(newModules),
			})
		}

		// Add comps if not empty
		if len(newGroups) > 0 {
			repomdRoot.Data = append(repomdRoot.Data, &yummeta.RepoMdData{
				Type: "group",
				Checksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: newChecksums[4],
				},
				Location: &yummeta.RepoMdDataLocation{
					Href: strings.TrimSuffix(blobHref("GROUPS"), ".gz"),
				},
				Timestamp: now.Unix(),
				Size:      len(newGroups),
			})
			repomdRoot.Data = append(repomdRoot.Data, &yummeta.RepoMdData{
				Type: "group_gz",
				Checksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: newGzChecksums[4],
				},
				OpenChecksum: &yummeta.RepoMdDataChecksum{
					Type:  "sha256",
					Value: newChecksums[4],
				},
				Location: &yummeta.RepoMdDataLocation{
					Href: blobHref("GROUPS"),
				},
				Timestamp: now.Unix(),
				Size:      len(newGroupsGz),
				OpenSize:  len(newGroups),
			})
		}

		var newRepoMd []byte
		err = xmlMarshal(repomdRoot, &newRepoMd)
		if err != nil {
			return nil, err
		}

		newRepoMdB64 := base64.StdEncoding.EncodeToString(newRepoMd)
		newPrimaryGzB64 := base64.StdEncoding.EncodeToString(newPrimaryGz)
		newFilelistsGzB64 := base64.StdEncoding.EncodeToString(newFilelistsGz)
		newOtherGzB64 := base64.StdEncoding.EncodeToString(newOtherGz)
		newModulesGzB64 := base64.StdEncoding.EncodeToString(newModulesGz)
		defaultsYamlB64 := base64.StdEncoding.EncodeToString(repo.DefaultsYaml)
		newGroupsGzB64 := base64.StdEncoding.EncodeToString(newGroupsGz)

		revision := &models.RepositoryRevision{
			ID:                 newRevision,
			ProjectRepoId:      repo.Repo.ID.String(),
			Arch:               repo.Arch,
			RepomdXml:          newRepoMdB64,
			PrimaryXml:         newPrimaryGzB64,
			FilelistsXml:       newFilelistsGzB64,
			OtherXml:           newOtherGzB64,
			UpdateinfoXml:      "",
			ModuleDefaultsYaml: defaultsYamlB64,
			ModulesYaml:        newModulesGzB64,
			GroupsXml:          newGroupsGzB64,
		}
		_, err = tx.CreateRevisionForRepository(revision.ID.String(), revision.ProjectRepoId, repo.Arch, revision.RepomdXml, revision.PrimaryXml, revision.FilelistsXml, revision.OtherXml, revision.UpdateinfoXml, revision.ModuleDefaultsYaml, revision.ModulesYaml, revision.GroupsXml, "{}")
		if err != nil {
			return nil, fmt.Errorf("error creating new revision: %v", err)
		}
	}

	err = beginTx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %v", err)
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return updateRepoTask, nil
}

// todo(mustafa): Convert to request struct
func (c *Controller) makeRepoChanges(tx peridotdb.Access, req *UpdateRepoRequest, errorDetails *peridotpb.TaskErrorDetails, packageName string, buildId string, moduleStream *peridotpb.ModuleStream, gpgId *string, signArtifactsTasks *keykeeperpb.BatchSignArtifactsTask, cache *Cache) (*yumrepofspb.UpdateRepoTask, error) {
	build, err := c.db.GetBuildByID(buildId)
	if err != nil {
		c.log.Errorf("error getting build: %v", err)
		return nil, err
	}

	repoTask := &yumrepofspb.UpdateRepoTask{
		Changes: []*yumrepofspb.RepositoryChange{},
	}

	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectID),
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}
	project := projects[0]

	artifacts, err := c.db.GetArtifactsForBuild(buildId)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, fmt.Errorf("failed to get artifacts for build: %v", err)
	}

	var currentActiveArtifacts models.TaskArtifacts
	var skipDeleteArtifacts []string
	if moduleStream == nil {
		// Get currently active artifacts
		latestBuilds, err := c.db.GetLatestBuildIdsByPackageName(build.PackageName, nil)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, fmt.Errorf("failed to get latest build ids: %v", err)
		}
		for _, latestBuild := range latestBuilds {
			buildArtifacts, err := c.db.GetArtifactsForBuild(latestBuild)
			if err != nil {
				setInternalError(errorDetails, err)
				return nil, fmt.Errorf("failed to get artifacts for build: %v", err)
			}
			currentActiveArtifacts = append(currentActiveArtifacts, buildArtifacts...)
		}
	} else {
		// Get currently active artifacts
		latestBuilds, err := c.db.GetBuildIDsByPackageNameAndBranchName(build.PackageName, moduleStream.Stream)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, fmt.Errorf("failed to get latest build ids: %v", err)
		}
		for _, latestBuild := range latestBuilds {
			buildArtifacts, err := c.db.GetArtifactsForBuild(latestBuild)
			if err != nil {
				setInternalError(errorDetails, err)
				return nil, fmt.Errorf("failed to get artifacts for build: %v", err)
			}
			currentActiveArtifacts = append(currentActiveArtifacts, buildArtifacts...)
		}
	}

	// Get artifacts to skip deletion
	for _, artifact := range artifacts {
		skipDeleteArtifacts = append(skipDeleteArtifacts, strings.TrimSuffix(filepath.Base(artifact.Name), ".rpm"))
	}

	var repos models.Repositories

	if req.ForceRepoId != "" {
		var err error
		repo, err := c.db.GetRepository(&req.ForceRepoId, nil, nil)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, fmt.Errorf("failed to get repo: %v", err)
		}
		repos = models.Repositories{*repo}
	} else {
		var err error
		repos, err = c.db.FindRepositoriesForPackage(req.ProjectID, packageName, false)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, fmt.Errorf("failed to find repo: %v", err)
		}
	}

	defaultsIndex, err := c.generateIndexedModuleDefaults(req.ProjectID)
	if err != nil && err != sql.ErrNoRows {
		setInternalError(errorDetails, err)
		return nil, err
	}

	for _, repo := range repos {
		c.log.Infof("repo: %s, buildId: %s", repo.ID.String(), buildId)
		artifactArchMap, err := GenerateArchMapForArtifacts(artifacts, &project, &repo)
		if err != nil {
			setInternalError(errorDetails, err)
			return nil, err
		}
		c.log.Infof("generated arch map for build id %s", buildId)

		var compiledExcludeGlobs []*CompiledGlobFilter

		for _, excludeGlob := range repo.ExcludeFilter {
			if cache.GlobFilters[excludeGlob] != nil {
				compiledExcludeGlobs = append(compiledExcludeGlobs, cache.GlobFilters[excludeGlob])
				continue
			}
			var arch string
			var globVal string
			if globFilterArchRegex.MatchString(excludeGlob) {
				match := globFilterArchRegex.FindStringSubmatch(excludeGlob)
				arch = match[1]
				globVal = match[2]
			} else {
				globVal = excludeGlob
			}
			g, err := glob.Compile(globVal)
			if err != nil {
				return nil, fmt.Errorf("failed to compile glob: %v", err)
			}
			globFilter := &CompiledGlobFilter{
				Arch: arch,
				Glob: g,
			}
			compiledExcludeGlobs = append(compiledExcludeGlobs, globFilter)
			cache.GlobFilters[excludeGlob] = globFilter
		}

		for arch, archArtifacts := range artifactArchMap {
			c.log.Infof("arch: %s, buildId: %s", arch, buildId)
			noDebugArch := strings.TrimSuffix(arch, "-debug")
			var streamDocument *modulemd.ModuleMd

			if moduleStream != nil {
				var newArtifacts models.TaskArtifacts

				streamDocuments := moduleStream.ModuleStreamDocuments[noDebugArch]
				if streamDocuments != nil {
					streamDocumentNbc, err := modulemd.Parse(streamDocuments.Streams[moduleStream.Stream])
					if err != nil {
						return nil, fmt.Errorf("failed to decode modulemd: %v", err)
					}
					streamDocument = streamDocumentNbc.V2

					if arch != "src" {
						for _, artifact := range archArtifacts {
							for _, moduleArtifact := range streamDocument.Data.Artifacts.Rpms {
								moduleArtifactNoEpoch := rpmutils.Epoch().ReplaceAllString(moduleArtifact, "")
								if strings.TrimSuffix(filepath.Base(artifact.Name), ".rpm") == moduleArtifactNoEpoch {
									newArtifacts = append(newArtifacts, *artifact)
								}
							}
						}

						artifactArchMap2, err := GenerateArchMapForArtifacts(newArtifacts, &project, &repo)
						if err != nil {
							setInternalError(errorDetails, err)
							return nil, err
						}
						archArtifacts = artifactArchMap2[arch]
					} else {
						// Remove duplicates for src
						seen := make(map[string]bool)
						for _, artifact := range archArtifacts {
							if !seen[artifact.Name] {
								newArtifacts = append(newArtifacts, *artifact)
								seen[artifact.Name] = true
							}
						}

						artifactArchMap2, err := GenerateArchMapForArtifacts(newArtifacts, &project, &repo)
						if err != nil {
							setInternalError(errorDetails, err)
							return nil, err
						}
						archArtifacts = artifactArchMap2[arch]
					}
				}
				if strings.HasSuffix(arch, "-debug") || arch == "src" {
					streamDocument = nil
				}
			}

			changes := &yumrepofspb.RepositoryChange{
				Name:             fmt.Sprintf("%s-%s", repo.Name, arch),
				AddedPackages:    []string{},
				ModifiedPackages: []string{},
				RemovedPackages:  []string{},
				AddedModules:     []string{},
				ModifiedModules:  []string{},
				RemovedModules:   []string{},
			}

			primaryRoot := yummeta.PrimaryRoot{
				Rpm:      "http://linux.duke.edu/metadata/rpm",
				XmlnsRpm: "http://linux.duke.edu/metadata/rpm",
				Xmlns:    "http://linux.duke.edu/metadata/common",
			}
			filelistsRoot := yummeta.FilelistsRoot{
				Xmlns: "http://linux.duke.edu/metadata/filelists",
			}
			otherRoot := yummeta.OtherRoot{
				Xmlns: "http://linux.duke.edu/metadata/other",
			}
			var modulesRoot []*modulemd.ModuleMd

			idArch := fmt.Sprintf("%s-%s", repo.Name, arch)
			idArchNoDebug := fmt.Sprintf("%s-%s", repo.Name, noDebugArch)
			var currentRevision *models.RepositoryRevision
			var groupsXml string
			if cache.Repos[idArchNoDebug] != nil {
				c.log.Infof("found cache for %s", idArchNoDebug)
				if cache.Repos[idArchNoDebug].Modulemd != nil {
					modulesRoot = cache.Repos[idArchNoDebug].Modulemd
				}
			}
			if cache.Repos[idArch] != nil {
				c.log.Infof("found cache for %s", idArch)
				if cache.Repos[idArch].PrimaryRoot != nil {
					primaryRoot = *cache.Repos[idArch].PrimaryRoot
				}
				if cache.Repos[idArch].FilelistsRoot != nil {
					filelistsRoot = *cache.Repos[idArch].FilelistsRoot
				}
				if cache.Repos[idArch].OtherRoot != nil {
					otherRoot = *cache.Repos[idArch].OtherRoot
				}
				groupsXml = cache.Repos[idArch].GroupsXml
			} else {
				c.log.Infof("no cache for %s", idArch)
				var noDebugRevision *models.RepositoryRevision

				currentRevision, err = c.db.GetLatestActiveRepositoryRevision(repo.ID.String(), arch)
				if err != nil {
					if err != sql.ErrNoRows {
						return nil, fmt.Errorf("failed to get latest active repository revision: %v", err)
					}
				}
				if strings.HasSuffix(arch, "-debug") && moduleStream != nil {
					c.log.Infof("arch has debug and module stream is not nil")
					noDebugRevision, err = c.db.GetLatestActiveRepositoryRevision(repo.ID.String(), noDebugArch)
					if err != nil {
						if err != sql.ErrNoRows {
							return nil, fmt.Errorf("failed to get latest active repository revision: %v", err)
						}
					}
				} else {
					noDebugRevision = currentRevision
				}

				if currentRevision != nil {
					c.log.Infof("current revision is not nil")
					if currentRevision.PrimaryXml != "" {
						var primaryXmlGz []byte
						var primaryXml []byte
						err := multiErrorCheck(
							b64Decode(currentRevision.PrimaryXml, &primaryXmlGz),
							decompressWithGz(primaryXmlGz, &primaryXml),
						)
						if err != nil {
							return nil, err
						}

						err = yummeta.UnmarshalPrimary(primaryXml, &primaryRoot)
						if err != nil {
							return nil, err
						}
					}
					if currentRevision.FilelistsXml != "" {
						var filelistsXmlGz []byte
						var filelistsXml []byte

						err := multiErrorCheck(
							b64Decode(currentRevision.FilelistsXml, &filelistsXmlGz),
							decompressWithGz(filelistsXmlGz, &filelistsXml),
							xml.Unmarshal(filelistsXml, &filelistsRoot),
						)
						if err != nil {
							return nil, err
						}
					}
					if currentRevision.OtherXml != "" {
						var otherXmlGz []byte
						var otherXml []byte

						err := multiErrorCheck(
							b64Decode(currentRevision.OtherXml, &otherXmlGz),
							decompressWithGz(otherXmlGz, &otherXml),
							xml.Unmarshal(otherXml, &otherRoot),
						)
						if err != nil {
							return nil, err
						}
					}
					groupsXml = currentRevision.GroupsXml
				}
				if noDebugRevision != nil {
					if noDebugRevision.ModulesYaml != "" {
						var modulesYamlGz []byte
						var modulesYaml []byte
						err := multiErrorCheck(
							b64Decode(noDebugRevision.ModulesYaml, &modulesYamlGz),
							decompressWithGz(modulesYamlGz, &modulesYaml),
						)
						if err != nil {
							return nil, err
						}

						var buf bytes.Buffer
						buf.Write(modulesYaml)
						yamlDecoder := yaml.NewDecoder(&buf)

						for {
							var md modulemd.ModuleMd
							err := yamlDecoder.Decode(&md)
							if err != nil {
								if err == io.EOF {
									break
								}
								if !strings.Contains(err.Error(), "!!seq") {
									return nil, fmt.Errorf("could not decode module document: %v", err)
								}
							}
							// Discard defaults entries, we're generating that
							if md.Document == "modulemd-defaults" {
								continue
							}

							modulesRoot = append(modulesRoot, &*&md)
						}
					}
				}
			}

			c.log.Infof("processing %d artifacts", len(archArtifacts))
			for _, artifact := range archArtifacts {
				// This shouldn't happen
				if !artifact.Metadata.Valid {
					continue
				}
				c.log.Infof("processing artifact %s", artifact.Name)

				var name string
				base := strings.TrimSuffix(filepath.Base(artifact.Name), ".rpm")
				if rpmutils.NVRUnusualRelease().MatchString(base) {
					nvr := rpmutils.NVRUnusualRelease().FindStringSubmatch(base)
					name = nvr[1]
				} else if rpmutils.NVR().MatchString(base) {
					nvr := rpmutils.NVR().FindStringSubmatch(base)
					name = nvr[1]
				}

				noDebugInfoName := strings.ReplaceAll(strings.ReplaceAll(name, "-debuginfo", ""), "-debugsource", "")
				archName := fmt.Sprintf("%s.%s", name, artifact.Arch)
				if strings.HasSuffix(arch, "-debug") {
					archName = fmt.Sprintf("%s.%s", noDebugInfoName, artifact.Arch)
				}

				shouldAdd := true
				if arch != "src" && moduleStream == nil {
					// If repo has a list for inclusion, then the artifact has to pass that first
					if len(repo.IncludeFilter) > 0 {
						// If the artifact isn't forced, it should be in the include list or additional multilib list
						if !artifact.Forced && !utils.StrContains(archName, repo.IncludeFilter) && !utils.StrContains(noDebugInfoName, repo.AdditionalMultilib) {
							shouldAdd = false
						}
					}

					// Check if it matches any exclude filter
					for _, excludeFilter := range compiledExcludeGlobs {
						if excludeFilter.Arch != "" && excludeFilter.Arch != strings.TrimSuffix(arch, "-debug") {
							continue
						}
						if excludeFilter.Glob.Match(noDebugInfoName) || excludeFilter.Glob.Match(archName) {
							shouldAdd = false
						}
					}
				}
				c.log.Infof("should add %s: %v", artifact.Name, shouldAdd)

				baseNoRpm := strings.Replace(filepath.Base(artifact.Name), ".rpm", "", 1)

				if !shouldAdd {
					changes.RemovedPackages = append(changes.RemovedPackages, baseNoRpm)
					continue
				}

				var anyMetadata anypb.Any
				err := protojson.Unmarshal(artifact.Metadata.JSONText, &anyMetadata)
				if err != nil {
					return nil, err
				}

				var rpmMetadata peridotpb.RpmArtifactMetadata
				err = anypb.UnmarshalTo(&anyMetadata, &rpmMetadata, proto.UnmarshalOptions{})
				if err != nil {
					return nil, err
				}

				var pkgPrimary yummeta.PrimaryRoot
				var pkgFilelists yummeta.FilelistsRoot
				var pkgOther yummeta.OtherRoot
				err = multiErrorCheck(
					yummeta.UnmarshalPrimary(rpmMetadata.Primary, &pkgPrimary),
					xml.Unmarshal(rpmMetadata.Filelists, &pkgFilelists),
					xml.Unmarshal(rpmMetadata.Other, &pkgOther),
				)
				if err != nil {
					return nil, err
				}

				c.log.Infof("unmarshalled metadata for %s", artifact.Name)

				if gpgId != nil {
					newObjectKey := fmt.Sprintf("%s/%s/%s", filepath.Dir(artifact.Name), *gpgId, filepath.Base(artifact.Name))

					var signedArtifact *keykeeperpb.SignedArtifact
					for _, signArtifactsTask := range signArtifactsTasks.Tasks {
						for _, sArtifact := range signArtifactsTask.SignedArtifacts {
							if sArtifact.Path == newObjectKey {
								signedArtifact = sArtifact
								break
							}
						}
					}
					if signedArtifact == nil {
						return nil, fmt.Errorf("could not find signed artifact: %s", newObjectKey)
					}

					pkgPrimary.Packages[0].Location.Href = fmt.Sprintf("Packages/%s", newObjectKey)
					pkgPrimary.Packages[0].Checksum.Value = signedArtifact.HashSha256

					for _, pkg := range pkgFilelists.Packages {
						pkg.PkgId = signedArtifact.HashSha256
					}
					for _, pkg := range pkgOther.Packages {
						pkg.PkgId = signedArtifact.HashSha256
					}
				} else {
					pkgPrimary.Packages[0].Location.Href = fmt.Sprintf("Packages/%s", artifact.Name)
				}

				var pkgId *string
				var primaryIndex *int
				for i, primaryPackage := range primaryRoot.Packages {
					if moduleStream == nil {
						// If not a module stream, search for a non-module entry
						// Double check arch as well for multilib purposes
						if primaryPackage.Name == name && !strings.Contains(primaryPackage.Version.Rel, ".module+") && primaryPackage.Arch == artifact.Arch {
							ix := i
							primaryIndex = &ix
							pkgId = &primaryPackage.Checksum.Value
							break
						}
					} else {
						if !strings.Contains(primaryPackage.Version.Rel, ".module+") {
							continue
						}

						for _, streamArtifact := range artifacts {
							var rpmName string
							var rpmVersion string
							var rpmRelease string
							rpmBase := strings.TrimSuffix(filepath.Base(streamArtifact.Name), ".rpm")
							if rpmutils.NVRUnusualRelease().MatchString(rpmBase) {
								nvr := rpmutils.NVRUnusualRelease().FindStringSubmatch(rpmBase)
								rpmName = nvr[1]
								rpmVersion = nvr[2]
								rpmRelease = nvr[3]
							} else if rpmutils.NVR().MatchString(rpmBase) {
								nvr := rpmutils.NVR().FindStringSubmatch(rpmBase)
								rpmName = nvr[1]
								rpmVersion = nvr[2]
								rpmRelease = nvr[3]
							}

							if name == rpmName && primaryPackage.Name == name && primaryPackage.Version.Ver == rpmVersion && primaryPackage.Version.Rel == rpmRelease && primaryPackage.Arch == artifact.Arch {
								ix := i
								primaryIndex = &ix
								pkgId = &primaryPackage.Checksum.Value
								break
							}
						}
					}
				}
				if primaryIndex != nil {
					c.log.Infof("found primary index %d", *primaryIndex)
					if !utils.StrContains(baseNoRpm, changes.ModifiedPackages) && !utils.StrContains(baseNoRpm, changes.AddedPackages) {
						changes.ModifiedPackages = append(changes.ModifiedPackages, baseNoRpm)
					}
					primaryRoot.Packages[*primaryIndex] = pkgPrimary.Packages[0]
				} else {
					c.log.Infof("did not find primary index")
					if !utils.StrContains(baseNoRpm, changes.AddedPackages) && !utils.StrContains(baseNoRpm, changes.ModifiedPackages) {
						changes.AddedPackages = append(changes.AddedPackages, baseNoRpm)
					}
					primaryRoot.Packages = append(primaryRoot.Packages, pkgPrimary.Packages[0])
				}

				var filelistsIndex *int
				if pkgId != nil {
					for i, filelistsPackage := range filelistsRoot.Packages {
						if filelistsPackage.PkgId == *pkgId {
							ix := i
							filelistsIndex = &ix
							break
						}
					}
				}
				if filelistsIndex != nil {
					filelistsRoot.Packages[*filelistsIndex] = pkgFilelists.Packages[0]
				} else {
					filelistsRoot.Packages = append(filelistsRoot.Packages, pkgFilelists.Packages[0])
				}

				var otherIndex *int
				if pkgId != nil {
					for i, otherPackage := range otherRoot.Packages {
						if otherPackage.PkgId == *pkgId {
							ix := i
							otherIndex = &ix
							break
						}
					}
				}
				if otherIndex != nil {
					otherRoot.Packages[*otherIndex] = pkgOther.Packages[0]
				} else {
					otherRoot.Packages = append(otherRoot.Packages, pkgOther.Packages[0])
				}
			}
			c.log.Infof("processed %d artifacts", len(archArtifacts))

			// First let's delete older artifacts
			// Instead of doing re-slicing, let's just not add anything matching the artifacts
			// to a temporary slice, and then let's just swap the primaryRoot.Packages entry
			// with the new slice.
			var nPackages []*yummeta.PrimaryPackage
			var nFilelists []*yummeta.FilelistsPackage
			var nOther []*yummeta.OtherPackage
			var deleteIds []string
			if !req.NoDeletePrevious || moduleStream != nil {
				for _, pkg := range primaryRoot.Packages {
					shouldAdd := true
					for _, artifact := range currentActiveArtifacts {
						noRpmName := strings.TrimSuffix(filepath.Base(artifact.Name), ".rpm")
						if filepath.Base(artifact.Name) == filepath.Base(pkg.Location.Href) && !utils.StrContains(noRpmName, changes.ModifiedPackages) && !utils.StrContains(noRpmName, changes.AddedPackages) && !utils.StrContains(noRpmName, skipDeleteArtifacts) {
							shouldAdd = false
						}
					}
					if !shouldAdd {
						c.log.Infof("deleting %s", pkg.Location.Href)
						deleteIds = append(deleteIds, pkg.Checksum.Value)
					}
				}
				for _, pkg := range primaryRoot.Packages {
					if utils.StrContains(pkg.Checksum.Value, deleteIds) {
						continue
					}

					nPackages = append(nPackages, pkg)
				}
				for _, pkg := range filelistsRoot.Packages {
					if utils.StrContains(pkg.PkgId, deleteIds) {
						continue
					}
					nFilelists = append(nFilelists, &*pkg)
				}
				for _, pkg := range otherRoot.Packages {
					if utils.StrContains(pkg.PkgId, deleteIds) {
						continue
					}
					nOther = append(nOther, &*pkg)
				}
				// Swap filtered slices
				primaryRoot.Packages = nPackages
				filelistsRoot.Packages = nFilelists
				otherRoot.Packages = nOther
			}

			primaryRoot.PackageCount = len(primaryRoot.Packages)
			filelistsRoot.PackageCount = len(filelistsRoot.Packages)
			otherRoot.PackageCount = len(filelistsRoot.Packages)

			// Module builds needs a few more steps
			if moduleStream != nil && streamDocument != nil {
				// If a previous entry exists, we need to overwrite that
				var moduleIndex *int
				for i, moduleMd := range modulesRoot {
					if moduleMd.Data.Name == moduleStream.Name && moduleMd.Data.Stream == moduleStream.Stream {
						c.log.Infof("found existing module entry for %s:%s", moduleStream.Name, moduleStream.Stream)
						moduleIndex = &*&i
						break
					}
				}
				if moduleIndex != nil {
					changes.ModifiedModules = append(changes.ModifiedModules, fmt.Sprintf("%s:%s", moduleStream.Name, moduleStream.Stream))
					modulesRoot[*moduleIndex] = streamDocument
				} else {
					c.log.Infof("adding new module entry for %s:%s", moduleStream.Name, moduleStream.Stream)
					changes.AddedModules = append(changes.AddedModules, fmt.Sprintf("%s:%s", moduleStream.Name, moduleStream.Stream))
					modulesRoot = append(modulesRoot, streamDocument)
				}
			}

			var moduleDefaults []*modulemd.Defaults
			if defaultsIndex != nil {
				for module, defaults := range defaultsIndex {
					for _, rootModule := range modulesRoot {
						if module == rootModule.Data.Name {
							moduleDefaults = append(moduleDefaults, defaults)
							break
						}
					}
				}
			}

			var defaultsYaml []byte

			if len(moduleDefaults) > 0 {
				var defaultsBuf bytes.Buffer
				_, _ = defaultsBuf.WriteString("---\n")
				defaultsEncoder := yaml.NewEncoder(&defaultsBuf)
				for _, def := range moduleDefaults {
					err := defaultsEncoder.Encode(def)
					if err != nil {
						return nil, fmt.Errorf("failed to encode defaults: %v", err)
					}
				}
				err = defaultsEncoder.Close()
				if err != nil {
					return nil, fmt.Errorf("failed to close defaults encoder: %v", err)
				}
				defaultsYaml = defaultsBuf.Bytes()
			}

			nRepo := repo
			cache.Repos[idArch] = &CachedRepo{
				Arch:           arch,
				Repo:           &nRepo,
				PrimaryRoot:    &primaryRoot,
				FilelistsRoot:  &filelistsRoot,
				OtherRoot:      &otherRoot,
				Modulemd:       modulesRoot,
				DefaultsYaml:   defaultsYaml,
				ModuleDefaults: moduleDefaults,
				GroupsXml:      groupsXml,
			}
			if strings.HasSuffix(arch, "-debug") || arch == "src" {
				cache.Repos[idArch].Modulemd = nil
			}
			c.log.Infof("set cache for %s", idArch)

			repoTask.Changes = append(repoTask.Changes, changes)
		}
	}

	if !req.DisableSetActive {
		err = tx.MakeActiveInRepoForPackageVersion(build.PackageVersionId, build.PackageId, build.ProjectId)
		if err != nil {
			c.log.Errorf("failed to set active build for project %s, package %s: %s", project.ID.String(), packageName, err)
			return nil, err
		}
	}

	c.log.Infof("finished processing %d artifacts", len(artifacts))

	return repoTask, nil
}
