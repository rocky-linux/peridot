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
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/uuid"
	"github.com/rocky-linux/srpmproc/modulemd"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/composetools"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/yummeta"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"sort"
	"strings"
	"time"
)

var (
	ErrInvalidModule = errors.New("invalid module metadata in repo")
)

type ModuleStreamBuildOptions struct {
	Dist           string                         `json:"dist"`
	Increment      int64                          `json:"increment"`
	Name           string                         `json:"name"`
	Stream         string                         `json:"stream"`
	Version        string                         `json:"version"`
	Context        string                         `json:"context"`
	BuildId        string                         `json:"buildId"`
	PackageId      string                         `json:"packageId"`
	BuildBatchId   string                         `json:"buildBatchId"`
	ImportRevision *peridotpb.ImportRevision      `json:"importRevision"`
	Document       *modulemd.ModuleMd             `json:"document"`
	Configuration  *peridotpb.ModuleConfiguration `json:"configuration"`
	Project        *models.Project                `json:"project"`
}

type ModuleStreamBuildTask struct {
	Builds []*peridotpb.SubmitBuildTask `json:"builds"`
}

type ArtifactIndex struct {
	PrimaryRoot   yummeta.PrimaryRoot
	ExcludeArch   []string
	ExclusiveArch []string
	SrpmName      string
	SrpmNevra     string
}

