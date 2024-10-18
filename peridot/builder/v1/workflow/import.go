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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	http2 "net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
	"github.com/rocky-linux/srpmproc/pkg/srpmproc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/rpmbuild"
	"peridot.resf.org/utils"
)

// This should probably reside somewhere else
// todo(mustafa): Move some stuff into another package

var (
	ModuleReleaseRegex = regexp.MustCompile(`(.+\.module)\+(el\d+\.\d+\.\d+)\+(\d+)\+([A-Za-z0-9]{8})(\..+)?`)
)

type OpenPatchSection string

type UpstreamDistGitActivityRequest struct {
	Project        *models.Project           `json:"project,omitempty"`
	Package        *models.Package           `json:"package,omitempty"`
	ParentTaskId   string                    `json:"parent_task_id,omitempty"`
	VersionRelease *peridotpb.VersionRelease `json:"version_release,omitempty"`
}

type UpstreamDistGitActivityResponse struct {
	ImportRevisions []*peridotpb.ImportRevision `json:"import_revisions"`
}

type sideEffectImpBatch struct {
	Task   *models.Task
	Import *models.Import
}

const (
	OpenPatchSrc     OpenPatchSection = "src"
	OpenPatchRpms    OpenPatchSection = "rpms"
	OpenPatchModules OpenPatchSection = "modules"
)

var fieldValueRegex = regexp.MustCompile("^[a-zA-Z0-9]+:")

