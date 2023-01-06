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
	"database/sql"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/gobwas/glob"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotdb "peridot.resf.org/peridot/db"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/yummeta"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
)

var (
	archPrefixRegex = regexp.MustCompile("(\\[.+])?")
)

type RepoSyncPackage struct {
	Name string
	Type peridotpb.PackageType
}

type RepoSyncIndex struct {
	Packages              []RepoSyncPackage
	IncludeFilter         map[string][]string
	ExcludeFilter         []string
	Multilib              []string
	AdditionalMultilib    []string
	ExcludeMultilibFilter []string
	GlobIncludeFilter     []string
}

func recursiveSearchBillyFs(fs billy.Filesystem, root string, ext string) ([]string, error) {
	var files []string
	read, err := fs.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, f := range read {
		if f.IsDir() {
			nFiles, err := recursiveSearchBillyFs(fs, path.Join(root, f.Name()), ext)
			if err != nil {
				return nil, err
			}
			files = append(files, nFiles...)
		} else {
			if path.Ext(f.Name()) == ext {
				files = append(files, path.Join(root, f.Name()))
			}
		}
	}
	return files, nil
}

func (c *Controller) SyncCatalogWorkflow(ctx workflow.Context, req *peridotpb.SyncCatalogRequest, task *models.Task) (*peridotpb.SyncCatalogTask, error) {
	syncCatalogTask := peridotpb.SyncCatalogTask{}

	deferTask, errorDetails, err := c.commonCreateTask(task, &syncCatalogTask)
	defer deferTask()
	if err != nil {
		return nil, err
	}

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

	syncCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 25 * time.Minute,
		StartToCloseTimeout:    15 * time.Minute,
		TaskQueue:              taskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err = workflow.ExecuteActivity(syncCtx, c.SyncCatalogActivity, req).Get(ctx, &syncCatalogTask)
	if err != nil {
		setActivityError(errorDetails, err)
		return nil, err
	}

	if len(syncCatalogTask.ReprocessBuildIds) > 0 {
		yumrepoCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue: "yumrepofs",
		})
		taskID := task.ID.String()
		updateRepoRequest := &UpdateRepoRequest{
			ProjectID: req.ProjectId.Value,
			BuildIDs:  syncCatalogTask.ReprocessBuildIds,
			Delete:    false,
			TaskID:    &taskID,
		}
		updateRepoTask := &yumrepofspb.UpdateRepoTask{}
		err = workflow.ExecuteChildWorkflow(yumrepoCtx, c.RepoUpdaterWorkflow, updateRepoRequest).Get(yumrepoCtx, updateRepoTask)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &syncCatalogTask, nil
}

func globFiltering(incomingFilter []string, currentFilter []string) []string {
	var globs []string
	for _, filter := range incomingFilter {
		if utils.StrContains(filter, currentFilter) {
			continue
		}
		noPrefix := archPrefixRegex.ReplaceAllString(filter, "")
		if !utils.StrContains(noPrefix, globs) {
			globs = append(globs, noPrefix)
		}
	}
	for _, filter := range currentFilter {
		if !utils.StrContains(filter, incomingFilter) {
			noPrefix := archPrefixRegex.ReplaceAllString(filter, "")
			globs = append(globs, noPrefix)
		}
	}
	return globs
}

