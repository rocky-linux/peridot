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
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cavaliergopher/rpm"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rocky-linux/srpmproc/pkg/srpmproc"
	"go.temporal.io/sdk/activity"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/rpmbuild"
)

func gitlabify(str string) string {
	if str == "tree" {
		return "treepkg"
	}

	return strings.Replace(str, "+", "plus", -1)
}

func findSpec() (string, error) {
	var specFilePath string
	err := filepath.Walk(filepath.Join(rpmbuild.GetCloneDirectory(), "SPECS"), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(filepath.Base(path), ".spec") {
			specFilePath = path
			return nil
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	if specFilePath == "" {
		return "", fmt.Errorf("could not find a valid spec file")
	}

	return specFilePath, nil
}

func findSrpm() (string, error) {
	var srpmFilePath string
	err := filepath.Walk(rpmbuild.GetCloneDirectory()+"/SRPMS", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(filepath.Base(path), ".src.rpm") {
			srpmFilePath = path
			return nil
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	if srpmFilePath == "" {
		return "", fmt.Errorf("could not find a valid srpm file")
	}

	return srpmFilePath, nil
}

func decompressGz(path string) ([]byte, error) {
	if filepath.Ext(path) != ".gz" {
		return nil, errors.New("gz file must end with .gz")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Controller) uploadArtifact(projectId string, parentTaskId string, filePath string, arch string, taskType peridotpb.TaskType) (*UploadActivityResult, error) {
	task, err := c.db.CreateTask(nil, "noarch", taskType, &projectId, &parentTaskId)
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
			c.log.Errorf("could not set task status in uploadArtifact: %v", err)
		}
	}()

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	// if the artifact is a rpm, generate repo and save metadata
	var metadata *anypb.Any

	objectName := filepath.Join(parentTaskId, filepath.Base(filePath))
	f, err := os.Open(filePath)
	if err != nil {
		_ = c.logToMon([]string{fmt.Sprintf("could not open file %s: %v", filePath, err)}, task.ID.String(), parentTaskId)
		return nil, fmt.Errorf("could not open file: %v", err)
	}

	hasher := sha256.New()
	buf := make([]byte, 1024*1024)
	for {
		bytesRead, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not read file: %v", err)
		}

		hasher.Write(buf[:bytesRead])
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	exists, err := c.storage.Exists(objectName)
	if exists || err == nil {
		_ = c.logToMon([]string{fmt.Sprintf("skipping upload of %s, already exists", objectName)}, task.ID.String(), parentTaskId)
		task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

		return &UploadActivityResult{
			ObjectName: objectName,
			Subtask:    task,
			HashSha256: hash,
			Arch:       arch,
			Skip:       true,
		}, nil
	}

	if filepath.Ext(filePath) == ".rpm" {
		tmp, err := os.MkdirTemp("", "")
		if err != nil {
			return nil, fmt.Errorf("could not create temp dir: %v", err)
		}
		err = os.Link(filePath, filepath.Join(tmp, filepath.Base(filePath)))
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not link file: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not link %s to %s: %v", filePath, filepath.Join(tmp, filepath.Base(filePath)), err)
		}
		err = runCmd("createrepo_c", "--basedir="+tmp, tmp)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not create repo: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not create repo: %v", err)
		}

		var primaryGzPath string
		var filelistsGzPath string
		var otherGzPath string
		err = filepath.Walk(filepath.Join(tmp, "repodata"), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, "primary.xml.gz") {
				primaryGzPath = path
			} else if strings.HasSuffix(path, "filelists.xml.gz") {
				filelistsGzPath = path
			} else if strings.HasSuffix(path, "other.xml.gz") {
				otherGzPath = path
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		primaryXmlBytes, err := decompressGz(primaryGzPath)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not decompress primary.xml.gz: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not decompress primary.xml.gz: %v", err)
		}
		filelistsXmlBytes, err := decompressGz(filelistsGzPath)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not decompress filelists.xml.gz: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not decompress filelists.xml.gz: %v", err)
		}
		otherXmlBytes, err := decompressGz(otherGzPath)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not decompress other.xml.gz: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not decompress other.xml.gz: %v", err)
		}

		f, err := os.Open(filePath)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not open file: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not open file %s: %v", filePath, err)
		}
		rpmPkg, err := rpm.Read(f)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not read rpm: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not read rpm: %v", err)
		}
		// Use header tags in RPM headers directly
		// todo(mustafa): Abstract away
		// Source: https://github.com/rpm-software-management/rpm/blob/82dafa39a2dfd3e24858681ca75f467c1e1b3635/lib/rpmtag.h
		buildArch := rpmPkg.Header.GetTag(1089).StringSlice()
		excludeArch := rpmPkg.Header.GetTag(1059).StringSlice()
		exclusiveArch := rpmPkg.Header.GetTag(1061).StringSlice()

		rpmMetadata := &peridotpb.RpmArtifactMetadata{
			Primary:       primaryXmlBytes,
			Filelists:     filelistsXmlBytes,
			Other:         otherXmlBytes,
			ExcludeArch:   excludeArch,
			ExclusiveArch: exclusiveArch,
			BuildArch:     buildArch,
		}
		metadata, err = anypb.New(rpmMetadata)
		if err != nil {
			_ = c.logToMon([]string{fmt.Sprintf("could not create metadata: %v", err)}, task.ID.String(), parentTaskId)
			return nil, fmt.Errorf("could not create metadata: %v", err)
		}
	}

	_, err = c.storage.PutObject(objectName, filePath)
	if err != nil {
		_ = c.logToMon([]string{fmt.Sprintf("could not upload file %s: %v", filePath, err)}, task.ID.String(), parentTaskId)
		return nil, fmt.Errorf("could not upload artifact: %v", err)
	}
	_ = c.logToMon(
		[]string{fmt.Sprintf("uploaded %s to blob storage", filePath)},
		task.ID.String(),
		parentTaskId,
	)

	err = c.db.AttachArtifactToTask(objectName, hash, arch, metadata, task.ID.String())
	if err != nil {
		_ = c.logToMon([]string{fmt.Sprintf("could not attach artifact to task: %v", err)}, task.ID.String(), parentTaskId)
		return nil, fmt.Errorf("could not attach artifact to task: %v", err)
	}
	err = c.db.SetTaskMetadata(task.ID.String(), metadata)
	if err != nil {
		_ = c.logToMon([]string{fmt.Sprintf("could not set task metadata: %v", err)}, task.ID.String(), parentTaskId)
		return nil, fmt.Errorf("could not set task metadata: %v", err)
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return &UploadActivityResult{
		ObjectName: objectName,
		Subtask:    task,
		HashSha256: hash,
		Arch:       arch,
	}, nil
}