// BuildModuleWorkflow builds a module.
// A child workflow of BuildWorkflow is used to build the module components.
// This workflow is interesting as it's used as a "trigger" for module components
// for all streams of a module.
// A module stream may as well fail, but the other module streams are still built.
// One stream failing will fail the whole module task, but subsequent streams will
// still be recognized as successful.
// Re-triggering a module task will therefore only build the failed streams.
// This workflow will also try to only rebuild what's necessary.
// Currently, only the SCM hash is used to determine if a module component needs to be rebuilt,
// but we may introduce further VRE based checks in the Future.
// User based input for rebuilds may also be introduced in the Future.
//
// Some information about how the module is built below.
// Source: https://rocky-linux.github.io/wiki.rockylinux.org/#team/release_engineering/rpm/local_module_builds/
// Thanks Louis!
//
// Module version has the following format M0m0zYYYYMMDDhhmmss
// M = major, m = minor, z = patch, YYYYMMDDhhmmss = timestamp
// The %dist tag has the following format: .module+elX.Y.Z+i+C
// X = major, Y = minor, Z = patch, i = increment, C = context
//
// We're trying to mimic MBS' build process as much as possible.
// Peridot doesn't need to generate and install arbitrary macro RPMs to define macros.
// For that reason, we can define macros in the container we create.
// Module components requires the following macros:
//   - %dist -> as described above
//   - %_module_build -> the module increment (iteration)
//   - %_module_name -> the module name
//   - %_module_stream -> the module stream
//   - %_module_version -> generated version (as describe above)
//   - %_module_context -> generated context (calculate sha1 of the buildrequires section) # todo(mustafa): Currently the yaml content is used to calculate the context
//
// The macros above will be written to the following file: /etc/rpm/macros.zz-module
// This is to ensure that the macros are applied last.
// Build opt macros are written to the same file.
//
// Where we're differing is how we declare metadata.
// MBS uses data.xmd.mbs, but Peridot uses data.xmd.peridot.
// Peridot is also using a custom (but very translatable) format for module defaults and the platform module.
// See `//peridot/data/el8/85.cfg` for an example
func (c *Controller) BuildModuleWorkflow(ctx workflow.Context, req *peridotpb.SubmitBuildRequest, task *models.Task, extraBuildOptions *peridotpb.ExtraBuildOptions) (*peridotpb.ModuleBuildTask, error) {
	// Prepopulate the task with the module metadata
	moduleBuildTask := &peridotpb.ModuleBuildTask{
		Streams: []*peridotpb.ModuleStream{},
		RepoChanges: &yumrepofspb.UpdateRepoTask{
			Changes: []*yumrepofspb.RepositoryChange{},
		},
	}

	deferTask, errorDetails, err := c.commonCreateTask(task, moduleBuildTask)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	filters := &peridotpb.PackageFilters{}
	switch p := req.Package.(type) {
	case *peridotpb.SubmitBuildRequest_PackageId:
		filters.Id = p.PackageId
	case *peridotpb.SubmitBuildRequest_PackageName:
		filters.NameExact = p.PackageName
	}

	pkgs, err := c.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if len(pkgs) != 1 {
		setPackageNotFoundError(errorDetails, req.ProjectId, ErrorDomainBuildsPeridot)
		return nil, utils.CouldNotRetrieveObjects
	}
	pkg := pkgs[0]

	metadataAnyPb, err := anypb.New(&peridotpb.PackageOperationMetadata{
		PackageName: pkg.Name,
		Modular:     true,
	})
	if err != nil {
		return nil, err
	}
	err = c.db.SetTaskMetadata(task.ID.String(), metadataAnyPb)
	if err != nil {
		return nil, err
	}

	packageType := pkg.PackageType
	if pkg.PackageTypeOverride.Valid {
		packageType = peridotpb.PackageType(pkg.PackageTypeOverride.Int32)
	}

	// If the package does not have type MODULE_FORK or NORMAL_FORK_MODULE then return an error
	// as we can't build a module from a non-module package.
	if packageType != peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK && packageType != peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK_MODULE && packageType != peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT {
		setInternalError(errorDetails, fmt.Errorf("package %s is not a module", pkg.Name))
		return nil, utils.CouldNotRetrieveObjects
	}

	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
		Id: wrapperspb.String(req.ProjectId),
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if len(projects) != 1 {
		setInternalError(errorDetails, errors.New("project could not be found"))
		return nil, utils.CouldNotRetrieveObjects
	}
	project := projects[0]

	increment, err := c.db.GetBuildCount()
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}

	conf, err := c.db.GetProjectModuleConfiguration(req.ProjectId)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}

	// Get import revisions and only use the ones marked modular
	importRevisions, err := c.db.GetLatestImportRevisionsForPackageInProject(pkg.Name, req.ProjectId)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}

	branchIndex := map[string]bool{}
	var streamRevisions models.ImportRevisions
	for _, revision := range importRevisions {
		if revision.Modular {
			if len(req.Branches) > 0 && !utils.StrContains(revision.ScmBranchName, req.Branches) {
				continue
			}
			if !strings.HasPrefix(revision.ScmBranchName, fmt.Sprintf("%s%d%s-stream", project.TargetBranchPrefix, project.MajorVersion, project.BranchSuffix.String)) {
				continue
			}
			if branchIndex[revision.ScmBranchName] {
				continue
			}
			streamRevisions = append(streamRevisions, revision)
			branchIndex[revision.ScmBranchName] = true
		}
	}

	if len(streamRevisions) == 0 {
		noStreamErr := errors.New("no stream revisions found for module " + pkg.Name)
		setActivityError(errorDetails, noStreamErr)
		return nil, noStreamErr
	}

	// Clone the module repo
	upstreamPrefix := fmt.Sprintf("%s/%s", project.TargetGitlabHost, project.TargetPrefix)

	storer := memory.NewStorage()
	worktree := memfs.New()

	repoUrl := fmt.Sprintf("%s/modules/%s", upstreamPrefix, gitlabify(pkg.Name))

	authenticator, err := c.getAuthenticator(req.ProjectId)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	r, err := git.Clone(storer, worktree, &git.CloneOptions{
		URL:  repoUrl,
		Auth: authenticator,
	})
	if err != nil {
		newErr := fmt.Errorf("failed to clone module repo: %s", err)
		setActivityError(errorDetails, newErr)
		return nil, newErr
	}
	w, err := r.Worktree()
	if err != nil {
		newErr := fmt.Errorf("failed to get worktree: %s", err)
		setActivityError(errorDetails, newErr)
		return nil, newErr
	}

	var streamBuildOptions []*ModuleStreamBuildOptions

	// Checkout every stream revision and trigger a build
	for _, revision := range streamRevisions {
		err := w.Checkout(&git.CheckoutOptions{
			Hash:  plumbing.NewHash(revision.ScmHash),
			Force: true,
		})
		if err != nil {
			newErr := fmt.Errorf("failed to checkout revision %s: %s", revision.ScmHash, err)
			setActivityError(errorDetails, newErr)
			return nil, newErr
		}

		// Read the yaml file with the same name as package name
		yamlF, err := w.Filesystem.Open(fmt.Sprintf("%s.yaml", pkg.Name))
		if err != nil {
			newErr := fmt.Errorf("could not open yaml file from modules repo in branch %s: %v", revision.ScmBranchName, err)
			setActivityError(errorDetails, newErr)
			return nil, newErr
		}
		yamlContent, err := ioutil.ReadAll(yamlF)
		if err != nil {
			newErr := fmt.Errorf("could not read yaml file from modules repo in branch %s: %v", revision.ScmBranchName, err)
			setActivityError(errorDetails, newErr)
			return nil, newErr
		}

		// Parse yaml content to module metadata
		moduleMdNotBackwardsCompatible, err := modulemd.Parse(yamlContent)
		if err != nil {
			newErr := fmt.Errorf("could not parse yaml file from modules repo in branch %s: %v", revision.ScmBranchName, err)
			setActivityError(errorDetails, newErr)
			return nil, newErr
		}

		var moduleMd *modulemd.ModuleMd
		if moduleMdNotBackwardsCompatible.V2 != nil {
			moduleMd = moduleMdNotBackwardsCompatible.V2
		} else if moduleMdNotBackwardsCompatible.V3 != nil {
			v3 := moduleMdNotBackwardsCompatible.V3
			moduleMd = &modulemd.ModuleMd{
				Document: "modulemd",
				Version:  2,
				Data: &modulemd.Data{
					Name:          v3.Data.Name,
					Stream:        v3.Data.Stream,
					Summary:       v3.Data.Summary,
					Description:   v3.Data.Description,
					ServiceLevels: nil,
					License: &modulemd.License{
						Module: v3.Data.License,
					},
					Xmd:        v3.Data.Xmd,
					References: v3.Data.References,
					Profiles:   v3.Data.Profiles,
					Profile:    v3.Data.Profile,
					API:        v3.Data.API,
					Filter:     v3.Data.Filter,
					BuildOpts:  nil,
					Components: v3.Data.Components,
					Artifacts:  nil,
				},
			}
			if len(v3.Data.Configurations) > 0 {
				cfg := v3.Data.Configurations[0]
				if cfg.BuildOpts != nil {
					moduleMd.Data.BuildOpts = &modulemd.BuildOpts{
						Rpms:   cfg.BuildOpts.Rpms,
						Arches: cfg.BuildOpts.Arches,
					}
					moduleMd.Data.Dependencies = []*modulemd.Dependencies{
						{
							BuildRequires: cfg.BuildRequires,
							Requires:      cfg.Requires,
						},
					}
				}
			}
		}
		if moduleMd.Data.Name == "" {
			moduleMd.Data.Name = pkg.Name
		}

		// Invalid modulemd in repo
		if moduleMd.Data == nil || moduleMd.Data.Components == nil {
			setActivityError(errorDetails, ErrInvalidModule)
			errorDetails.ErrorInfo.Metadata["module"] = pkg.Name
			return nil, ErrInvalidModule
		}

		moduleVersion := fmt.Sprintf("%d0%d0%d%s", conf.Platform.Major, conf.Platform.Minor, conf.Platform.Patch, time.Now().Format("20060102150405"))

		hasher := sha1.New()
		_, err = hasher.Write(yamlContent)
		if err != nil {
			newErr := fmt.Errorf("could not hash yaml file from modules repo in branch %s: %v", revision.ScmBranchName, err)
			setActivityError(errorDetails, newErr)
			return nil, newErr
		}
		context := hex.EncodeToString(hasher.Sum(nil))[:8]
		// todo(mustafa): Evaluate whether we should do `module_` instead of `module+`
		// Currently RHEL uses `+` but Fedora uses `_` so we'll use `+` for now
		dist := fmt.Sprintf("module+el%d.%d.%d+%d+%s", conf.Platform.Major, conf.Platform.Minor, conf.Platform.Patch, increment, context)

		build, err := c.db.CreateBuild(pkg.ID.String(), revision.PackageVersionId, task.ID.String(), req.ProjectId)
		if err != nil {
			err = fmt.Errorf("failed to create build: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}

		for name, component := range moduleMd.Data.Components.Rpms {
			component.Name = name
		}

		options := &ModuleStreamBuildOptions{
			Dist:           dist,
			Increment:      increment,
			Name:           moduleMd.Data.Name,
			Stream:         moduleMd.Data.Stream,
			Version:        moduleVersion,
			Context:        context,
			ImportRevision: revision.ToProto(),
			Document:       moduleMd,
			Configuration:  conf,
			Project:        &*&project,
			BuildId:        build.ID.String(),
			BuildBatchId:   extraBuildOptions.BuildBatchId,
			PackageId:      pkg.ID.String(),
		}

		streamBuildOptions = append(streamBuildOptions, options)
	}

	var futures []FutureContext

	for _, options := range streamBuildOptions {
		// Create a new build for each stream revision
		subtriggerCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue: c.mainQueue,
		})

		future := workflow.ExecuteChildWorkflow(subtriggerCtx, c.BuildModuleStreamWorkflow, req, options, task.ID.String())
		futures = append(futures, FutureContext{
			Ctx:       subtriggerCtx,
			Future:    future,
			TaskQueue: c.mainQueue,
		})
	}

	for _, future := range futures {
		var moduleStream peridotpb.ModuleStream
		err = future.Future.Get(ctx, &moduleStream)
		if err != nil {
			err = fmt.Errorf("failed to build module stream: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}
		moduleBuildTask.Streams = append(moduleBuildTask.Streams, &moduleStream)
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_RUNNING

	// Save once here so RepoUpdaterWorkflow can use it
	deferTask()

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	var buildIDs []string
	for _, options := range streamBuildOptions {
		buildIDs = append(buildIDs, options.BuildId)
	}

	yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue: "yumrepofs",
	})
	taskID := task.ID.String()
	updateRepoRequest := &UpdateRepoRequest{
		ProjectID: req.ProjectId,
		BuildIDs:  buildIDs,
		Delete:    false,
		TaskID:    &taskID,
	}
	updateRepoTask := &yumrepofspb.UpdateRepoTask{}
	err = workflow.ExecuteChildWorkflow(yumrepoCtx, c.RepoUpdaterWorkflow, updateRepoRequest).Get(ctx, updateRepoTask)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	moduleBuildTask.RepoChanges.Changes = append(moduleBuildTask.RepoChanges.Changes, updateRepoTask.Changes...)

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return moduleBuildTask, nil
}