func innerCompress(path string, stripHeaderName string, tw *tar.Writer, fs billy.Filesystem) error {
	ls, err := fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, elem := range ls {
		filePath := filepath.Join(path, elem.Name())

		if elem.IsDir() {
			if err := innerCompress(filePath, stripHeaderName, tw, fs); err != nil {
				return err
			}
		} else {
			header, err := tar.FileInfoHeader(elem, filePath)
			if err != nil {
				return err
			}

			// Make tars reproducible
			header.Name = strings.Replace(filepath.ToSlash(filePath), stripHeaderName, "", 1)
			header.Gid = 0
			header.Uid = 0
			header.ModTime = time.Unix(0, 0)
			header.AccessTime = time.Unix(0, 0)

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			f, err := fs.Open(filePath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func compressFolder(path string, stripHeaderName string, buf io.Writer, fs billy.Filesystem) error {
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	defer func() {
		_ = tw.Close()
		_ = gz.Close()
	}()

	return innerCompress(path, stripHeaderName, tw, fs)
}

func genPushBranch(bp string, suffix string, mv int) string {
	return fmt.Sprintf("%s%d%s", bp, mv, suffix)
}

func recursiveRemove(path string, fs billy.Filesystem) error {
	read, err := fs.ReadDir(path)
	if err != nil {
		return fmt.Errorf("could not read dir: %v", err)
	}

	for _, fi := range read {
		fullPath := filepath.Join(path, fi.Name())

		if fi.IsDir() {
			err := recursiveRemove(fullPath, fs)
			if err != nil {
				return err
			}
		} else {
			err = fs.Remove(fullPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkoutRepo(project *models.Project, sourceBranchPrefix string, remoteUrl string, authenticator transport.AuthMethod, tagMode git.TagMode, onDisk bool) (*git.Repository, *git.Worktree, error) {
	var fs billy.Filesystem
	if onDisk {
		_ = os.MkdirAll(rpmbuild.GetCloneDirectory(), 0755)
		fs = osfs.New(rpmbuild.GetCloneDirectory())
	} else {
		fs = memfs.New()
	}
	repo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:        remoteUrl,
		Auth:       authenticator,
		RemoteName: "origin",
		Tags:       tagMode,
	})

	w, err := repo.Worktree()
	if err != nil {
		return repo, nil, err
	}

	pushBranch := genPushBranch(sourceBranchPrefix, project.BranchSuffix.String, project.MajorVersion)
	logrus.Infof("Checking out branch %s in remote %s", pushBranch, remoteUrl)

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", pushBranch),
		Force:  true,
	})
	if err != nil {
		return repo, nil, err
	}

	return repo, w, nil
}

func GetTargetScmUrl(project *models.Project, packageName string, section OpenPatchSection) string {
	return strings.Replace(strings.Replace(fmt.Sprintf("%s/%s/%s/%s.git", project.TargetGitlabHost, project.TargetPrefix, section, gitlabify(packageName)), "//", "/", -1), ":/", "://", 1)
}

func (c *Controller) getAuthenticator(projectId string) (transport.AuthMethod, error) {
	// Retrieve keys for the project
	projectKeys, err := c.db.GetProjectKeys(projectId)
	if err != nil {
		return nil, err
	}

	authenticator := &http.BasicAuth{
		Username: projectKeys.GitlabUsername,
		Password: projectKeys.GitlabSecret,
	}
	return authenticator, nil
}

func (c *Controller) getGitlabClient(project *models.Project) (*gitlab.Client, error) {
	// Retrieve keys for the project
	projectKeys, err := c.db.GetProjectKeys(project.ID.String())
	if err != nil {
		return nil, err
	}

	return gitlab.NewClient(projectKeys.GitlabSecret, gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", project.TargetGitlabHost)))
}

func (c *Controller) createProjectOrMakePublic(project *models.Project, packageName string, section OpenPatchSection) error {
	if !project.GitMakePublic {
		return nil
	}
	packageName = gitlabify(packageName)

	gitlabClient, err := c.getGitlabClient(project)
	if err != nil {
		return err
	}

	name := url.QueryEscape(fmt.Sprintf("%s/%s", project.TargetPrefix, section))
	ns, _, err := gitlabClient.Namespaces.GetNamespace(name)
	if err != nil {
		return err
	}

	_, resp, err := gitlabClient.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:        &packageName,
		NamespaceID: &ns.ID,
		Visibility:  gitlab.Visibility(gitlab.PublicVisibility),
	})
	if err != nil {
		if resp.StatusCode != http2.StatusBadRequest {
			return err
		} else {
			projectName := fmt.Sprintf("%s/%s/%s", project.TargetPrefix, section, packageName)
			_, _, err = gitlabClient.Projects.EditProject(projectName, &gitlab.EditProjectOptions{
				Visibility: gitlab.Visibility(gitlab.PublicVisibility),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Controller) ImportPackageBatchWorkflow(ctx workflow.Context, req *peridotpb.ImportPackageBatchRequest, importBatchId string, user *utils.ContextUser) error {
	var futures []FutureContext

	for _, importReq := range req.Imports {
		importReq.ProjectId = req.ProjectId
		triggerCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			TaskQueue:           c.mainQueue,
			WorkflowTaskTimeout: 3 * time.Hour,
		})
		futures = append(futures, FutureContext{
			Ctx:       triggerCtx,
			Future:    workflow.ExecuteChildWorkflow(triggerCtx, c.TriggerImportFromBatchWorkflow, importReq, importBatchId, user),
			TaskQueue: c.mainQueue,
		})
	}

	// Import failures doesn't mean a batch trigger has failed
	// A batch can contain failed imports
	for _, future := range futures {
		_ = future.Future.Get(future.Ctx, nil)
	}

	return nil
}

// TriggerImportFromBatchWorkflow is a sub-workflow to create a task and trigger an import
func (c *Controller) TriggerImportFromBatchWorkflow(ctx workflow.Context, req *peridotpb.ImportPackageRequest, importBatchId string, user *utils.ContextUser) error {
	var sideEffect sideEffectImpBatch
	sideEffectCall := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		beginTx, err := c.db.Begin()
		if err != nil {
			c.log.Errorf("error starting transaction: %s", err)
			return nil
		}
		tx := c.db.UseTransaction(beginTx)

		filters := &peridotpb.PackageFilters{}
		switch p := req.Package.(type) {
		case *peridotpb.ImportPackageRequest_PackageId:
			filters.Id = p.PackageId
		case *peridotpb.ImportPackageRequest_PackageName:
			filters.NameExact = p.PackageName
		}

		pkgs, err := c.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
		if len(pkgs) != 1 {
			return nil
		}

		projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
			Id: wrapperspb.String(req.ProjectId),
		})
		if err != nil {
			return nil
		}
		if len(projects) != 1 {
			return nil
		}
		project := projects[0]

		task, err := tx.CreateTask(user, "noarch", peridotpb.TaskType_TASK_TYPE_IMPORT, &req.ProjectId, nil)
		if err != nil {
			c.log.Errorf("could not create import task in TriggerImportFromBatchWorkflow: %v", err)
			return nil
		}

		metadataAnyPb, err := anypb.New(&peridotpb.PackageOperationMetadata{
			PackageName: pkgs[0].Name,
		})
		if err != nil {
			return nil
		}
		err = tx.SetTaskMetadata(task.ID.String(), metadataAnyPb)
		if err != nil {
			c.log.Errorf("could not set metadata for import task in TriggerImportFromBatchWorkflow: %v", err)
			return nil
		}

		imp, err := tx.CreateImport(GetTargetScmUrl(&project, pkgs[0].Name, "rpms"), task.ID.String(), pkgs[0].ID.String(), req.ProjectId)
		if err != nil {
			return nil
		}

		err = tx.AttachImportToBatch(imp.ID.String(), importBatchId)
		if err != nil {
			return nil
		}

		err = beginTx.Commit()
		if err != nil {
			c.log.Errorf("error committing transaction: %s", err)
			return nil
		}

		return &sideEffectImpBatch{
			Task:   task,
			Import: imp,
		}
	})
	err := sideEffectCall.Get(&sideEffect)
	if err != nil {
		return err
	}
	if sideEffect.Task == nil {
		return fmt.Errorf("could not create import task")
	}

	buildCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: sideEffect.Task.ID.String(),
		TaskQueue:  c.mainQueue,
	})
	return workflow.ExecuteChildWorkflow(buildCtx, c.ImportPackageWorkflow, req, sideEffect.Task, sideEffect.Import).Get(buildCtx, nil)
}