func kindCatalogSync(tx peridotdb.Access, req *peridotpb.SyncCatalogRequest, catalogs []*peridotpb.CatalogSync) (*peridotpb.KindCatalogSync, error) {
	var ret peridotpb.KindCatalogSync

	repoIndex := map[string]*RepoSyncIndex{}
	// This index contains package names
	// Example:
	//   perl.aarch64 -> perl
	nvrIndex := map[string]string{}
	for _, catalog := range catalogs {
		if catalog.ModuleConfiguration != nil {
			if ret.ModuleConfiguration != nil {
				return nil, fmt.Errorf("multiple module configurations found")
			}
			ret.ModuleConfiguration = catalog.ModuleConfiguration
		}

		for _, pkg := range catalog.Package {
			for _, repo := range pkg.Repository {
				if repoIndex[repo.Name] == nil {
					repoIndex[repo.Name] = &RepoSyncIndex{
						Packages:              []RepoSyncPackage{},
						IncludeFilter:         map[string][]string{},
						ExcludeFilter:         []string{},
						Multilib:              []string{},
						AdditionalMultilib:    []string{},
						ExcludeMultilibFilter: []string{},
						GlobIncludeFilter:     []string{},
					}
				}

				pkgExistsAlready := false
				for _, p := range repoIndex[repo.Name].Packages {
					if p.Name == pkg.Name {
						pkgExistsAlready = true
						break
					}
				}
				if !pkgExistsAlready {
					repoIndex[repo.Name].Packages = append(repoIndex[repo.Name].Packages, RepoSyncPackage{
						Name: pkg.Name,
						Type: pkg.Type,
					})
				}
				for _, moduleStream := range repo.ModuleStream {
					modulePkg := fmt.Sprintf("module:%s:%s", pkg.Name, moduleStream)
					alreadyExists := false
					for _, p := range repoIndex[repo.Name].Packages {
						if p.Name == modulePkg {
							alreadyExists = true
							break
						}
					}
					if !alreadyExists {
						repoIndex[repo.Name].Packages = append(repoIndex[repo.Name].Packages, RepoSyncPackage{
							Name: modulePkg,
							Type: pkg.Type,
						})
					}
				}
				for _, inf := range repo.IncludeFilter {
					nvrIndex[inf] = pkg.Name
					if repoIndex[repo.Name].IncludeFilter[pkg.Name] == nil {
						repoIndex[repo.Name].IncludeFilter[pkg.Name] = []string{}
					}
					if !utils.StrContains(inf, repoIndex[repo.Name].IncludeFilter[pkg.Name]) {
						repoIndex[repo.Name].IncludeFilter[pkg.Name] = append(repoIndex[repo.Name].IncludeFilter[pkg.Name], inf)
					}
				}
				for _, multilib := range repo.Multilib {
					if !utils.StrContains(multilib, repoIndex[repo.Name].Multilib) {
						repoIndex[repo.Name].Multilib = append(repoIndex[repo.Name].Multilib, multilib)
					}
				}
			}
		}
		for _, excludeFilter := range catalog.ExcludeFilter {
			for repoName, repo := range repoIndex {
				if excludeFilter.RepoMatch != "*" {
					matchRegex, err := regexp.Compile(excludeFilter.RepoMatch)
					if err != nil {
						return nil, fmt.Errorf("failed to compile repo match regex: %w", err)
					}
					if !matchRegex.MatchString(repoName) {
						continue
					}
				}

				for _, arch := range excludeFilter.Arch {
					for _, archGlob := range arch.GlobMatch {
						var filterString string
						if arch.Key != "*" {
							filterString = fmt.Sprintf("[%s]", arch.Key)
						}
						filterString += archGlob
						if !utils.StrContains(filterString, repo.ExcludeFilter) {
							repo.ExcludeFilter = append(repo.ExcludeFilter, filterString)
						}
					}
				}
			}
		}
		for _, includeFilter := range catalog.IncludeFilter {
			for repoName, repo := range repoIndex {
				if includeFilter.RepoMatch != "*" {
					matchRegex, err := regexp.Compile(includeFilter.RepoMatch)
					if err != nil {
						return nil, fmt.Errorf("failed to compile repo match regex: %w", err)
					}
					if !matchRegex.MatchString(repoName) {
						continue
					}
				}

				for _, arch := range includeFilter.Arch {
					for _, archGlob := range arch.GlobMatch {
						var filterString string
						if arch.Key != "*" {
							filterString = fmt.Sprintf("[%s]", arch.Key)
						}
						filterString += archGlob
						if !utils.StrContains(filterString, repo.GlobIncludeFilter) {
							repo.GlobIncludeFilter = append(repo.GlobIncludeFilter, filterString)
						}
					}
				}
			}
		}

		for _, repo := range repoIndex {
			if catalog.AdditionalMultilib != nil {
				repo.AdditionalMultilib = catalog.AdditionalMultilib
			}
			if catalog.ExcludeMultilibFilter != nil {
				repo.ExcludeMultilibFilter = catalog.ExcludeMultilibFilter
			}
		}

		logrus.Infof("Syncing %d repositories", len(repoIndex))

		// Create a package index, so we can skip previously synced packages
		packageExistsIndex := map[string]bool{}
		for repoName, _ := range repoIndex {
			dbRepo, err := tx.GetRepository(nil, &repoName, &req.ProjectId.Value)
			if err != nil && err != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to get repository: %w", err)
			}

			// Repo doesn't exist yet, so we can skip it
			if err == sql.ErrNoRows {
				continue
			}
			for _, pkg := range dbRepo.Packages {
				packageExistsIndex[pkg] = true
			}
		}

		for _, repo := range repoIndex {
			for _, pkg := range repo.Packages {
				// Skip if it starts with module: as it's a module stream
				if strings.HasPrefix(pkg.Name, "module:") {
					continue
				}

				// Always refresh type, expensive but necessary
				if err := tx.SetPackageType(req.ProjectId.Value, pkg.Name, pkg.Type); err != nil {
					return nil, fmt.Errorf("failed to update package type: %w", err)
				}

				// Skip if already in project
				if packageExistsIndex[pkg.Name] {
					continue
				}

				pkgId, err := tx.GetPackageID(pkg.Name)
				if err != nil && err != sql.ErrNoRows {
					return nil, fmt.Errorf("failed to check if package exists: %w", err)
				}

				if pkgId != "" {
					pkgs, err := tx.GetPackagesInProject(&peridotpb.PackageFilters{NameExact: wrapperspb.String(pkg.Name)}, req.ProjectId.Value, 0, 1)
					if err != nil {
						return nil, fmt.Errorf("failed to get package %s: %w", pkg.Name, err)
					}

					// Package already in project, skip
					if len(pkgs) > 0 {
						continue
					}
				}

				logrus.Infof("Package %s not found in project %s, creating", pkg.Name, req.ProjectId.Value)
				ret.NewPackages = append(ret.NewPackages, pkg.Name)

				if pkgId == "" {
					dbPkg, err := tx.CreatePackage(pkg.Name, pkg.Type)
					if err != nil {
						return nil, fmt.Errorf("failed to upsert package %s: %w", pkg.Name, err)
					}

					pkgId = dbPkg.ID.String()
				}

				err = tx.AddPackageToProject(req.ProjectId.Value, pkgId, pkg.Type)
				if err != nil {
					return nil, fmt.Errorf("failed to add package %s to project %s: %w", pkg.Name, req.ProjectId.Value, err)
				}
			}
		}

		for repoName, repo := range repoIndex {
			dbRepo, err := tx.GetRepository(nil, &repoName, &req.ProjectId.Value)
			if err != nil && err != sql.ErrNoRows {
				return nil, fmt.Errorf("failed to get repository: %w", err)
			}

			var stringPkgs []string
			if dbRepo != nil {
				stringPkgs = dbRepo.Packages
			}
			for _, pkg := range repo.Packages {
				if !utils.StrContains(pkg.Name, stringPkgs) {
					stringPkgs = append(stringPkgs, pkg.Name)
				}
			}
			var stringIncludeFilter []string
			for _, includeFilter := range repo.IncludeFilter {
				for _, inf := range includeFilter {
					stringIncludeFilter = append(stringIncludeFilter, inf)
				}
			}

			if dbRepo == nil {
				logrus.Infof("Creating repository %s", repoName)
				ret.NewRepositories = append(ret.NewRepositories, repoName)

				dbRepo, err = tx.CreateRepositoryWithPackages(repoName, req.ProjectId.Value, false, stringPkgs)
				if err != nil {
					return nil, fmt.Errorf("failed to create repository: %w", err)
				}
				// This is a new repo, force a new sync for non-new packages as well
				ret.ModifiedPackages = append(ret.ModifiedPackages, stringPkgs...)
			} else {
				logrus.Infof("Updating repository %s", repoName)
				ret.ModifiedRepositories = append(ret.ModifiedRepositories, repoName)

				// Let's go through all lists and filters and create a list of packages this touches
				// The result will then be used to submit correct updates to correct repositories
				// Include filters comes from prepopulate, so we can just use the list as-is.
				for pkg, filters := range repo.IncludeFilter {
					// Only re-process if not already in repository
					shouldContinue := false
					for _, filter := range filters {
						if utils.StrContains(filter, dbRepo.IncludeFilter) {
							continue
						}
						shouldContinue = true
					}
					if !shouldContinue {
						continue
					}
					if !utils.StrContains(pkg, ret.ModifiedPackages) {
						ret.ModifiedPackages = append(ret.ModifiedPackages, pkg)
					}
				}

				// Exclude filter / additional include may include arch prefix, but we'll trim and force a sync on
				// all arches.
				// We're going to store the globs in a temporary slice, and replace
				// wildcard "*" with postgres wildcard "%"
				var globs []string
				globs = append(globs, globFiltering(repo.ExcludeFilter, dbRepo.ExcludeFilter)...)
				globs = append(globs, globFiltering(repo.ExcludeMultilibFilter, dbRepo.ExcludeMultilibFilter)...)
				globs = append(globs, globFiltering(repo.AdditionalMultilib, dbRepo.AdditionalMultilib)...)
				globs = append(globs, globFiltering(repo.GlobIncludeFilter, dbRepo.GlobIncludeFilter)...)
				for _, g := range globs {
					// Last wildcard may be a bit problematic because of conflicting package names,
					// but currently the only thing that does is bringing in extra builds to re-sync.
					// Processing extra builds is better than missing any builds
					// todo(mustafa): Evaluate
					ret.AdditionalNvrGlobs = append(ret.AdditionalNvrGlobs, "%/"+strings.ReplaceAll(g, "*", "%")+"%")
				}
			}

			err = tx.SetRepositoryOptions(dbRepo.ID.String(), stringPkgs, repo.ExcludeFilter, stringIncludeFilter, repo.AdditionalMultilib, repo.ExcludeMultilibFilter, repo.Multilib, repo.GlobIncludeFilter)
			if err != nil {
				return nil, fmt.Errorf("failed to set repository options: %w", err)
			}
		}
	}

	logrus.Infof("New packages: %v", ret.NewPackages)
	logrus.Infof("Modified packages: %v", ret.ModifiedPackages)
	logrus.Infof("New repositories: %v", ret.NewRepositories)
	logrus.Infof("Modified repositories: %v", ret.ModifiedRepositories)
	logrus.Infof("Additional NVR globs: %v", ret.AdditionalNvrGlobs)

	return &ret, nil
}