// BuildModuleStreamWorkflow triggers a build for components for the given stream revision of a module
func (c *Controller) BuildModuleStreamWorkflow(ctx workflow.Context, req *peridotpb.SubmitBuildRequest, streamBuildOptions *ModuleStreamBuildOptions, parentTaskId string) (*peridotpb.ModuleStream, error) {
	buildTask := &peridotpb.ModuleStream{
		Dist:                  streamBuildOptions.Dist,
		Increment:             streamBuildOptions.Increment,
		Name:                  streamBuildOptions.Name,
		Stream:                streamBuildOptions.Stream,
		Version:               streamBuildOptions.Version,
		Context:               streamBuildOptions.Context,
		ImportRevision:        streamBuildOptions.ImportRevision,
		Configuration:         streamBuildOptions.Configuration,
		Builds:                []*peridotpb.SubmitBuildTask{},
		ModuleStreamDocuments: map[string]*peridotpb.ModuleStreamDocument{},
	}

	zzModuleMacro := `
%dist .{dist}
%modularitylabel {name}:{stream}:{version}:{context}
%_module_build {increment}
%_module_name {name}
%_module_stream {stream}
%_module_version {version}
%_module_context {context}

{buildopts}
`
	var buildOpts string

	md := streamBuildOptions.Document
	if md.Data.BuildOpts != nil && md.Data.BuildOpts.Rpms != nil {
		buildOpts = md.Data.BuildOpts.Rpms.Macros
	}

	zzModuleMacroCompiled := strings.NewReplacer(
		"{dist}", streamBuildOptions.Dist,
		"{name}", streamBuildOptions.Name,
		"{stream}", streamBuildOptions.Stream,
		"{version}", streamBuildOptions.Version,
		"{context}", streamBuildOptions.Context,
		"{buildopts}", buildOpts,
	).Replace(zzModuleMacro)

	buildOrderIndex := map[int][]*modulemd.ComponentRPM{}
	for _, component := range streamBuildOptions.Document.Data.Components.Rpms {
		if buildOrderIndex[component.Buildorder] == nil {
			buildOrderIndex[component.Buildorder] = []*modulemd.ComponentRPM{}
		}
		buildOrderIndex[component.Buildorder] = append(buildOrderIndex[component.Buildorder], component)
	}
	var buildOrders []int
	for buildOrder := range buildOrderIndex {
		buildOrders = append(buildOrders, buildOrder)
	}
	sort.Ints(buildOrders)

	var repo *models.Repository
	var extraRepos []*peridotpb.ExtraYumrepofsRepo
	if len(buildOrders) > 1 {
		var err error
		repo, err = c.db.CreateRepositoryWithPackages(uuid.New().String(), req.ProjectId, true, []string{})
		if err != nil {
			c.log.Errorf("failed to create repository: %v", err)
			return nil, status.Error(codes.Internal, "failed to create repository")
		}
		extraRepos = []*peridotpb.ExtraYumrepofsRepo{
			{
				Name:           repo.Name,
				ModuleHotfixes: true,
				Priority:       -1,
				IgnoreExclude:  true,
			},
		}
	}

	// todo(mustafa): Very unfinished, and doesn't support all features yet
	// Trigger a BuildWorkflow for each component
	// Currently we treat all modules as very simple declarations
	// Building all project architectures is the default and cannot currently be overridden by the MD.
	// The MD can't override generated values such as repository or cache either yet.
	// Name specified by the component is also currently ignored and the key is forcefully used.
	// Whatever is available in the latest revision of yumrepofs for the project is what's used (including external repos).
	var nonY1Excludes []string
	for _, buildOrder := range buildOrders {
		var futures []FutureContext
		for _, component := range buildOrderIndex[buildOrder] {
			var buildRequiresModules []string
			for _, dependency := range md.Data.Dependencies {
				for module, stream := range dependency.BuildRequires {
					if module == "platform" {
						continue
					}

					if len(stream) == 0 {
						return nil, status.Error(codes.Unimplemented, "buildrequires that don't specify a stream are not supported until we have a registry")
					}
					for _, s := range stream {
						buildRequiresModules = append(buildRequiresModules, fmt.Sprintf("%s:%s", module, s))
					}
				}
			}

			name := component.Name
			childSubmitBuildRequest := &peridotpb.SubmitBuildRequest{
				ProjectId: req.ProjectId,
				Package: &peridotpb.SubmitBuildRequest_PackageName{
					PackageName: wrapperspb.String(name),
				},
				ScmHash:       wrapperspb.String(component.Ref),
				DisableChecks: req.DisableChecks,
				SideNvrs:      req.SideNvrs,
			}
			extraOptions := &peridotpb.ExtraBuildOptions{
				DisableYumrepofsUpdates: true,
				BuildArchExtraFiles: map[string]string{
					"/etc/rpm/macros.zz-module": zzModuleMacroCompiled,
				},
				ReusableBuildId:     streamBuildOptions.BuildId,
				ExtraYumrepofsRepos: extraRepos,
				BuildBatchId:        streamBuildOptions.BuildBatchId,
				Modules:             buildRequiresModules,
				ForceDist:           streamBuildOptions.Dist,
				ExcludePackages:     nonY1Excludes,
			}

			task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_BUILD, &req.ProjectId, &parentTaskId)
			if err != nil {
				c.log.Errorf("could not create build task in BuildModuleStreamWorkflow: %v", err)
				return nil, status.Error(codes.InvalidArgument, "could not create build task")
			}

			buildCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				TaskQueue: c.mainQueue,
			})
			futures = append(futures, FutureContext{
				Ctx:       buildCtx,
				Future:    workflow.ExecuteChildWorkflow(buildCtx, c.BuildWorkflow, childSubmitBuildRequest, task, extraOptions),
				TaskQueue: c.mainQueue,
			})
		}
		for _, future := range futures {
			var btask peridotpb.SubmitBuildTask
			err := future.Future.Get(ctx, &btask)
			if err != nil {
				return nil, err
			}
			buildTask.Builds = append(buildTask.Builds, &btask)
			for _, a := range btask.Artifacts {
				match := rpmutils.NVR().FindStringSubmatch(filepath.Base(a.Name))
				if !utils.StrContains(match[1], nonY1Excludes) {
					nonY1Excludes = append(nonY1Excludes, match[1])
				}
			}

			if repo != nil {
				yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
					TaskQueue: "yumrepofs",
				})
				updateRepoRequest := &UpdateRepoRequest{
					ProjectID:        req.ProjectId,
					TaskID:           &btask.BuildTaskId,
					BuildIDs:         []string{btask.BuildId},
					Delete:           false,
					ForceRepoId:      repo.ID.String(),
					ForceNonModular:  true,
					DisableSigning:   true,
					DisableSetActive: true,
				}
				updateRepoTask := &yumrepofspb.UpdateRepoTask{}
				err = workflow.ExecuteChildWorkflow(yumrepoCtx, c.RepoUpdaterWorkflow, updateRepoRequest).Get(ctx, updateRepoTask)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Create an index, so we don't have to re-unmarshal later on
	artifactPrimaryIndex := map[string]ArtifactIndex{}

	artifacts, err := c.db.GetArtifactsForBuild(streamBuildOptions.BuildId)
	if err != nil {
		return nil, err
	}

	// Pre-warm SRPM NEVRAs
	srpmNevras := map[string]string{}
	for _, artifact := range artifacts {
		if artifact.Arch != "src" {
			continue
		}

		rpmArtifactMetadata := &peridotpb.RpmArtifactMetadata{}
		artifactMetadataAny := &anypb.Any{}
		err = protojson.Unmarshal(artifact.Metadata.JSONText, artifactMetadataAny)
		if err != nil {
			return nil, err
		}
		err := artifactMetadataAny.UnmarshalTo(rpmArtifactMetadata)
		if err != nil {
			return nil, err
		}

		var primary yummeta.PrimaryRoot
		err = yummeta.UnmarshalPrimary(rpmArtifactMetadata.Primary, &primary)
		if err != nil {
			return nil, err
		}

		for _, build := range buildTask.Builds {
			if build.PackageName == primary.Packages[0].Name {
				srpmNevras[build.PackageName] = composetools.GenNevraPrimaryPkg(primary.Packages[0])
			}
		}
	}

	// Group content licenses
	var licenses []string
	for _, build := range buildTask.Builds {
		var buildTaskArtifactNames []string
		for _, a := range build.Artifacts {
			buildTaskArtifactNames = append(buildTaskArtifactNames, a.Name)
		}

		for _, artifact := range artifacts {
			if !utils.StrContains(artifact.Name, buildTaskArtifactNames) {
				continue
			}

			rpmArtifactMetadata := &peridotpb.RpmArtifactMetadata{}
			artifactMetadataAny := &anypb.Any{}
			err = protojson.Unmarshal(artifact.Metadata.JSONText, artifactMetadataAny)
			if err != nil {
				return nil, err
			}
			err := artifactMetadataAny.UnmarshalTo(rpmArtifactMetadata)
			if err != nil {
				return nil, err
			}

			var primary yummeta.PrimaryRoot
			err = yummeta.UnmarshalPrimary(rpmArtifactMetadata.Primary, &primary)
			if err != nil {
				return nil, err
			}
			for _, pkg := range primary.Packages {
				if !utils.StrContains(pkg.Format.RpmLicense, licenses) && len(pkg.Format.RpmLicense) > 0 {
					licenses = append(licenses, pkg.Format.RpmLicense)
				}
			}

			if err != nil {
				return nil, err
			}
			artifactPrimaryIndex[artifact.Name] = ArtifactIndex{
				PrimaryRoot:   primary,
				ExcludeArch:   rpmArtifactMetadata.ExcludeArch,
				ExclusiveArch: rpmArtifactMetadata.ExclusiveArch,
				SrpmName:      build.PackageName,
				SrpmNevra:     srpmNevras[build.PackageName],
			}
		}
	}

	// Generate a modulemd for each arch
	for _, arch := range streamBuildOptions.Project.Archs {
		newMd := copyModuleMd(*md)
		err := fillInRpmArtifactsForModuleMd(newMd, streamBuildOptions, buildTask, artifactPrimaryIndex, arch, licenses, false)
		if err != nil {
			return nil, err
		}
		err = fillInRpmArtifactsForModuleMd(newMd, streamBuildOptions, buildTask, artifactPrimaryIndex, arch, licenses, true)
		if err != nil {
			return nil, err
		}
	}

	return buildTask, nil
}