// ImportPackageWorkflow imports a package from the project specified target upstream.
// Currently VRE is not reported nor respected
// todo(mustafa): Actually respect VRE
func (c *Controller) ImportPackageWorkflow(ctx workflow.Context, req *peridotpb.ImportPackageRequest, task *models.Task, imp *models.Import) (*peridotpb.ImportPackageTask, error) {
	importPackageTask := peridotpb.ImportPackageTask{}

	deferTask, errorDetails, err := c.commonCreateTask(task, &importPackageTask)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	filters := &peridotpb.PackageFilters{}
	switch p := req.Package.(type) {
	case *peridotpb.ImportPackageRequest_PackageId:
		filters.Id = p.PackageId
	case *peridotpb.ImportPackageRequest_PackageName:
		filters.NameExact = p.PackageName
	}

	pkgs, err := c.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	if len(pkgs) != 1 {
		setPackageNotFoundError(errorDetails, req.ProjectId, ErrorDomainImportsPeridot)
		return nil, utils.CouldNotRetrieveObjects
	}
	pkg := pkgs[0]

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

	// Provision a new worker specifically to import
	// Imports can be done with any architecture
	importTaskQueue, cleanupWorkerImport, err := c.provisionWorker(ctx, &ProvisionWorkerRequest{
		TaskId:       task.ID.String(),
		ParentTaskId: task.ParentTaskId,
		Purpose:      "import",
		Arch:         "noarch",
		ProjectId:    req.ProjectId,
	})
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}
	defer cleanupWorkerImport()

	importCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,
		HeartbeatTimeout:    10 * time.Second,
		TaskQueue:           importTaskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 6,
		},
	})

	var importRevisions []*peridotpb.ImportRevision

	packageType := pkg.PackageType
	if pkg.PackageTypeOverride.Valid {
		packageType = peridotpb.PackageType(pkg.PackageTypeOverride.Int32)
	}

	switch packageType {
	case peridotpb.PackageType_PACKAGE_TYPE_NORMAL,
		peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK,
		peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK,
		peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_COMPONENT,
		peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK_MODULE,
		peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK_MODULE_COMPONENT,
		peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT:
		distGitReq := &UpstreamDistGitActivityRequest{
			Project:        &project,
			Package:        &pkg,
			ParentTaskId:   task.ID.String(),
			VersionRelease: req.Vre,
		}
		var res UpstreamDistGitActivityResponse
		err = workflow.ExecuteActivity(importCtx, c.UpstreamDistGitActivity, distGitReq).Get(ctx, &res)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}
		importRevisions = res.ImportRevisions
		break
	case peridotpb.PackageType_PACKAGE_TYPE_NORMAL_SRC:
		var packageSrcGitRes peridotpb.PackageSrcGitResponse
		err = workflow.ExecuteActivity(importCtx, c.PackageSrcGitActivity, pkg.Name, project, task.ID.String()).Get(ctx, &packageSrcGitRes)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}

		var revision peridotpb.ImportRevision
		err = workflow.ExecuteActivity(importCtx, c.UpdateDistGitForSrcGitActivity, pkg.Name, project, &packageSrcGitRes).Get(ctx, &revision)
		if err != nil {
			setActivityError(errorDetails, err)
			return nil, err
		}
		importRevisions = append(importRevisions, &revision)
		break
	default:
		return nil, status.Error(codes.InvalidArgument, "unsupported import source")
	}

	beginTx, err := c.db.Begin()
	if err != nil {
		setInternalError(errorDetails, err)
		return nil, err
	}

	tx := c.db.UseTransaction(beginTx)

	// Loop through all revisions and deactivate previous import revisions (if exists)
	// The latest import revisions should be the only one active
	if !req.SetInactive {
		// Deactivate previous package version (newer versions even if lower take precedent)
		// todo(mustafa): Maybe we should add a config option later?
		err = tx.DeactivateProjectPackageVersionByPackageIdAndProjectId(pkg.ID.String(), project.ID.String())
		if err != nil {
			err = status.Errorf(codes.Internal, "could not deactivate package version: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}
	}

	for _, revision := range importRevisions {
		var packageVersionId string
		packageVersionId, err = tx.GetPackageVersionId(pkg.ID.String(), revision.Vre.Version.Value, revision.Vre.Release.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				packageVersionId, err = tx.CreatePackageVersion(pkg.ID.String(), revision.Vre.Version.Value, revision.Vre.Release.Value)
				if err != nil {
					err = status.Errorf(codes.Internal, "could not create package version: %v", err)
					setInternalError(errorDetails, err)
					return nil, err
				}
			} else {
				err = status.Errorf(codes.Internal, "could not get package version id: %v", err)
				setInternalError(errorDetails, err)
				return nil, err
			}
		}

		// todo(mustafa): Add published check, as well as limitations for overriding existing versions
		// TODO URGENT: Don't allow nondeterministic behavior regarding versions
		err = tx.AttachPackageVersion(project.ID.String(), pkg.ID.String(), packageVersionId, !req.SetInactive)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not attach package version: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}

		_, err = tx.CreateImportRevision(imp.ID.String(), revision.ScmHash, revision.ScmBranchName, revision.ScmUrl, packageVersionId, revision.Module)
		if err != nil {
			err = status.Errorf(codes.Internal, "could not create import revision: %v", err)
			setInternalError(errorDetails, err)
			return nil, err
		}
	}

	err = beginTx.Commit()
	if err != nil {
		err = status.Errorf(codes.Internal, "could not commit transaction: %v", err)
		setInternalError(errorDetails, err)
		return nil, err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED
	importPackageTask = peridotpb.ImportPackageTask{
		ImportId:        imp.ID.String(),
		PackageName:     pkg.Name,
		ImportRevisions: importRevisions,
	}

	return &importPackageTask, nil
}

