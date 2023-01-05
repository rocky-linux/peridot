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

package serverdb

import (
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/anypb"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

type Access interface {
	ListProjects(filters *peridotpb.ProjectFilters) (models.Projects, error)
	GetProjectKeys(projectId string) (*models.ProjectKey, error)
	GetProjectModuleConfiguration(projectId string) (*peridotpb.ModuleConfiguration, error)
	CreateProjectModuleConfiguration(projectId string, config *peridotpb.ModuleConfiguration) error
	CreateProject(project *peridotpb.Project) (*models.Project, error)
	UpdateProject(id string, project *peridotpb.Project) (*models.Project, error)
	SetProjectKeys(projectId string, username string, password string) error
	SetBuildRootPackages(projectId string, srpmPackages pq.StringArray, buildPackages pq.StringArray) error

	CreateBuild(packageId string, packageVersionId string, taskId string, projectId string) (*models.Build, error)
	GetArtifactsForBuild(buildId string) (models.TaskArtifacts, error)
	GetBuildCount() (int64, error)
	CreateBuildBatch(projectId string) (string, error)
	AttachBuildToBatch(buildId string, batchId string) error
	ListBuilds(filters *peridotpb.BuildFilters, projectId string, page int32, limit int32) (models.Builds, error)
	GetSuccessfulBuildIDsAsc(projectId string) ([]string, error)
	BuildCountInProject(projectId string) (int64, error)
	GetBuild(projectId string, buildId string) (*models.Build, error)
	GetBuildByID(buildId string) (*models.Build, error)
	GetBuildByTaskIdAndPackageId(taskId string, packageId string) (*models.Build, error)
	GetBuildBatch(projectId string, batchId string, batchFilter *peridotpb.BatchFilter, page int32, limit int32) (models.Builds, error)
	ListBuildBatches(projectId string, batchId *string, page int32, limit int32) (models.BuildBatches, error)
	BuildBatchCountInProject(projectId string) (int64, error)
	BuildsInBatchCount(projectId string, batchId string) (int64, error)
	LockNVRA(nvra string) error
	UnlockNVRA(nvra string) error
	NVRAExists(nvra string) (bool, error)
	GetBuildByPackageNameAndVersionAndRelease(name string, version string, release string) (models.Builds, error)
	GetLatestBuildIdsByPackageName(name string, projectId *string) ([]string, error)
	GetBuildIDsByPackageNameAndBranchName(name string, branchName string) ([]string, error)
	GetActiveBuildIdsByTaskArtifactGlob(taskArtifactGlob string, projectId string) ([]string, error)
	GetAllBuildIdsByPackageName(name string, projectId string) ([]string, error)

	CreateImport(scmUrl string, taskId string, packageId string, projectId string) (*models.Import, error)
	CreateImportRevision(importId string, scmHash string, scmBranchName string, scmUrl string, packageVersionId string, modular bool) (*models.ImportRevision, error)
	GetLatestImportRevisionsForPackageInProject(packageName string, projectId string) (models.ImportRevisions, error)
	GetImportRevisionByScmHash(scmHash string) (*models.ImportRevision, error)
	DeactivateImportRevisionsByPackageVersionId(packageVersionId string) error
	CreateImportBatch(projectId string) (string, error)
	AttachImportToBatch(importId string, batchId string) error
	ListImports(projectId string, page int32, limit int32) (models.Imports, error)
	ImportCountInProject(projectId string) (int64, error)
	GetImport(projectId string, importId string) (*models.Import, error)
	GetImportBatch(projectId string, batchId string, batchFilter *peridotpb.BatchFilter, page int32, limit int32) (models.Imports, error)
	ListImportBatches(projectId string, batchId *string, page int32, limit int32) (models.ImportBatches, error)
	ImportBatchCountInProject(projectId string) (int64, error)
	ImportsInBatchCount(projectId string, batchId string) (int64, error)

	GetPackagesInProject(filters *peridotpb.PackageFilters, projectId string, page int32, limit int32) (models.Packages, error)
	PackageCountInProject(projectId string) (int64, error)
	GetPackageVersion(packageVersionId string) (*models.PackageVersion, error)
	GetPackageVersionId(packageId string, version string, release string) (string, error)
	CreatePackageVersion(packageId string, version string, release string) (string, error)
	AttachPackageVersion(projectId string, packageId string, packageVersionId string, active bool) error
	GetProjectPackageVersionFromPackageVersionId(packageVersionId string, projectId string) (string, error)
	DeactivateProjectPackageVersionByPackageIdAndProjectId(packageId string, projectId string) error
	MakeActiveInRepoForPackageVersion(packageVersionId string, packageId string, projectId string) error
	CreatePackage(name string, packageType peridotpb.PackageType) (*models.Package, error)
	AddPackageToProject(projectId string, packageId string, packageTypeOverride peridotpb.PackageType) error
	GetPackageID(name string) (string, error)
	SetExtraOptionsForPackage(projectId string, packageName string, withFlags pq.StringArray, withoutFlags pq.StringArray) error
	GetExtraOptionsForPackage(projectId string, packageName string) (*models.ExtraOptions, error)
	SetGroupInstallOptionsForPackage(projectId string, packageName string, dependsOn pq.StringArray, enableModule pq.StringArray, disableModule pq.StringArray) error
	SetPackageType(projectId string, packageName string, packageType peridotpb.PackageType) error

	CreateTask(user *utils.ContextUser, arch string, taskType peridotpb.TaskType, projectId *string, parentTaskId *string) (*models.Task, error)
	SetTaskStatus(id string, status peridotpb.TaskStatus) error
	SetTaskResponse(id string, response *anypb.Any) error
	SetTaskMetadata(id string, metadata *anypb.Any) error
	// ListTasks returns only parent tasks
	ListTasks(projectId *string, page int32, limit int32) (models.Tasks, error)
	// GetTask returns a parent task as well as all it's child tasks
	GetTask(id string, projectId *string) (models.Tasks, error)
	// GetTaskByBuildId returns the task of a build (only parent task)
	GetTaskByBuildId(buildId string) (*models.Task, error)
	AttachTaskToBuild(buildId string, taskId string) error
	AttachArtifactToTask(objectName string, hashSha256 string, arch string, metadata *anypb.Any, taskId string) error
	CreateTaskArtifactSignature(taskArtifactId string, keyId string, hashSha256 string) error
	GetTaskArtifactSignatureHash(name string, gpgKeyId string) (string, error)
	TaskCountInProject(projectId string) (int64, error)
	GetTaskArtifactById(taskArtifactId string) (*models.TaskArtifact, error)
	GetTaskStatus(id string) (peridotpb.TaskStatus, error)

	GetPluginsForProject(projectId string) (models.Plugins, error)

	GetExternalRepositoriesForProject(projectId string) (models.ExternalRepositories, error)
	DeleteExternalRepositoryForProject(projectId string, externalRepositoryId string) error
	CreateExternalRepositoryForProject(projectId string, repoURL string, priority *int32, moduleHotfixes bool) (*models.ExternalRepository, error)
	FindRepositoriesForPackage(projectId string, pkg string, internalOnly bool) (models.Repositories, error)
	FindRepositoriesForProject(projectId string, id *string, internalOnly bool) (models.Repositories, error)
	GetRepositoryRevision(revisionId string) (*models.RepositoryRevision, error)
	GetLatestActiveRepositoryRevision(repoId string, arch string) (*models.RepositoryRevision, error)
	GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(projectId string, name string, arch string) (*models.RepositoryRevision, error)
	CreateRevisionForRepository(id string, repoId string, arch string, repomdXml string, primaryXml string, filelistsXml string, otherXml string, updateInfoXml string, moduleDefaultsYaml string, modulesYaml string, groupsXml string, urlMappings string) (*models.RepositoryRevision, error)
	CreateRepositoryWithPackages(name string, projectId string, internalOnly bool, packages pq.StringArray) (*models.Repository, error)
	GetRepository(id *string, name *string, projectId *string) (*models.Repository, error)
	SetRepositoryOptions(id string, packages pq.StringArray, excludeFilter pq.StringArray, includeFilter pq.StringArray, additionalMultilib pq.StringArray, excludeMultilibFilter pq.StringArray, multilib pq.StringArray, globIncludeFilter pq.StringArray) error

	CreateKey(id string, name string, email string, gpgId string, encKey string, nonce string, publicKey string, extStoreType string, extStoreId string) (*models.Key, error)
	AttachKeyToProject(projectId string, keyId string, defaultKey bool) error
	GetKeyByProjectIdAndId(projectId string, keyId string) (*models.Key, error)
	GetDefaultKeyForProject(projectId string) (*models.Key, error)
	GetKeyByName(name string) (*models.Key, error)

	InsertLogs(lines pq.StringArray, taskId string, parentTaskId string) error
	GetLogsForTaskIdOrParentTaskId(taskId *string, parentTaskId *string, offset *int64) ([]pq.StringArray, error)

	Begin() (utils.Tx, error)
	UseTransaction(tx utils.Tx) Access
}