func doesRpmPassFilter(artifact *ArtifactIndex, md *modulemd.ModuleMd, arch string, multilibArches []string) bool {
	// If we have a whitelist, then we need to check if the artifact is in the whitelist
	var whitelist []string
	if md.Data.BuildOpts != nil && md.Data.BuildOpts.Rpms != nil {
		whitelist = md.Data.BuildOpts.Rpms.Whitelist
	}
	if len(whitelist) > 0 {
		if !utils.StrContains(artifact.SrpmName, whitelist) {
			return false
		}
	}

	// Check if the RPM is filtered
	var rpmFilters []string
	if md.Data.Filter != nil {
		rpmFilters = md.Data.Filter.Rpms
	}
	if len(rpmFilters) > 0 {
		if utils.StrContains(artifact.PrimaryRoot.Packages[0].Name, rpmFilters) {
			return false
		}
	}

	// Get the correct RPM entry from mmd
	var component *modulemd.ComponentRPM
	for _, mdComponent := range md.Data.Components.Rpms {
		if mdComponent.Name == artifact.SrpmName {
			component = mdComponent
			break
		}
	}
	// This should in theory never happen
	if component == nil {
		return false
	}

	// Check if the multilib RPM should be included
	// This is done by checking if the md is declaring the
	// component arch as multilib compatible
	if !utils.StrContains(arch, component.Multilib) && utils.StrContains(artifact.PrimaryRoot.Packages[0].Arch, multilibArches) {
		return false
	}

	// If whitelist exists, the components and whitelists may have different names
	// Skip multilib then
	if len(whitelist) > 0 {
		if !utils.StrContains(artifact.PrimaryRoot.Packages[0].Arch, []string{arch, "noarch"}) {
			return false
		}
	}

	return true
}