func processGroupInstallScopedPackageOptions(tx peridotdb.Access, req *peridotpb.SyncCatalogRequest, groupInstallOptionSet *peridotpb.CatalogGroupInstallOption) (scopedPackages *peridotpb.CatalogGroupInstallScopedPackage, err error) {
	// handle scoped packages relationships on packages for injection into build root
	for _, scopedPackage := range groupInstallOptionSet.ScopedPackage {
		filters := &peridotpb.PackageFilters{NameExact: wrapperspb.String(scopedPackage.Name)}
		isGlob := false
		if strings.HasPrefix(scopedPackage.Name, "*") || strings.HasSuffix(scopedPackage.Name, "*") {
			filters.Name = wrapperspb.String(strings.TrimSuffix(strings.TrimPrefix(scopedPackage.Name, "*"), "*"))
			filters.NameExact = nil
			isGlob = true
		}

		pkgs, err := tx.GetPackagesInProject(filters, req.ProjectId.Value, 0, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to get package %s: %w", scopedPackage.Name, err)
		}
		if len(pkgs) == 0 {
			return nil, fmt.Errorf("package %s not found in project %s (scoped package)", scopedPackage.Name, req.ProjectId.Value)
		}

		var dbPkgs []models.Package
		if isGlob {
			for _, p := range pkgs {
				g, err := glob.Compile(scopedPackage.Name)
				if err != nil {
					return nil, fmt.Errorf("failed to compile glob %s: %w", scopedPackage.Name, err)
				}
				if g.Match(p.Name) {
					dbPkgs = append(dbPkgs, p)
				}
			}
		} else {
			if scopedPackage.Name != pkgs[0].Name {
				return nil, fmt.Errorf("package %s not found in project %s (cannot set extra options, not glob)", scopedPackage.Name, req.ProjectId.Value)
			}
			dbPkgs = append(dbPkgs, pkgs[0])
		}
		if len(dbPkgs) == 0 {
			return nil, fmt.Errorf("package %s not found in project %s (cannot set extra options, glob)", scopedPackage.Name, req.ProjectId.Value)
		}

		for _, dbPkg := range dbPkgs {
			err = tx.SetGroupInstallOptionsForPackage(req.ProjectId.Value, dbPkg.Name, scopedPackage.DependsOn, scopedPackage.EnableModule, scopedPackage.DisableModule)
			if err != nil {
				return nil, fmt.Errorf("failed to set scoped package options for package %s", scopedPackage.Name)
			}
		}
	}
	return scopedPackages, nil
}