func (c *Controller) BuildSRPMActivity(ctx context.Context, upstreamPrefix string, scmHash string, projectId string, packageName string, packageVersion *models.PackageVersion, task *models.Task, extraOptions *peridotpb.ExtraBuildOptions) error {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(30 * time.Second)
		}
	}()

	err := c.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		return err
	}

	defer func() {
		err := c.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			c.log.Errorf("could not set task status in BuildSRPMActivity: %v", err)
		}
	}()

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{Id: wrapperspb.String(projectId)})
	if err != nil {
		return err
	}
	project := projects[0]

	pkgEo, err := c.db.GetExtraOptionsForPackage(project.ID.String(), packageName)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	authenticator, _ := c.getAuthenticator(projectId)
	repoUrl := fmt.Sprintf("%s/rpms/%s.git", upstreamPrefix, gitlabify(packageName))
	r, err := git.PlainClone(rpmbuild.GetCloneDirectory(), false, &git.CloneOptions{
		Auth: authenticator,
		URL:  repoUrl,
		Tags: git.AllTags,
	})
	if err != nil {
		return fmt.Errorf("could not clone rpmbuild repo %s: %v", repoUrl, err)
	}

	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"+refs/heads/*:refs/remotes/*"},
		Auth:     authenticator,
		Tags:     git.AllTags,
		Force:    true,
	})
	if err != nil {
		return fmt.Errorf("could not fetch rpmbuild repo: %v", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("could not get worktree: %v", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(scmHash),
		Force: true,
		Keep:  false,
	})
	if err != nil {
		return fmt.Errorf("could not checkout %s: %v", scmHash, err)
	}

	cloneDir := rpmbuild.GetCloneDirectory()
	err = os.MkdirAll(filepath.Join(cloneDir, "SRPMS"), 0755)
	if err != nil {
		return err
	}

	err = srpmproc.Fetch(os.Stdout, "", cloneDir, osfs.New("/"), c.storage)
	if err != nil {
		return fmt.Errorf("could not import using srpmproc: %v", err)
	}

	err = runCmd("chown", "-R", "peridotbuilder:mock", cloneDir)
	if err != nil {
		return fmt.Errorf("could not chown clone dir: %v", err)
	}

	specFilePath, err := findSpec()
	if err != nil {
		return fmt.Errorf("could not find spec file: %v", err)
	}

	var pkgGroup = DefaultSrpmBuildPkgGroup

	if len(project.SrpmStagePackages) != 0 {
		pkgGroup = project.SrpmStagePackages
	}

	var enableModules []string
	var disableModules []string
	err = ParsePackageExtraOptions(pkgEo, &pkgGroup, &enableModules, &disableModules)

	if err != nil {
		c.log.Infof("no extra options to process for package")
	}

	extraOptions.DisabledModules = disableModules
	extraOptions.Modules = enableModules

	hostArch := os.Getenv("REAL_BUILD_ARCH")
	extraOptions.EnableNetworking = true
	err = c.writeMockConfig(&project, packageVersion, extraOptions, "noarch", hostArch, pkgGroup)
	if err != nil {
		return fmt.Errorf("could not write mock config: %v", err)
	}
	// The SOURCES dir should always be available. Some packages don't have that
	// and Mock complains. Loudly. About that
	_ = os.MkdirAll(filepath.Join(cloneDir, "SOURCES"), 0644)

	args := []string{
		"mock",
		"--isolation=simple",
		"-r",
		"/var/peridot/mock.cfg",
		"--target",
		"noarch",
		"--resultdir",
		filepath.Join(cloneDir, "SRPMS"),
		"--sources",
		filepath.Join(cloneDir, "SOURCES"),
	}
	if pkgEo != nil {
		for _, with := range pkgEo.WithFlags {
			args = append(args, "--with="+with)
		}
		for _, without := range pkgEo.WithoutFlags {
			args = append(args, "--without="+without)
		}
	}
	args = append(args, []string{"--buildsrpm", "--spec", specFilePath}...)

	cmd := exec.Command("/bundle/fork-exec.py", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not mock build: %v", err)
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return nil
}

func ParsePackageExtraOptions(pkgEo *models.ExtraOptions, pkgGroup *[]string, enableModules *[]string, disableModules *[]string) error {

	if pkgEo == nil {
		return fmt.Errorf("no extra options to parse for package")
	}

	if len(pkgEo.DependsOn) != 0 {
		for _, pkg := range pkgEo.DependsOn {
			*pkgGroup = append(*pkgGroup, pkg)
		}
	}

	if len(pkgEo.EnableModule) != 0 {
		for _, pkg := range pkgEo.EnableModule {
			*enableModules = append(*enableModules, pkg)
		}
	}

	if len(pkgEo.DisableModule) != 0 {
		for _, pkg := range pkgEo.DisableModule {
			*disableModules = append(*disableModules, pkg)
		}
	}
	return nil
}

type UploadActivityResult struct {
	ObjectName string       `json:"objectName"`
	Subtask    *models.Task `json:"subtask"`
	HashSha256 string       `json:"hashSha256"`
	Arch       string       `json:"arch"`
	Skip       bool         `json:"skip"`
}

func (c *Controller) UploadSRPMActivity(ctx context.Context, projectId string, parentTaskId string) (*UploadActivityResult, error) {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(4 * time.Second)
		}
	}()

	srpmFilePath, err := findSrpm()
	if err != nil {
		return nil, err
	}

	return c.uploadArtifact(projectId, parentTaskId, srpmFilePath, "src", peridotpb.TaskType_TASK_TYPE_BUILD_SRPM_UPLOAD)
}