func copyModuleMd(md modulemd.ModuleMd) *modulemd.ModuleMd {
	ret := modulemd.ModuleMd{
		Document: md.Document,
		Version:  md.Version,
		Data: &modulemd.Data{
			Name:          md.Data.Name,
			Stream:        md.Data.Stream,
			Version:       md.Data.Version,
			StaticContext: md.Data.StaticContext,
			Context:       md.Data.Context,
			Arch:          md.Data.Arch,
			Summary:       md.Data.Summary,
			Description:   md.Data.Description,
			ServiceLevels: map[modulemd.ServiceLevelType]*modulemd.ServiceLevel{},
			License:       &modulemd.License{},
			Xmd:           map[string]map[string]string{},
			Dependencies:  []*modulemd.Dependencies{},
			References:    &modulemd.References{},
			Profiles:      map[string]*modulemd.Profile{},
			Profile:       map[string]*modulemd.Profile{},
			API:           &modulemd.API{},
			Filter:        &modulemd.API{},
			BuildOpts:     &modulemd.BuildOpts{},
			Components:    &modulemd.Components{},
			Artifacts:     &modulemd.Artifacts{},
		},
	}
	if md.Data.ServiceLevels != nil {
		for k, v := range md.Data.ServiceLevels {
			c := *v
			ret.Data.ServiceLevels[k] = &c
		}
	} else {
		ret.Data.ServiceLevels = nil
	}
	if md.Data.License != nil {
		c := *md.Data.License
		ret.Data.License = &c
	} else {
		ret.Data.License = nil
	}
	if md.Data.Xmd != nil {
		for k, v := range md.Data.Xmd {
			c := map[string]string{}
			for k2, v2 := range v {
				c[k2] = v2
			}
			ret.Data.Xmd[k] = c
		}
	} else {
		ret.Data.Xmd = nil
	}
	if md.Data.Dependencies != nil {
		for _, v := range md.Data.Dependencies {
			c := *v
			ret.Data.Dependencies = append(ret.Data.Dependencies, &c)
		}
	}
	if md.Data.References != nil {
		c := *md.Data.References
		ret.Data.References = &c
	} else {
		ret.Data.References = nil
	}
	if md.Data.Profiles != nil {
		for k, v := range md.Data.Profiles {
			c := *v
			ret.Data.Profiles[k] = &c
		}
	} else {
		ret.Data.Profiles = nil
	}
	if md.Data.Profile != nil {
		for k, v := range md.Data.Profile {
			c := *v
			ret.Data.Profile[k] = &c
		}
	} else {
		ret.Data.Profile = nil
	}
	if md.Data.API != nil {
		c := *md.Data.API
		ret.Data.API = &c
	} else {
		ret.Data.API = nil
	}
	if md.Data.Filter != nil {
		c := *md.Data.Filter
		ret.Data.Filter = &c
	} else {
		ret.Data.Filter = nil
	}
	if md.Data.BuildOpts != nil {
		c := *md.Data.BuildOpts
		if md.Data.BuildOpts.Rpms != nil {
			rpms := *md.Data.BuildOpts.Rpms
			c.Rpms = &rpms
		}
		ret.Data.BuildOpts = &c
	} else {
		ret.Data.BuildOpts = nil
	}
	if md.Data.Components != nil {
		c := *md.Data.Components
		if md.Data.Components.Rpms != nil {
			rpms := map[string]*modulemd.ComponentRPM{}
			for k, v := range md.Data.Components.Rpms {
				x := *v
				rpms[k] = &x
			}
			c.Rpms = rpms
		}
		if md.Data.Components.Modules != nil {
			modules := map[string]*modulemd.ComponentModule{}
			for k, v := range md.Data.Components.Modules {
				x := *v
				modules[k] = &x
			}
			c.Modules = modules
		}
		ret.Data.Components = &c
	} else {
		ret.Data.Components = nil
	}
	if md.Data.Artifacts != nil {
		c := *md.Data.Artifacts
		if md.Data.Artifacts.RpmMap != nil {
			rpmMap := map[string]map[string]*modulemd.ArtifactsRPMMap{}
			for k, v := range md.Data.Artifacts.RpmMap {
				x := map[string]*modulemd.ArtifactsRPMMap{}
				for k2, v2 := range v {
					y := *v2
					x[k2] = &y
				}
				rpmMap[k] = x
			}
		}
		ret.Data.Artifacts = &c
	} else {
		ret.Data.Artifacts = nil
	}

	return &ret
}