func processGroupInstallOptionSet(groupInstallOptionSet *peridotpb.CatalogGroupInstallOption) (packages []string, err error) {
	for _, name := range groupInstallOptionSet.Name {
		packages = append(packages, name)
	}
	if len(packages) == 0 {
		return nil, fmt.Errorf("failed to parse packages from GroupInstall options")
	}

	return packages, nil
}

func kindCatalogGroupInstallOptions(tx peridotdb.Access, req *peridotpb.SyncCatalogRequest, groupInstallOptions []*peridotpb.CatalogGroupInstallOptions) (*peridotpb.KindCatalogGroupInstallOptions, error) {
	ret := &peridotpb.KindCatalogGroupInstallOptions{}

	for _, groupInstallOption := range groupInstallOptions {

		// Proces scoped packages
		scopedPackages, err := processGroupInstallScopedPackageOptions(tx, req, groupInstallOption.Srpm)
		if err != nil {
			return nil, fmt.Errorf("failed to parse srpm groupinstall options: %s", err.Error())
		}
		ret.ScopedPackage = append(ret.ScopedPackage, scopedPackages)

		// Process build root packages
		srpmPackages, err := processGroupInstallOptionSet(groupInstallOption.Srpm)
		if err != nil {
			return nil, fmt.Errorf("failed to parse srpm groupinstall options: %w", err)
		}
		buildPackages, err := processGroupInstallOptionSet(groupInstallOption.Build)
		if err != nil {
			return nil, fmt.Errorf("failed to parse build groupinstall options: %w", err)
		}
		err = tx.SetBuildRootPackages(req.ProjectId.Value, srpmPackages, buildPackages)
		if err != nil {
			return nil, fmt.Errorf("failed to set buildroot packages for project: %w", err)
		}

		ret.SrpmPackages = append(ret.SrpmPackages, srpmPackages...)
		ret.BuildPackages = append(ret.BuildPackages, buildPackages...)
	}

	return ret, nil
}

