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
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/peridot/rpmbuild"
	"peridot.resf.org/servicecatalog"
)

var (
	releaseDistRegex = regexp.MustCompile(".+\\.(el[^. \\t\\n]+)")
	// defaults are for el9
	DefaultBuildPkgGroup = []string{
		"bash",
		"bzip2",
		"coreutils",
		"cpio",
		"diffutils",
		"findutils",
		"gawk",
		"glibc-minimal-langpack",
		"grep",
		"gzip",
		"info",
		"make",
		"patch",
		"rpm-build",
		"sed",
		"shadow-utils",
		"tar",
		"unzip",
		"util-linux",
		"which",
		"xz",
	}
	// defaults are for EL9
	DefaultSrpmBuildPkgGroup = []string{
		"bash",
		"glibc-minimal-langpack",
		"gnupg2",
		"rpm-build",
		"shadow-utils",
	}
)

func runCmd(command string, args ...string) error {
	fmt.Printf("[+] %s %s\n", command, strings.Join(args, " "))

	var newArgs []string
	for _, arg := range args {
		newArgs = append(newArgs, strings.ReplaceAll(arg, "\"", ""))
	}

	cmd := exec.Command(command, newArgs...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(stdout io.ReadCloser) {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}(stdoutPipe)
	go func(stderr io.ReadCloser) {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}(stderrPipe)

	wg.Wait()

	return cmd.Wait()
}

func findRpms() ([]string, error) {
	var rpms []string
	err := filepath.Walk(rpmbuild.GetCloneDirectory()+"/RPMS", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(filepath.Base(path), ".rpm") && !strings.HasSuffix(filepath.Base(path), ".src.rpm") {
			rpms = append(rpms, path)
			return nil
		}

		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return rpms, nil
		}
		return nil, err
	}

	return rpms, nil
}