func fillInRpmArtifactsForModuleMd(md *modulemd.ModuleMd, streamBuildOptions *ModuleStreamBuildOptions, buildTask *peridotpb.ModuleStream, artifactPrimaryIndex map[string]ArtifactIndex, arch string, licenses []string, devel bool) error {
	newMd := copyModuleMd(*md)
	// Set version, context, arch and licenses
	newMd.Data.Version = streamBuildOptions.Version
	newMd.Data.Context = streamBuildOptions.Context
	newMd.Data.Arch = arch
	newMd.Data.License.Content = licenses

	// Set buildrequires platform to the one used
	didSetPlatform := false
	platform := streamBuildOptions.Configuration.Platform
	platformReq := fmt.Sprintf("el%d.%d.%d", platform.Major, platform.Minor, platform.Patch)
	for _, dep := range newMd.Data.Dependencies {
		if dep.BuildRequires["platform"] != nil {
			dep.BuildRequires["platform"] = []string{platformReq}
			didSetPlatform = true
			break
		}
	}
	if !didSetPlatform {
		newMd.Data.Dependencies = append(newMd.Data.Dependencies, &modulemd.Dependencies{
			BuildRequires: map[string][]string{
				"platform": {platformReq},
			},
		})
	}

	// Set arch for components
	for _, component := range newMd.Data.Components.Rpms {
		component.Arches = streamBuildOptions.Project.Archs
	}

	// Generate the artifacts list
	newMd.Data.Artifacts = &modulemd.Artifacts{
		Rpms: []string{},
	}

	multilibArches, err := composetools.GetMultilibArches(arch)
	if err != nil {
		return err
	}

	exclusiveArches := []string{arch, "noarch"}

	var binaryRpmNames []string
	var includedRpmNames []string
	var includedSrpmNames []string
	var nonDevelSourceRpms []string
	sourceRpms := map[string]string{}
	debugRpms := map[string]yummeta.PrimaryPackage{}
	nonDebugRpms := map[string]yummeta.PrimaryPackage{}

	nevraArtifactIndex := map[string]ArtifactIndex{}

	// We need to add non-debug RPMs first and determine if
	// the src and debug RPMs are included
	for _, artifact := range artifactPrimaryIndex {
		// We only have one package per primary root, so we can use it directly
		pkg := artifact.PrimaryRoot.Packages[0]
		nevra := composetools.GenNevraPrimaryPkg(pkg)
		nevraArtifactIndex[nevra] = artifact
		if pkg.Arch == "src" {
			sourceRpms[pkg.Name] = nevra
		} else {
			binaryRpmNames = append(binaryRpmNames, pkg.Name)
		}

		if composetools.IsDebugPackage(pkg.Name) {
			debugRpms[nevra] = *pkg
		} else {
			nonDebugRpms[nevra] = *pkg
		}
	}

	// Create a collected list of all the rpms
	// We're going to execute common actions for all of them
	collected := map[string]yummeta.PrimaryPackage{}
	for nevra, pkg := range debugRpms {
		collected[nevra] = pkg
	}
	for nevra, pkg := range nonDebugRpms {
		collected[nevra] = pkg
	}

	// Sort collected by rpmObj.Name
	var collectedNames []string
	for name := range collected {
		collectedNames = append(collectedNames, name)
	}
	sort.Strings(collectedNames)

	for _, nevra := range collectedNames {
		rpmObj := collected[nevra]
		// Skip source RPMs for now as they're
		// only added if the main RPM is included
		if rpmObj.Arch == "src" {
			continue
		}

		// If an RPM is not multilib compatible, not the same architecture
		// or is neither "noarch", then it's not included
		if !utils.StrContains(rpmObj.Arch, multilibArches) && !utils.StrContains(rpmObj.Arch, []string{arch, "noarch"}) {
			continue
		}

		artifact := nevraArtifactIndex[nevra]
		excludeArch := artifact.ExcludeArch
		exclusiveArch := artifact.ExclusiveArch
		// Skip RPM if it's excluded or exclusive to another arch
		if excludeArch != nil && len(excludeArch) > 0 && len(utils.IntersectString(excludeArch, exclusiveArches)) > 0 {
			continue
		}
		if exclusiveArch != nil && len(exclusiveArch) > 0 && len(utils.IntersectString(exclusiveArch, exclusiveArches)) == 0 {
			continue
		}

		shouldInclude := false

		if composetools.IsDebugPackage(rpmObj.Name) {
			// Debug packages are only included if the main package is included
			// or if the main package doesn't exist at all
			// This is all of course if the debug package itself isn't filtered out
			rpmNameWithoutDebug := composetools.StripDebugSuffixes(rpmObj.Name)
			if utils.StrContains(rpmNameWithoutDebug, includedRpmNames) || (!utils.StrContains(rpmNameWithoutDebug, binaryRpmNames) && utils.StrContains(artifact.SrpmName, includedSrpmNames)) {
				shouldInclude = doesRpmPassFilter(&artifact, md, arch, multilibArches)
			}
		} else {
			shouldInclude = doesRpmPassFilter(&artifact, md, arch, multilibArches)
		}

		// Source RPM should only be included in the "devel" variant
		// if all components created by the respective source RPM is
		// included in the "devel" variant.
		// Otherwise, including only in the main variant is fine.
		if shouldInclude {
			nonDevelSourceRpms = append(nonDevelSourceRpms, artifact.SrpmNevra)
			includedRpmNames = append(includedRpmNames, rpmObj.Name)
			includedSrpmNames = append(includedSrpmNames, artifact.SrpmName)
		}

		// Only components that wasn't included in the main variant
		// should be included in the "devel" one
		if devel && shouldInclude {
			continue
		} else if !devel && !shouldInclude {
			// This RPM is not included in the main variant
			continue
		}

		if !utils.StrContains(nevra, newMd.Data.Artifacts.Rpms) {
			newMd.Data.Artifacts.Rpms = append(newMd.Data.Artifacts.Rpms, nevra)
		}
	}

	if devel {
		var srpms []string
		for _, v := range sourceRpms {
			srpms = append(srpms, v)
		}
		var develOnlySourceRpms []string
		for _, sourceRpm := range srpms {
			if !utils.StrContains(sourceRpm, nonDevelSourceRpms) {
				develOnlySourceRpms = append(develOnlySourceRpms, sourceRpm)
			}
		}
		for _, sourceRpm := range develOnlySourceRpms {
			if !utils.StrContains(sourceRpm, newMd.Data.Artifacts.Rpms) {
				newMd.Data.Artifacts.Rpms = append(newMd.Data.Artifacts.Rpms, sourceRpm)
			}
		}
	} else {
		for _, sourceRpm := range nonDevelSourceRpms {
			if !utils.StrContains(sourceRpm, newMd.Data.Artifacts.Rpms) {
				newMd.Data.Artifacts.Rpms = append(newMd.Data.Artifacts.Rpms, sourceRpm)
			}
		}
	}

	streamName := streamBuildOptions.Stream
	if devel {
		newMd.Data.Name = streamBuildOptions.Name + "-devel"
		streamName = streamBuildOptions.Stream + "-devel"
	}

	yamlBytes, err := yaml.Marshal(&newMd)
	if err != nil {
		return err
	}
	yamlBytes = append([]byte("---\n"), yamlBytes...)
	if buildTask.ModuleStreamDocuments[arch] == nil {
		buildTask.ModuleStreamDocuments[arch] = &peridotpb.ModuleStreamDocument{Streams: map[string][]byte{}}
	}

	buildTask.ModuleStreamDocuments[arch].Streams[streamName] = yamlBytes

	return nil
}