func kindCatalogExtraOptions(tx peridotdb.Access, req *peridotpb.SyncCatalogRequest, extraOptions []*peridotpb.CatalogExtraOptions) (*peridotpb.KindCatalogExtraOptions, error) {
	ret := &peridotpb.KindCatalogExtraOptions{}

	for _, extraOption := range extraOptions {
		for _, pkg := range extraOption.PackageOptions {
			pkgs, err := tx.GetPackagesInProject(&peridotpb.PackageFilters{NameExact: wrapperspb.String(pkg.Name)}, req.ProjectId.Value, 0, 1)
			if err != nil {
				return nil, fmt.Errorf("failed to get package %s: %w", pkg.Name, err)
			}
			if len(pkgs) == 0 {
				return nil, fmt.Errorf("package %s not found in project %s (cannot set extra options)", pkg.Name, req.ProjectId.Value)
			}
			err = tx.SetExtraOptionsForPackage(req.ProjectId.Value, pkg.Name, pkg.With, pkg.Without)
			if err != nil {
				return nil, fmt.Errorf("failed to set extra options for package %s", pkg.Name)
			}
		}
	}

	return ret, nil
}

func checkApplyComps(w *git.Worktree, tx peridotdb.Access, projectId string) error {
	_, err := w.Filesystem.Stat("comps")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat comps: %w", err)
		}
		logrus.Infof("No comps detected")
		return nil
	}

	compsXml, err := w.Filesystem.ReadDir("comps")
	if err != nil {
		return fmt.Errorf("failed to read comps: %w", err)
	}
	for _, comps := range compsXml {
		if comps.IsDir() {
			continue
		}
		if !strings.HasSuffix(comps.Name(), ".xml") {
			logrus.Infof("Skipping non-xml file %s", comps.Name())
			continue
		}
		// Comps have the filename format of REPO-ARCH.xml
		logrus.Infof("Applying comps %s", comps.Name())
		repoArch := strings.SplitN(strings.TrimSuffix(comps.Name(), ".xml"), "-", 2)
		repo := repoArch[0]
		arch := repoArch[1]

		// Check if the repo exists
		dbRepo, err := tx.GetRepository(nil, &repo, &projectId)
		if err != nil {
			logrus.Infof("Repository %s does not exist, skipping comps", repo)
			continue
		}

		compFile, err := w.Filesystem.Open(filepath.Join("comps", comps.Name()))
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", comps.Name(), err)
		}
		defer compFile.Close()
		groupsXml, err := ioutil.ReadAll(compFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", comps.Name(), err)
		}
		var groupsXmlGz []byte
		err = compressWithGz(groupsXml, &groupsXmlGz)
		if err != nil {
			return fmt.Errorf("failed to compress %s: %w", comps.Name(), err)
		}
		groupsXmlB64 := base64.StdEncoding.EncodeToString(groupsXmlGz)
		openChecksums, err := getChecksums(groupsXml)
		if err != nil {
			return fmt.Errorf("failed to get checksums for %s: %w", comps.Name(), err)
		}
		closedChecksums, err := getChecksums(groupsXmlGz)
		if err != nil {
			return fmt.Errorf("failed to get checksums for %s: %w", comps.Name(), err)
		}

		activeRevision, err := tx.GetLatestActiveRepositoryRevision(dbRepo.ID.String(), arch)
		if err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("failed to get latest active revision for repo %s: %w", repo, err)
			}

			revisionId := uuid.New().String()
			repomdRoot := yummeta.RepoMdRoot{
				Rpm:      "http://linux.duke.edu/metadata/rpm",
				XmlnsRpm: "http://linux.duke.edu/metadata/rpm",
				Xmlns:    "http://linux.duke.edu/metadata/repo",
				Revision: revisionId,
			}

			var newRepoMd []byte
			err = xmlMarshal(repomdRoot, &newRepoMd)
			if err != nil {
				return fmt.Errorf("failed to marshal repomd root: %w", err)
			}
			newRepoMdB64 := base64.StdEncoding.EncodeToString(newRepoMd)

			activeRevision, err = tx.CreateRevisionForRepository(revisionId, dbRepo.ID.String(), arch, newRepoMdB64, "", "", "", "", "", "", groupsXmlB64, "{}")
			if err != nil {
				return fmt.Errorf("failed to create revision for repo %s: %w", repo, err)
			}
		}

		newRevision := uuid.New().String()
		blobHref := func(blob string) string {
			ext := "xml"
			if blob == "MODULES" {
				ext = "yaml"
			}
			return fmt.Sprintf("repodata/%s-%s.%s.gz", newRevision, blob, ext)
		}

		now := time.Now()

		// Create data entries
		openEntry := &yummeta.RepoMdData{
			Type: "group",
			Checksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: openChecksums[0],
			},
			Location: &yummeta.RepoMdDataLocation{
				Href: strings.TrimSuffix(blobHref("GROUPS"), ".gz"),
			},
			Timestamp: now.Unix(),
			Size:      len(groupsXml),
		}
		closedEntry := &yummeta.RepoMdData{
			Type: "group_gz",
			Checksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: closedChecksums[0],
			},
			OpenChecksum: &yummeta.RepoMdDataChecksum{
				Type:  "sha256",
				Value: openChecksums[0],
			},
			Location: &yummeta.RepoMdDataLocation{
				Href: blobHref("GROUPS"),
			},
			Timestamp: now.Unix(),
			Size:      len(groupsXmlGz),
			OpenSize:  len(groupsXml),
		}

		// Unmarshal the old repomd
		var oldRepoMd yummeta.RepoMdRoot
		repoMdXml, err := base64.StdEncoding.DecodeString(activeRevision.RepomdXml)
		if err != nil {
			return fmt.Errorf("failed to decode repomd xml: %w", err)
		}
		err = xml.Unmarshal(repoMdXml, &oldRepoMd)
		if err != nil {
			return fmt.Errorf("failed to unmarshal repomd xml: %w", err)
		}
		openEntrySet := false
		closedEntrySet := false
		for i := range oldRepoMd.Data {
			if oldRepoMd.Data[i].Type == "group" {
				oldRepoMd.Data[i] = openEntry
				openEntrySet = true
			}
			if oldRepoMd.Data[i].Type == "group_gz" {
				oldRepoMd.Data[i] = closedEntry
				closedEntrySet = true
			}
		}
		if !openEntrySet {
			oldRepoMd.Data = append(oldRepoMd.Data, openEntry)
		}
		if !closedEntrySet {
			oldRepoMd.Data = append(oldRepoMd.Data, closedEntry)
		}

		// Re-marshal repomd with new group information
		var newRepoMd []byte
		err = xmlMarshal(oldRepoMd, &newRepoMd)
		if err != nil {
			return fmt.Errorf("failed to marshal repomd root: %w", err)
		}
		_, err = tx.CreateRevisionForRepository(
			newRevision,
			dbRepo.ID.String(),
			arch,
			base64.StdEncoding.EncodeToString(newRepoMd),
			activeRevision.PrimaryXml,
			activeRevision.FilelistsXml,
			activeRevision.OtherXml,
			activeRevision.UpdateinfoXml,
			activeRevision.ModuleDefaultsYaml,
			activeRevision.ModulesYaml,
			groupsXmlB64,
			"{}",
		)
		if err != nil {
			return fmt.Errorf("failed to create revision for repo %s: %w", repo, err)
		}
	}

	return nil
}