func addExtraFiles(extraOptions *peridotpb.ExtraBuildOptions) error {
	if extraOptions.BuildArchExtraFiles != nil {
		for k, v := range extraOptions.BuildArchExtraFiles {
			// Create dir if now exists
			dir := filepath.Dir(k)
			_, err := os.Stat(dir)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(dir, 0755)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			}
			err = ioutil.WriteFile(k, []byte(v), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getRepoUrl(arch string, nRepoUrl string) string {
	repoUrl := strings.NewReplacer("$arch", arch, "$basearch", arch).Replace(nRepoUrl)
	// Sometimes i686 may be called i386 for repo purposes
	// Koji uses i386 for example for repo path but builds for i686
	// Let's just make sure we can use the repo
	if arch == "i686" {
		res, err := http.Get(fmt.Sprintf("%s/repodata/repomd.xml", repoUrl))
		if err != nil || res.StatusCode != 200 {
			repoUrl = strings.NewReplacer("$arch", "i386", "$basearch", "i386").Replace(nRepoUrl)
		}
	} else if arch == "noarch" {
		repoUrl = nRepoUrl
	}

	return repoUrl
}

func (c *Controller) repos(projectId string, arch string, extraOptions *peridotpb.ExtraBuildOptions) (map[string]string, error) {
	ret := map[string]string{}
	extraRepos := extraOptions.ExtraYumrepofsRepos

	yumrepofsRepos := extraRepos
	hasAll := false
	for _, repo := range yumrepofsRepos {
		if repo.Name == "all" {
			hasAll = true
		}
	}
	if !hasAll {
		yumrepofsRepos = []*peridotpb.ExtraYumrepofsRepo{
			{
				Name: "all",
			},
		}
		yumrepofsRepos = append(yumrepofsRepos, extraRepos...)
	}

	for i, repo := range yumrepofsRepos {
		repoUrl := getRepoUrl(arch, servicecatalog.YumrepofsRepo(projectId, repo.Name, "$arch"))

		yumrepofsConfig := `[yumrepofs_{i}]
name=Peridot Internal - Yumrepofs {i}
baseurl={url}
gpgcheck=0
enabled=1
priority={i}
module_hotfixes={mhf}
skip_if_unavailable=1`

		if extraOptions.ExcludePackages != nil && len(extraOptions.ExcludePackages) > 0 && !repo.IgnoreExclude {
			yumrepofsConfig += fmt.Sprintf("\nexclude=%s", strings.Join(extraOptions.ExcludePackages, " "))
		}

		mhf := "0"
		if repo.ModuleHotfixes {
			mhf = "1"
		}
		iStr := strconv.Itoa(i)
		rendered := strings.NewReplacer("{url}", repoUrl, "{i}", iStr, "{mhf}", mhf).Replace(yumrepofsConfig)
		ret[fmt.Sprintf("/etc/yum.repos.d/yumrepofs_%d.repo", i)] = rendered
	}

	repos, err := c.db.GetExternalRepositoriesForProject(projectId)
	if err != nil {
		return nil, err
	}

	for i, repo := range repos {
		repoConfig := `[peridotexternal_{i}]
name=Peridot External {i}
baseurl={url}
gpgcheck=0
enabled=1
priority={priority}
module_hotfixes={module_hotfixes}`

		if extraOptions.ExcludePackages != nil && len(extraOptions.ExcludePackages) > 0 {
			repoConfig += fmt.Sprintf("\nexclude=%s", strings.Join(extraOptions.ExcludePackages, " "))
		}

		repoUrl := strings.NewReplacer("$arch", arch, "$basearch", arch).Replace(repo.Url)
		// Sometimes i686 may be called i386 for repo purposes
		// Koji uses i386 for example for repo path but builds for i686
		// Let's just make sure we can use the repo
		if arch == "i686" {
			res, err := http.Get(fmt.Sprintf("%s/repodata/repomd.xml", repoUrl))
			if err != nil || res.StatusCode != 200 {
				repoUrl = strings.NewReplacer("$arch", "i386", "$basearch", "i386").Replace(repo.Url)
			}
		} else if arch == "noarch" {
			repoUrl = repo.Url
		}

		priority := strconv.Itoa(repo.Priority)
		moduleHotfixes := "0"
		if repo.ModuleHotfixes {
			moduleHotfixes = "1"
		}
		rendered := strings.NewReplacer("{i}", strconv.Itoa(i), "{url}", repoUrl, "{priority}", priority, "{module_hotfixes}", moduleHotfixes).Replace(repoConfig)
		ret[fmt.Sprintf("/etc/yum.repos.d/peridotexternal_%d.repo", i)] = rendered
	}

	for _, plugin := range c.plugins {
		entries := plugin.RepoEntries()
		if entries == nil {
			continue
		}

		for _, entry := range entries {
			ret[fmt.Sprintf("/etc/yum.repos.d/%s.repo", uuid.New().String())] = entry
		}
	}

	return ret, nil
}

func (c *Controller) initializeRepos(projectId string, arch string, extraOptions *peridotpb.ExtraBuildOptions) error {
	repos, err := c.repos(projectId, arch, extraOptions)
	if err != nil {
		return err
	}
	for k, v := range repos {
		err := os.WriteFile(k, []byte(v), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) buildMacros(project *models.Project, packageVersion *models.PackageVersion) map[string]string {
	distTag := fmt.Sprintf("el%d", project.MajorVersion)
	if project.DistTagOverride.Valid {
		distTag = project.DistTagOverride.String
	}
	if project.FollowImportDist && packageVersion != nil && !strings.Contains(packageVersion.Release, ".module+") {
		if releaseDistRegex.MatchString(packageVersion.Release) {
			subMatch := releaseDistRegex.FindStringSubmatch(packageVersion.Release)
			distTag = subMatch[1]
		}
	}
	majorVersion := strconv.Itoa(project.MajorVersion)

	vendor := strings.ToUpper(string(project.AdditionalVendor[0])) + project.AdditionalVendor[1:]
	packager := vendor
	if project.VendorMacro.String != "" {
		vendor = project.VendorMacro.String
	}
	if project.PackagerMacro.String != "" {
		packager = project.PackagerMacro.String
	}

	ret := map[string]string{
		"%__bootstrap":  "~bootstrap",
		"%vendor":       vendor,
		"%packager":     packager,
		"%distribution": project.Name,
		"%dist":         "%{!?distprefix0:%{?distprefix}}%{expand:%{lua:for i=0,9999 do print(\"%{?distprefix\" .. i ..\"}\") end}}." + distTag + "%{?with_bootstrap:~bootstrap}",
	}

	switch project.TargetVendor {
	case "redhat":
		ret["%rhel"] = majorVersion
	case "suse":
		ret["%sles_version"] = "0"
		ret["%suse_version"] = fmt.Sprintf("%s00", majorVersion)
	}

	return ret
}

func (c *Controller) setBuildMacros(project *models.Project, packageVersion *models.PackageVersion) error {
	err := os.RemoveAll("/etc/rpm/macros.dist")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	macros := c.buildMacros(project, packageVersion)
	var rendered string
	for k, v := range macros {
		rendered += fmt.Sprintf("%s %s\n", k, v)
	}

	err = os.WriteFile("/etc/rpm/macros.dist", []byte(rendered), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) yumConfig(project *models.Project) string {
	yumConfig := `[main]
debuglevel=1
retries=20
obsoletes=1
gpgcheck=0
assumeyes=1
keepcache=1
best=1
syslog_ident=peridotbuilder
syslog_device=
metadata_expire=0
install_weak_deps=0
protected_packages=
reposdir=/dev/null
logfile=/var/log/yum.log
mdpolicy=group:primary
metadata_expire=0
user_agent=peridotbuilder`

	switch project.TargetVendor {
	case "redhat":
		yumConfig += "\nmodule_platform_id=platform:el{majorVersion}"
	}

	majorVersion := strconv.Itoa(project.MajorVersion)
	return strings.NewReplacer("{majorVersion}", majorVersion).Replace(yumConfig)
}

func (c *Controller) setYumConfig(project *models.Project) error {
	yumConfigPath := "/etc/yum.conf"

	_, err := os.Stat(yumConfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		yumConfigPath = "/etc/dnf/dnf.conf"
	}
	yumConfigPath, err = filepath.EvalSymlinks(yumConfigPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(yumConfigPath, []byte(c.yumConfig(project)), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) chrootPkgs(project *models.Project, pkgGroup []string) []string {
	chrootPkgs := pkgGroup
	if project.TargetVendor == "redhat" {
		chrootPkgs = append(chrootPkgs, "redhat-rpm-config")
	}
	for _, plugin := range c.plugins {
		if plugin.Packages() != nil {
			chrootPkgs = append(chrootPkgs, plugin.Packages()...)
		}
	}

	return chrootPkgs
}

func (c *Controller) mockConfig(project *models.Project, packageVersion *models.PackageVersion, extra *peridotpb.ExtraBuildOptions, arch string, hostArch string, pkgGroup []string) (string, error) {
	// If we're building for i686 then force host arch to i686 even if we're building on x86_64
	if arch == "i686" {
		hostArch = "i686"
	}

	buildMacros := c.buildMacros(project, packageVersion)
	if extra != nil && extra.ForceDist != "" {
		buildMacros["%dist"] = "." + extra.ForceDist
	}

	mockConfig := `
config_opts['root'] = '{additionalVendor}-{majorVersion}-{hostArch}'
config_opts['target_arch'] = '{arch}'
config_opts['legal_host_arches'] = [{hostArches}]
config_opts['chroot_setup_cmd'] = 'install {chrootPkgs}'
config_opts['dist'] = '{dist}'
config_opts['releasever'] = '{majorVersion}'
config_opts['package_manager'] = 'dnf'
config_opts['extra_chroot_dirs'] = [ '/run/lock' ]
config_opts['rpmbuild_command'] = '{rpmbuildCommand}'
config_opts['use_bootstrap_image'] = False
config_opts['plugin_conf']['rpmautospec_enable'] = True
config_opts['plugin_conf']['rpmautospec_opts'] = {
  'requires': ['rpmautospec'],
  'cmd_base': ['/usr/bin/rpmautospec', 'process-distgit'],
}
{additionalVendorConfig}

config_opts['plugin_conf']['ccache_enable'] = False
config_opts['plugin_conf']['root_cache_enable'] = False
config_opts['plugin_conf']['yum_cache_enable'] = False
config_opts['rpmbuild_networking'] = {rpmbuildNetworking}
config_opts['use_host_resolv'] = {rpmbuildNetworking}
config_opts['print_main_output'] = True

config_opts['macros']['%_rpmfilename'] = '%%{NAME}-%%{VERSION}-%%{RELEASE}.%%{ARCH}.rpm'
config_opts['macros']['%_host'] = '{hostArch}-{targetVendor}-linux-gnu'
config_opts['macros']['%_host_cpu'] = '{hostArch}'
config_opts['macros']['%_vendor'] = "{targetVendor}"
config_opts['macros']['%_vendor_host'] = "{targetVendor}"

config_opts['module_setup_commands'] = [{moduleSetupCommands}]

`
	for k, v := range buildMacros {
		mockConfig += fmt.Sprintf("config_opts['macros']['%s'] = '%s'\n", k, v)
	}

	if extra == nil {
		extra = &peridotpb.ExtraBuildOptions{}
	}
	if extra.BuildArchExtraFiles == nil {
		extra.BuildArchExtraFiles = map[string]string{}
	}
	var macrosRendered string
	for k, v := range buildMacros {
		macrosRendered += fmt.Sprintf("%s %s\n", k, v)
	}
	extra.BuildArchExtraFiles["/usr/lib/rpm/macros.d/macros.dist"] = macrosRendered

	for k, v := range extra.BuildArchExtraFiles {
		mockConfig += fmt.Sprintf(`config_opts['files']['%s'] = """
%s
"""
`, strings.TrimPrefix(k, "/"), v)
	}

	for _, env := range rpmbuild.CmdDefaultArgs {
		spl := strings.SplitN(env, "=", 2)
		mockConfig += fmt.Sprintf("config_opts['environment']['%s'] = '%s'\n", spl[0], spl[1])
	}

	var tmpHostArches []string
	if hostArch == "i686" {
		tmpHostArches = []string{"i386", "i486", "i586", "i686", "x86_64"}
	} else {
		tmpHostArches = []string{hostArch}
	}
	tmpHostArches = append(tmpHostArches, "noarch")
	var hostArches []string
	for _, v := range tmpHostArches {
		hostArches = append(hostArches, fmt.Sprintf("'%s'", v))
	}

	var moduleSetupCommands []string
	for _, module := range extra.Modules {
		moduleSetupCommands = append(moduleSetupCommands, fmt.Sprintf("('enable', '%s')", module))
	}
	for _, module := range extra.DisabledModules {
		moduleSetupCommands = append(moduleSetupCommands, fmt.Sprintf("('disable', '%s')", module))
	}

	mockConfig += "\n"
	mockConfig += `
config_opts['dnf.conf'] = """
{yumConfig}
`

	repos, err := c.repos(project.ID.String(), arch, extra)
	if err != nil {
		return "", err
	}
	for _, repo := range repos {
		mockConfig += fmt.Sprintf("%s\n", repo)
	}
	mockConfig += `"""
`

	yumConfig := c.yumConfig(project)

	additionalVendorConfig := ""
	if project.TargetVendor == "suse" {
		additionalVendorConfig = `config_opts['useradd'] = '/usr/sbin/useradd -o -m -u {{chrootuid}} -g {{chrootgid}} -d {{chroothome}} {{chrootuser}}'
config_opts['ssl_ca_bundle_path'] = '/var/lib/ca-certificates/ca-bundle.pem'
config_opts['package_manager_max_attempts'] = 4
config_opts['package_manager_attempt_delay'] = 20`
	}

	rpmbuildNetworking := "False"
	if extra.EnableNetworking {
		rpmbuildNetworking = "True"
	}
	rendered := strings.NewReplacer(
		"{additionalVendor}", project.AdditionalVendor,
		"{majorVersion}", strconv.Itoa(project.MajorVersion),
		"{arch}", arch,
		"{hostArch}", hostArch,
		"{hostArches}", strings.Join(hostArches, ","),
		"{dist}", buildMacros["%dist"],
		"{yumConfig}", yumConfig,
		"{chrootPkgs}", strings.Join(c.chrootPkgs(project, pkgGroup), " "),
		"{rpmbuildCommand}", rpmbuild.ExecPath,
		"{targetVendor}", project.TargetVendor,
		"{moduleSetupCommands}", strings.Join(moduleSetupCommands, ","),
		"{rpmbuildNetworking}", rpmbuildNetworking,
		"{additionalVendorConfig}", additionalVendorConfig,
	).Replace(mockConfig)

	return rendered, nil
}

func (c *Controller) writeMockConfig(project *models.Project, packageVersion *models.PackageVersion, extra *peridotpb.ExtraBuildOptions, arch string, hostArch string, pkgGroup []string) error {
	mockConfig, err := c.mockConfig(project, packageVersion, extra, arch, hostArch, pkgGroup)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("/var/peridot/mock.cfg", []byte(mockConfig), 0644)
}

// BuildArchActivity builds a package for a given arch
// 26.04.2022: This activity had a huge rework with shelling out and chroot
// Previously it only used Go calls, but architectures like i686
// forced us to change to this method
// 03.05.2022: This activity was once again reworked to use mock.
// The reason being odd issues caused by the way we did chroot/unshare.
// This only affected a few packages, but we're in a hurry.
// Current implementation is broken for modules
// todo(mustafa): Evaluate if we can skip chroot again
func (c *Controller) BuildArchActivity(ctx context.Context, projectId string, packageName string, disableChecks bool, packageVersion *models.PackageVersion, uploadSRPMResult *UploadActivityResult, task *models.Task, arch string, extraOptions *peridotpb.ExtraBuildOptions) error {
	stopChan := makeHeartbeat(ctx, 10*time.Second)
	defer func() { stopChan <- true }()

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

	objBytes, err := c.storage.ReadObject(uploadSRPMResult.ObjectName)
	if err != nil {
		return err
	}

	cloneDir := rpmbuild.GetCloneDirectory()
	err = os.MkdirAll(filepath.Join(cloneDir, "SRPMS"), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(cloneDir, "RPMS"), 0755)
	if err != nil {
		return err
	}

	srpmPath := filepath.Join(cloneDir, "SRPMS", filepath.Base(uploadSRPMResult.ObjectName))
	f, err := os.OpenFile(srpmPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = f.Write(objBytes)
	if err != nil {
		return err
	}

	err = runCmd("chown", "-R", "peridotbuilder:mock", cloneDir)
	if err != nil {
		return fmt.Errorf("could not chown clone dir: %v", err)
	}

	// todo(mustafa): Temporal doesn't support Activity interceptors yet.
	// todo(mustafa): https://github.com/temporalio/proposals/pull/45
	if err := c.preExecPlugins("BuildArchActivity"); err != nil {
		return err
	}

	var pkgGroup = DefaultBuildPkgGroup

	if len(project.BuildStagePackages) != 0 {
		pkgGroup = project.BuildStagePackages
	}

	var enableModules []string
	var disableModules []string
	err = ParsePackageExtraOptions(pkgEo, &pkgGroup, &enableModules, &disableModules)

	if err != nil {
		c.log.Infof("no extra options to process for package")
	}

	if extraOptions.DisabledModules == nil {
		extraOptions.DisabledModules = []string{}
	}
	extraOptions.DisabledModules = append(extraOptions.DisabledModules, disableModules...)

	if extraOptions.Modules == nil {
		extraOptions.Modules = []string{}
	}
	extraOptions.Modules = append(extraOptions.Modules, enableModules...)

	hostArch := os.Getenv("REAL_BUILD_ARCH")
	err = c.writeMockConfig(&project, packageVersion, extraOptions, arch, hostArch, pkgGroup)
	if err != nil {
		return fmt.Errorf("could not write mock config: %v", err)
	}
	args := []string{
		"mock",
		"--isolation=simple",
		"-r",
		"/var/peridot/mock.cfg",
		"--target",
		arch,
		"--resultdir",
		filepath.Join(cloneDir, "RPMS"),
	}
	if disableChecks {
		args = append(args, "--nocheck")
	}
	if pkgEo != nil {
		for _, with := range pkgEo.WithFlags {
			args = append(args, "--with="+with)
		}
		for _, without := range pkgEo.WithoutFlags {
			args = append(args, "--without="+without)
		}
	}
	args = append(args, srpmPath)

	cmd := exec.Command("/bundle/fork-exec.py", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not mock build: %v", err)
	}

	// todo(mustafa): Remove once Temporal supports Activity interceptors
	if err := c.postExecPlugins("BuildArchActivity"); err != nil {
		return err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return nil
}

func (c *Controller) UploadArchActivity(ctx context.Context, projectId string, parentTaskId string) ([]*UploadActivityResult, error) {
	stopChan := makeHeartbeat(ctx, 4*time.Second)
	defer func() { stopChan <- true }()

	rpms, err := findRpms()
	if err != nil {
		return nil, err
	}

	var ret []*UploadActivityResult

	for _, rpm := range rpms {
		var nvr []string
		base := strings.TrimSuffix(filepath.Base(rpm), ".rpm")
		if rpmutils.NVRUnusualRelease().MatchString(base) {
			nvr = rpmutils.NVRUnusualRelease().FindStringSubmatch(base)
		} else if rpmutils.NVR().MatchString(base) {
			nvr = rpmutils.NVR().FindStringSubmatch(base)
		}
		if !rpmutils.NVR().MatchString(base) {
			return nil, errors.New("invalid rpm")
		}
		res, err := c.uploadArtifact(projectId, parentTaskId, rpm, nvr[4], peridotpb.TaskType_TASK_TYPE_BUILD_ARCH_UPLOAD)
		if err != nil {
			return nil, err
		}
		ret = append(ret, res)
	}

	return ret, nil
}