func (c *Controller) PackageSrcGitActivity(ctx context.Context, packageName string, project *models.Project, parentTaskId string) (*peridotpb.PackageSrcGitResponse, error) {
	stopChan := makeHeartbeat(ctx, 4*time.Second)
	defer func() { stopChan <- true }()

	task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_IMPORT_SRC_GIT, utils.StringP(project.ID.String()), &parentTaskId)
	if err != nil {
		return nil, err
	}

	err = c.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := c.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			c.log.Errorf("could not set task status in PackageSrcGitActivity: %v", err)
		}
	}()

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	authenticator, err := c.getAuthenticator(project.ID.String())
	if err != nil {
		return nil, err
	}

	targetScmUrl := GetTargetScmUrl(project, packageName, OpenPatchSrc)
	logrus.Infof("Packaging src git for package %s, targetScmUrl: %s", packageName, targetScmUrl)

	_, w, err := checkoutRepo(project, project.TargetBranchPrefix, targetScmUrl, authenticator, git.AllTags, false)
	if err != nil {
		return nil, fmt.Errorf("could not checkout repo: %v", err)
	}

	res := peridotpb.PackageSrcGitResponse{
		TaskId:     task.ID.String(),
		NameHashes: map[string]string{},
	}

	sourcesStat, err := w.Filesystem.Stat("SOURCES")
	if err != nil {
		if os.IsNotExist(err) {
			return &res, nil
		}

		return nil, err
	}

	if !sourcesStat.IsDir() {
		return nil, temporal.NewNonRetryableApplicationError("SOURCES should be a directory", "SOURCES_IS_FILE_SRC_GIT", nil)
	}

	logrus.Infof("Reading SOURCES directory for directories to package")
	ls, err := w.Filesystem.ReadDir("SOURCES")
	if err != nil {
		return nil, err
	}

	for _, elem := range ls {
		if !elem.IsDir() {
			continue
		}
		logrus.Infof("Found directory %s", elem.Name())

		var buf bytes.Buffer
		err = compressFolder(filepath.Join("SOURCES", elem.Name()), "SOURCES", &buf, w.Filesystem)
		if err != nil {
			return nil, err
		}
		tarBts := buf.Bytes()

		h := sha256.New()
		_, err = h.Write(buf.Bytes())
		if err != nil {
			return nil, err
		}
		sumBytes := h.Sum(nil)
		sum := hex.EncodeToString(sumBytes)
		name := fmt.Sprintf("%s.tar.gz", elem.Name())

		f, err := w.Filesystem.OpenFile(filepath.Join("SOURCES", name), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(f, bytes.NewReader(tarBts))
		if err != nil {
			return nil, err
		}

		_, err = c.storage.PutObjectBytes(sum, tarBts)
		if err != nil {
			return nil, err
		}

		logrus.Infof("Directory %s (packaged as %s) has checksum %s", elem.Name(), name, sum)

		res.NameHashes[name] = sum
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &res, nil
}

func (c *Controller) UpdateDistGitForSrcGitActivity(ctx context.Context, packageName string, project *models.Project, packageRes *peridotpb.PackageSrcGitResponse) (*peridotpb.ImportRevision, error) {
	stopChan := makeHeartbeat(ctx, 4*time.Second)
	defer func() { stopChan <- true }()

	task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_IMPORT_SRC_GIT_TO_DIST_GIT, utils.StringP(project.ID.String()), &packageRes.TaskId)
	if err != nil {
		return nil, err
	}

	deferTask, _, err := c.commonCreateTask(task, nil)
	defer deferTask()
	if err != nil {
		return nil, err
	}

	authenticator, err := c.getAuthenticator(project.ID.String())
	if err != nil {
		return nil, err
	}

	// Src-git repository should be available at this point
	_, srcW, err := checkoutRepo(project, project.TargetBranchPrefix, GetTargetScmUrl(project, packageName, OpenPatchSrc), authenticator, git.AllTags, false)
	if err != nil {
		return nil, err
	}

	ls, err := srcW.Filesystem.ReadDir("SOURCES")
	if err != nil {
		return nil, err
	}

	// List and iterate through the SOURCES directory and remove any directory
	// since those are stored as tar blobs
	for _, elem := range ls {
		if elem.IsDir() {
			err := recursiveRemove(filepath.Join("SOURCES", elem.Name()), srcW.Filesystem)
			if err != nil {
				return nil, err
			}
		}
	}
	_ = srcW.Filesystem.Remove(".gitlab-ci.yml")

	// Try checking out the dist-git repo
	createRepo := false
	targetScmUrl := GetTargetScmUrl(project, packageName, OpenPatchRpms)
	repo, _, err := checkoutRepo(project, project.TargetBranchPrefix, targetScmUrl, authenticator, git.NoTags, true)
	if err != nil {
		// Or create a new one if it doesn't exist already
		repo, err = git.Init(memory.NewStorage(), osfs.New(rpmbuild.GetCloneDirectory()))
		if err != nil {
			return nil, err
		}

		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{targetScmUrl},
		})
		if err != nil {
			return nil, fmt.Errorf("could not create remote: %v", err)
		}

		// Create a new pointing to `sourceBranchPrefix``majorVersion` - r8 for Rocky for example
		refName := plumbing.NewBranchReferenceName(genPushBranch(project.TargetBranchPrefix, project.BranchSuffix.String, project.MajorVersion))
		h := plumbing.NewSymbolicReference(plumbing.HEAD, refName)
		if err := repo.Storer.CheckAndSetReference(h, nil); err != nil {
			return nil, fmt.Errorf("could not set reference: %v", err)
		}

		createRepo = true
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	// Get dist-git work tree and recursively remove older content
	err = recursiveRemove(".", w.Filesystem)
	if err != nil {
		return nil, err
	}
	// Copy src-git over to dist-git
	err = data.CopyFromFs(srcW.Filesystem, w.Filesystem, ".")
	if err != nil {
		return nil, err
	}
	_, err = w.Add(".")
	if err != nil {
		return nil, fmt.Errorf("could not add all files: %v", err)
	}

	// Build SRPM to get version and release
	err = c.setYumConfig(project)
	if err != nil {
		return nil, err
	}

	err = c.setBuildMacros(project, nil)
	if err != nil {
		return nil, fmt.Errorf("could not set build macros: %v", err)
	}

	err = srpmproc.Fetch(os.Stdout, "", rpmbuild.GetCloneDirectory(), osfs.New("/"), c.storage)
	if err != nil {
		return nil, fmt.Errorf("could not import using srpmproc: %v", err)
	}

	specFilePath, err := findSpec()
	if err != nil {
		return nil, fmt.Errorf("could not find spec file: %v", err)
	}
	logrus.Infof("Using spec: %s", specFilePath)

	pkgEo, err := c.db.GetExtraOptionsForPackage(project.ID.String(), packageName)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var options rpmbuild.Options
	if pkgEo != nil {
		options.With = pkgEo.WithFlags
		options.Without = pkgEo.WithoutFlags
	}
	err = c.rpmbuild.Exec(rpmbuild.ModeBuildSRPM|rpmbuild.ModeNoDeps|rpmbuild.ModePrivileged, "", specFilePath, options)
	if err != nil {
		return nil, fmt.Errorf("could not build srpm: %v", err)
	}

	srpmPath, err := findSrpm()
	if err != nil {
		return nil, fmt.Errorf("could not find srpm: %v", err)
	}
	srpmFile := filepath.Base(srpmPath)
	if !rpmutils.NVR().MatchString(srpmFile) {
		return nil, fmt.Errorf("invalid srpm file: %s", srpmFile)
	}
	nvrMatch := rpmutils.NVR().FindStringSubmatch(srpmFile)

	version := nvrMatch[2]
	release := nvrMatch[3]
	logrus.Infof("Version: %s, Release: %s", version, release)

	newTag := fmt.Sprintf("imports/%s%d%s/%s-%s-%s", project.TargetBranchPrefix, project.MajorVersion, project.BranchSuffix.String, packageName, version, release)
	pushBranch := genPushBranch(project.TargetBranchPrefix, project.BranchSuffix.String, project.MajorVersion)

	if !createRepo {
		tags, err := repo.TagObjects()
		if err != nil {
			return nil, err
		}
		var ir *peridotpb.ImportRevision
		err = tags.ForEach(func(t *object.Tag) error {
			if t.Name == newTag {
				task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

				commit, err := t.Commit()
				if err != nil {
					return err
				}

				ir = &peridotpb.ImportRevision{
					ScmHash:       commit.Hash.String(),
					ScmBranchName: pushBranch,
					ScmUrl:        targetScmUrl,
					Vre: &peridotpb.VersionRelease{
						Version: wrapperspb.String(version),
						Release: wrapperspb.String(release),
					},
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
		if ir != nil {
			return ir, nil
		}
	}

	// Add blobs to the metadata file
	// These can later be fetched with `srpmproc fetch`
	metadataFile := fmt.Sprintf(".%s.metadata", packageName)
	metadata, err := w.Filesystem.OpenFile(metadataFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not create metadata file: %v", err)
	}
	for name, hash := range packageRes.NameHashes {
		checksumLine := fmt.Sprintf("%s %s\n", hash, filepath.Join("SOURCES", name))
		_, err = metadata.Write([]byte(checksumLine))
		if err != nil {
			return nil, fmt.Errorf("could not write to metadata file: %v", err)
		}
	}

	_, err = w.Add(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("could not add %s: %v", metadataFile, err)
	}

	s, err := w.Status()
	if err != nil {
		return nil, err
	}

	statusLines := strings.Split(s.String(), "\n")
	for _, line := range statusLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "D") {
			path := strings.TrimPrefix(trimmed, "D ")
			_, err := w.Remove(path)
			if err != nil {
				return nil, fmt.Errorf("could not delete extra file %s: %v", path, err)
			}
		}
	}

	var hashes []plumbing.Hash
	var pushRefspecs []config.RefSpec

	head, err := repo.Head()
	if err != nil {
		// if no remote origin exists, just push all
		hashes = nil
		pushRefspecs = append(pushRefspecs, "*:*")
	} else {
		// push to the cloned remote origin only
		hashes = append(hashes, head.Hash())
		refOrigin := "refs/heads/" + pushBranch
		pushRefspecs = append(pushRefspecs, config.RefSpec(fmt.Sprintf("HEAD:%s", refOrigin)))
	}

	commitSig := &object.Signature{
		Name:  "Peridot Bot",
		Email: "rockyautomation@rockylinux.org",
		When:  time.Now(),
	}

	// we are now finished with the tree and are going to push it to the dist-git repo
	// create import commit
	// todo(mustafa): Add customization options for name and email
	commit, err := w.Commit(fmt.Sprintf("import %s", newTag), &git.CommitOptions{
		Author:  commitSig,
		Parents: hashes,
	})
	if err != nil {
		return nil, fmt.Errorf("could not commit object: %v", err)
	}

	_, err = repo.CommitObject(commit)
	if err != nil {
		return nil, fmt.Errorf("could not get commit object: %v", err)
	}

	_, err = repo.CreateTag(newTag, commit, &git.CreateTagOptions{
		Tagger:  commitSig,
		Message: "sync from src-git to dist-git",
	})
	if err != nil {
		return nil, fmt.Errorf("could not create tag: %v", err)
	}

	if createRepo {
		err := c.createProjectOrMakePublic(project, packageName, OpenPatchRpms)
		if err != nil {
			return nil, err
		}
	}

	// Push to Gitlab using HTTP-auth
	pushRefspecs = append(pushRefspecs, config.RefSpec("HEAD:"+plumbing.NewTagReferenceName(newTag)))
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       authenticator,
		RefSpecs:   pushRefspecs,
		Force:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not push to remote: %v", err)
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &peridotpb.ImportRevision{
		ScmHash:       commit.String(),
		ScmBranchName: pushBranch,
		ScmUrl:        targetScmUrl,
		Vre: &peridotpb.VersionRelease{
			Version: wrapperspb.String(version),
			Release: wrapperspb.String(release),
		},
	}, nil
}

// srpmprocToImportRevisions translates the response from srpmproc to peridotpb.ImportRevision
func (c *Controller) srpmprocToImportRevisions(project *models.Project, pkg string, res *srpmprocpb.ProcessResponse, module bool) []*peridotpb.ImportRevision {
	var importRevisions []*peridotpb.ImportRevision
	for branch, commit := range res.BranchCommits {
		moduleStream := false
		if !module && strings.Contains(branch, "-stream-") {
			moduleStream = true
		}

		version, ok := res.BranchVersions[branch]

		if !ok {
			c.log.Errorf("unable to find branch %s in BranchVersions", branch)
		}

		if version == nil {
			c.log.Errorf("version for branch %s is nil", branch)
		}

		// For now let's just include all module metadata in the release field.
		// We might use it to match upstream versions in the Future.
		// If it doesn't work out as expected, we can always resort back to replacing.
		release := version.Release
		if release == "" {
			c.log.Errorf("release information missing for branch %s", branch)
		}

		section := OpenPatchRpms
		if module {
			section = OpenPatchModules
		}
		targetScmUrl := GetTargetScmUrl(project, pkg, section)

		importRevision := &peridotpb.ImportRevision{
			ScmHash:       commit,
			ScmBranchName: branch,
			ScmUrl:        targetScmUrl,
			Vre: &peridotpb.VersionRelease{
				Version: wrapperspb.String(version.Version),
				Release: wrapperspb.String(release),
			},
			Module:       module,
			ModuleStream: moduleStream,
		}
		importRevisions = append(importRevisions, importRevision)
	}

	return importRevisions
}

// UpstreamDistGitActivity imports from a "source of truth" and applies downstream patches.
// Matches current Rocky workflow with distrobuild+srpmrpoc.
// This activity also uses srpmproc as a library instead of a CLI tool
func (c *Controller) UpstreamDistGitActivity(ctx context.Context, greq *UpstreamDistGitActivityRequest) (*UpstreamDistGitActivityResponse, error) {
	stopChan := makeHeartbeat(ctx, 4*time.Second)
	defer func() { stopChan <- true }()

	project := greq.Project
	parentTaskId := greq.ParentTaskId

	// Create subtask
	task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_IMPORT_UPSTREAM, utils.StringP(project.ID.String()), &parentTaskId)
	if err != nil {
		return nil, err
	}

	err = c.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		return nil, err
	}

	// Set task status
	defer func() {
		err := c.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			c.log.Errorf("could not set task status in UpdateDistGitForSrcGitActivity: %v", err)
		}
	}()

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	if !project.SourceGitHost.Valid || !project.SourcePrefix.Valid || !project.SourceBranchPrefix.Valid {
		return nil, status.Error(codes.FailedPrecondition, "no upstream info provided")
	}

	packageType := greq.Package.PackageType
	if greq.Package.PackageTypeOverride.Valid {
		packageType = peridotpb.PackageType(greq.Package.PackageTypeOverride.Int32)
	}

	moduleMode := false
	switch packageType {
	case peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK:
		moduleMode = true
	}

	// Retrieve keys for the project
	projectKeys, err := c.db.GetProjectKeys(project.ID.String())
	if err != nil {
		return nil, err
	}

	gitHost := project.SourceGitHost.String
	sourcePrefix := project.SourcePrefix.String
	branchPrefix := project.SourceBranchPrefix.String
	if packageType == peridotpb.PackageType_PACKAGE_TYPE_NORMAL {
		gitHost = project.TargetGitlabHost
		sourcePrefix = project.TargetPrefix
		branchPrefix = project.TargetBranchPrefix
	}

	modulePrefix := fmt.Sprintf("%s/%s/modules", gitHost, sourcePrefix)
	rpmPrefix := fmt.Sprintf("%s/%s/rpms", gitHost, sourcePrefix)
	var packageVersion string
	var packageRelease string

	// Modules do not have a persistent VRE, so don't set it when in module mode
	if greq.VersionRelease != nil && !moduleMode {
		packageVersion = greq.VersionRelease.Version.Value
		packageRelease = greq.VersionRelease.Release.Value
	}

	// Use the helper PD tool from srpmproc to create pre-run metadata
	pd, err := srpmproc.NewProcessData(&srpmproc.ProcessDataRequest{
		Version:            project.MajorVersion,
		StorageAddr:        fmt.Sprintf("s3://%s", viper.GetString("s3-bucket")),
		Package:            greq.Package.Name,
		PackageGitName:     gitlabify(greq.Package.Name),
		ModulePrefix:       modulePrefix,
		RpmPrefix:          rpmPrefix,
		HttpUsername:       projectKeys.GitlabUsername,
		HttpPassword:       projectKeys.GitlabSecret,
		UpstreamPrefix:     fmt.Sprintf("%s/%s", project.TargetGitlabHost, project.TargetPrefix),
		GitCommitterName:   "Peridot Bot",
		GitCommitterEmail:  "rockyautomation@rockylinux.org",
		ImportBranchPrefix: branchPrefix,
		BranchPrefix:       project.TargetBranchPrefix,
		BranchSuffix:       project.BranchSuffix.String,
		StrictBranchMode:   true,
		ModuleMode:         moduleMode,
		CdnUrl:             project.CdnUrl.String,
		PackageVersion:     packageVersion,
		PackageRelease:     packageRelease,
	})
	if err != nil {
		return nil, err
	}

	// Invoke srpmproc, this will push to the project target gitlab
	res, err := srpmproc.ProcessRPM(&*pd)
	if err != nil {
		return nil, err
	}

	// Make project public if enabled in settings
	section := OpenPatchRpms
	if moduleMode {
		section = OpenPatchModules
	}
	err = c.createProjectOrMakePublic(project, greq.Package.Name, section)
	if err != nil {
		return nil, err
	}

	importRevisions := c.srpmprocToImportRevisions(project, greq.Package.Name, res, moduleMode)

	// If the package is both a module and a package, we need to import the module too
	if packageType == peridotpb.PackageType_PACKAGE_TYPE_NORMAL_FORK_MODULE || packageType == peridotpb.PackageType_PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT {
		pd2, err := srpmproc.NewProcessData(&srpmproc.ProcessDataRequest{
			Version:            project.MajorVersion,
			StorageAddr:        fmt.Sprintf("s3://%s", viper.GetString("s3-bucket")),
			Package:            greq.Package.Name,
			PackageGitName:     gitlabify(greq.Package.Name),
			ModulePrefix:       modulePrefix,
			RpmPrefix:          rpmPrefix,
			HttpUsername:       projectKeys.GitlabUsername,
			HttpPassword:       projectKeys.GitlabSecret,
			UpstreamPrefix:     fmt.Sprintf("%s/%s", project.TargetGitlabHost, project.TargetPrefix),
			GitCommitterName:   "Peridot Bot",
			GitCommitterEmail:  "rockyautomation@rockylinux.org",
			ImportBranchPrefix: branchPrefix,
			BranchPrefix:       project.TargetBranchPrefix,
			BranchSuffix:       project.BranchSuffix.String,
			StrictBranchMode:   true,
			ModuleMode:         true,
			CdnUrl:             project.CdnUrl.String,
		})
		if err != nil {
			return nil, err
		}
		mRes, err := srpmproc.ProcessRPM(pd2)
		if err != nil {
			return nil, err
		}
		importRevisions = append(importRevisions, c.srpmprocToImportRevisions(project, greq.Package.Name, mRes, true)...)

		err = c.createProjectOrMakePublic(project, greq.Package.Name, OpenPatchModules)
		if err != nil {
			return nil, err
		}
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &UpstreamDistGitActivityResponse{
		ImportRevisions: importRevisions,
	}, nil
}

// ImportSetGitlabStatusActivity sets a successful import status on the commit for traceability
// Later a build will also set a success/failed status on the commit (if it's queued for build)
func (c *Controller) ImportSetGitlabStatusActivity(packageName string, project *models.Project, task *models.Task, section OpenPatchSection, shas []string, state gitlab.BuildStateValue) error {
	gitlabClient, err := c.getGitlabClient(project)
	if err != nil {
		return err
	}

	projectName := fmt.Sprintf("%s/%s/%s", project.TargetPrefix, section, packageName)

	for _, sha := range shas {
		_, _, err := gitlabClient.Commits.SetCommitStatus(projectName, sha, &gitlab.SetCommitStatusOptions{
			State: state,
			Name:  gitlab.String("peridot-import"),
			// todo(mustafa): Do not hardcode rockylinux.org here
			TargetURL:   gitlab.String(fmt.Sprintf("https://peridot.rockylinux.org/%s/tasks/%s", project.ID.String(), task.ID.String())),
			Description: gitlab.String("Peridot Import"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