func (c *Controller) SyncCatalogActivity(req *peridotpb.SyncCatalogRequest) (*peridotpb.SyncCatalogTask, error) {
	var ret peridotpb.SyncCatalogTask

	beginTx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	tx := c.db.UseTransaction(beginTx)

	fmt.Printf("Cloning repository %s\n", req.ScmUrl.Value)

	authenticator, err := c.getAuthenticator(req.ProjectId.Value)
	cloneOpts := &git.CloneOptions{
		URL:        req.ScmUrl.Value,
		Auth:       authenticator,
		RemoteName: "origin",
	}
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), cloneOpts)
	if err != nil {
		if err == transport.ErrInvalidAuthMethod || err == transport.ErrAuthenticationRequired {
			cloneOpts.Auth = nil
			repo, err = git.Clone(memory.NewStorage(), memfs.New(), cloneOpts)
			if err != nil {
				return nil, err
			}
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	fmt.Printf("Checking out branch %s\n", req.Branch.Value)

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", req.Branch.Value),
		Force:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch: %w", err)
	}

	fmt.Println("Scanning repository...")

	var catalogs []*peridotpb.CatalogSync
	var extraOptions []*peridotpb.CatalogExtraOptions
	var groupInstallOptions []*peridotpb.CatalogGroupInstallOptions

	files, err := recursiveSearchBillyFs(w.Filesystem, ".", ".cfg")
	if err != nil {
		fmt.Printf("Failed to scan repository: %s\n", err)
		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}

	for _, file := range files {
		f, err := w.Filesystem.Open(file)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file, err)
		}
		bts, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		messageDecorator := strings.Split(string(bts), "\n")[0]

		if messageDecorator == "" {
			return nil, fmt.Errorf("invalid format for %s", file)
		}

		format := strings.TrimSpace(strings.Replace(messageDecorator, "#", "", 1))
		format = strings.TrimSpace(strings.Replace(format, "kind:", "", 1))

		switch format {
		case "resf.peridot.v1.CatalogSync":
			cs1 := &peridotpb.CatalogSync{}
			err = prototext.Unmarshal(bts, cs1)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kind resf.peridot.v1.CatalogSync: %w", err)
			}
			catalogs = append(catalogs, cs1)
		case "resf.peridot.v1.CatalogExtraOptions":
			ce1 := &peridotpb.CatalogExtraOptions{}
			err = prototext.Unmarshal(bts, ce1)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kind resf.peridot.v1.CatalogExtraOptions: %w", err)
			}
			extraOptions = append(extraOptions, ce1)
		case "resf.peridot.v1.CatalogGroupInstallOptions":
			cg1 := &peridotpb.CatalogGroupInstallOptions{}
			err = prototext.Unmarshal(bts, cg1)
			if err != nil {
				return nil, fmt.Errorf("failed to parse kind resf.peridot.v1.CatalogGroupInstallOptions: %w", err)
			}
			groupInstallOptions = append(groupInstallOptions, cg1)
		default:
			return nil, fmt.Errorf("unknown format %s", format)
		}
	}

	resKindCatalogSync, err := kindCatalogSync(tx, req, catalogs)
	if err != nil {
		return nil, fmt.Errorf("failed to process kind CatalogSync: %w", err)
	}
	ret.CatalogSync = resKindCatalogSync

	// Set module configuration if it exists
	if resKindCatalogSync.ModuleConfiguration != nil {
		err := tx.CreateProjectModuleConfiguration(req.ProjectId.Value, resKindCatalogSync.ModuleConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to create project module configuration: %w", err)
		}
	}

	// Check if we have comps
	err = checkApplyComps(w, tx, req.ProjectId.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to apply comps: %w", err)
	}

	resKindCatalogExtraOptions, err := kindCatalogExtraOptions(tx, req, extraOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to process kind CatalogSyncExtraOptions: %w", err)
	}
	ret.ExtraOptions = resKindCatalogExtraOptions

	resKindCatalogGroupInstallOptions, err := kindCatalogGroupInstallOptions(tx, req, groupInstallOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to process kind CatalogSyncGroupInstallOptions: %w", err)
	}
	ret.GroupInstallOptions = resKindCatalogGroupInstallOptions

	var buildIDs []string
	var newBuildPackages []string
	for _, newPackage := range ret.CatalogSync.NewPackages {
		// Skip module streams
		if strings.Contains(newPackage, "module:") {
			continue
		}
		if utils.StrContains(newPackage, newBuildPackages) {
			continue
		}
		dbIDs, err := c.db.GetLatestBuildIdsByPackageName(newPackage, &req.ProjectId.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, fmt.Errorf("failed to get build for package %s: %w", newPackage, err)
		}
		for _, buildID := range dbIDs {
			if utils.StrContains(buildID, buildIDs) {
				continue
			}
			buildIDs = append(buildIDs, buildID)
		}
		newBuildPackages = append(newBuildPackages, newPackage)
	}
	for _, newPackage := range ret.CatalogSync.ModifiedPackages {
		// Skip module streams
		if strings.Contains(newPackage, "module:") {
			continue
		}
		if utils.StrContains(newPackage, newBuildPackages) {
			continue
		}
		dbIDs, err := c.db.GetLatestBuildIdsByPackageName(newPackage, &req.ProjectId.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, fmt.Errorf("failed to get build for package %s: %w", newPackage, err)
		}
		for _, buildID := range dbIDs {
			if utils.StrContains(buildID, buildIDs) {
				continue
			}
			buildIDs = append(buildIDs, buildID)
		}
		newBuildPackages = append(newBuildPackages, newPackage)
	}
	for _, nvr := range ret.CatalogSync.AdditionalNvrGlobs {
		dbIDs, err := c.db.GetActiveBuildIdsByTaskArtifactGlob(nvr, req.ProjectId.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, fmt.Errorf("failed to get build for nvr %s: %w", nvr, err)
		}
		for _, buildID := range dbIDs {
			if utils.StrContains(buildID, buildIDs) {
				continue
			}
			buildIDs = append(buildIDs, buildID)
		}
	}
	ret.ReprocessBuildIds = buildIDs

	fmt.Println("Following build IDs will be re-processed:")
	for _, buildID := range buildIDs {
		fmt.Printf("\t* %s\n", buildID)
	}

	err = beginTx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &ret, nil
}
