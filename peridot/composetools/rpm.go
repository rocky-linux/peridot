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

package composetools

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/peridot/yummeta"
	"peridot.resf.org/utils"
)

var (
	ErrInvalidNVR = errors.New("invalid NVR")
)

var (
	debugSuffixes         = []string{"-debuginfo", "-debugsource", "-debuginfo-common"}
	debugRegex            = regexp.MustCompile(`^(?:.*-debuginfo(?:-.*)?|.*-debuginfo-.*|.*-debugsource)$`)
	soGlob                = glob.MustCompile("*.so.*")
	gtk2EnginesGlob       = glob.MustCompile("/usr/lib*/gtk-2.0/*/engines")
	gtk2ModulesGlob       = glob.MustCompile("/usr/lib*/gtk-2.0/*/modules")
	gtk2ImmodulesGlob     = glob.MustCompile("/usr/lib*/gtk-2.0/*/immodules")
	gtk2LoadersGlob       = glob.MustCompile("/usr/lib*/gtk-2.0/*/loaders")
	gtk2PrintBackendsGlob = glob.MustCompile("/usr/lib*/gtk-2.0/*/printbackends")
	gtk2FilesystemsGlob   = glob.MustCompile("/usr/lib*/gtk-2.0/*/filesystems")
	qtPluginsGlob         = glob.MustCompile("/usr/lib*/qt*/plugins/*")
	kdePluginsGlob        = glob.MustCompile("/usr/lib*/kde*/plugins/*")
	qt5QmlGlob            = glob.MustCompile("/usr/lib*/qt5/qml/*")
	gdkPixbufLoadersGlob  = glob.MustCompile("/usr/lib*/gdk-pixbuf-2.0/*/loaders")
	xinePluginsGlob       = glob.MustCompile("/usr/lib*/xine/plugins/*")
)

func StripDebugSuffixes(name string) string {
	ret := name
	for _, suffix := range debugSuffixes {
		ret = strings.TrimSuffix(ret, suffix)
	}
	return ret
}

func IsDebugPackage(name string) bool {
	if debugRegex.MatchString(name) {
		return true
	}
	return false
}

func IsDebugPackageNvra(nvra string) (bool, error) {
	if !rpmutils.NVR().MatchString(nvra) {
		return false, ErrInvalidNVR
	}

	if rpmutils.NVRUnusualRelease().MatchString(nvra) {
		match := rpmutils.NVRUnusualRelease().FindStringSubmatch(nvra)
		return IsDebugPackage(match[1]), nil
	}

	match := rpmutils.NVR().FindStringSubmatch(nvra)
	return IsDebugPackage(match[1]), nil
}

func GenNevraPrimaryPkg(pkg *yummeta.PrimaryPackage) string {
	return fmt.Sprintf("%s-%s:%s-%s.%s", pkg.Name, pkg.Version.Epoch, pkg.Version.Ver, pkg.Version.Rel, pkg.Arch)
}

func MultilibMethod(pkg *yummeta.PrimaryPackage, pkgFiles []*yummeta.FilelistsFile, excludeFilter []string, includeList []string) bool {
	prefer64 := []string{"gdb", "frysk", "systemtap", "systemtap-runtime", "ltrace", "strace"}

	if strings.Contains(pkg.Arch, "64") {
		if utils.StrContains(pkg.Name, prefer64) {
			return true
		}
		if strings.HasPrefix(pkg.Name, "kernel") {
			for _, prov := range pkg.Format.RpmProvides.RpmEntries {
				if prov.Name == "kernel" || prov.Name == "kernel-devel" {
					return true
				}
			}
		}
	}

	return false
}

// RuntimeMultilib returns if a given package should be multilib tested by runtime rules (contains *.so)
// Ported from https://pagure.io/releng/python-multilib/blob/master/f/multilib/multilib.py#_117
func RuntimeMultilib(pkg *yummeta.PrimaryPackage, pkgFiles []*yummeta.FilelistsFile, excludeFilter []string, includeList []string) (bool, error) {
	rootLibDirs := []string{"/lib", "/lib64"}
	usrLibDirs := []string{"/usr/lib", "/usr/lib64"}
	libDirs := append(rootLibDirs, usrLibDirs...)
	oprofileLibDirs := []string{"/usr/lib/oprofile", "/usr/lib64/oprofile"}
	wineDirs := []string{"/usr/lib/wine", "/usr/lib64/wine"}
	saneDirs := []string{"/usr/lib/sane", "/usr/lib64/sane"}

	byDir := []string{"/etc/lsb-release.d"}
	extraPkgs := []string{"alsa-lib", "dri", "gtk-2.0/modules", "gtk-2.0/immodules", "krb5/plugins", "sasl2", "vdpau"}

	for _, p := range extraPkgs {
		byDir = append(byDir, filepath.Join("/usr/lib/", p))
		byDir = append(byDir, filepath.Join("/usr/lib64/", p))
	}
	for _, p := range rootLibDirs {
		byDir = append(byDir, filepath.Join(p, "security"))
	}

	// If excluded, return false
	for _, filter := range excludeFilter {
		g, err := glob.Compile(filter)
		if err != nil {
			return false, err
		}
		if g.Match(pkg.Name) {
			return false, nil
		}
	}
	// If forcefully included, return true
	if utils.StrContains(pkg.Name, includeList) {
		return true, nil
	}
	if MultilibMethod(pkg, pkgFiles, excludeFilter, includeList) {
		return true, nil
	}
	if strings.HasPrefix(pkg.Name, "kernel") {
		for _, prov := range pkg.Format.RpmProvides.RpmEntries {
			if prov.Name == "kernel" {
				return false, nil
			}
		}
	}

	for _, file := range pkgFiles {
		dirName := filepath.Dir(file.Value)
		fileName := filepath.Base(file.Value)

		// If *.so file in LIBDIRS, return true
		if utils.StrContains(dirName, libDirs) && soGlob.Match(fileName) {
			return true, nil
		}
		if utils.StrContains(dirName, byDir) {
			return true, nil
		}
		if dirName == "/etc/ld.so.conf.d" && strings.HasSuffix(fileName, ".conf") {
			return true, nil
		}
		if utils.StrContains(dirName, rootLibDirs) && (strings.HasPrefix(fileName, "libnss_") || strings.HasPrefix(fileName, "libdb-")) {
			return true, nil
		}

		// Rest of the checks here is for usrLibDirs, skip if it doesn't start with usrLibDirs
		startsWithUsrLibDirs := false
		for _, usrLibDir := range usrLibDirs {
			if strings.HasPrefix(dirName, usrLibDir) {
				startsWithUsrLibDirs = true
				break
			}
		}
		if !startsWithUsrLibDirs {
			continue
		}

		if strings.HasPrefix(dirName, "/usr/lib/gtk-2.0") || strings.HasPrefix(dirName, "/usr/lib64/gtk-2.0") {
			if gtk2EnginesGlob.Match(dirName) {
				return true, nil
			}
			if gtk2ModulesGlob.Match(dirName) {
				return true, nil
			}
			if gtk2ImmodulesGlob.Match(dirName) {
				return true, nil
			}
			if gtk2LoadersGlob.Match(dirName) {
				return true, nil
			}
			if gtk2PrintBackendsGlob.Match(dirName) {
				return true, nil
			}
			if gtk2FilesystemsGlob.Match(dirName) {
				return true, nil
			}

			// If none-matches, continue
			continue
		}

		// gstreamer
		if strings.HasPrefix(dirName, "/usr/lib/gstreamer-") || strings.HasPrefix(dirName, "/usr/lib64/gstreamer-") {
			return true, nil
		}
		// qt/kde
		if qtPluginsGlob.Match(dirName) {
			return true, nil
		}
		if kdePluginsGlob.Match(dirName) {
			return true, nil
		}
		// qml
		if qt5QmlGlob.Match(dirName) {
			return true, nil
		}
		if gdkPixbufLoadersGlob.Match(dirName) {
			return true, nil
		}
		if xinePluginsGlob.Match(dirName) {
			return true, nil
		}

		if utils.StrContains(dirName, oprofileLibDirs) && soGlob.Match(fileName) {
			return true, nil
		}
		if utils.StrContains(dirName, wineDirs) && strings.HasSuffix(fileName, ".so") {
			return true, nil
		}
		if utils.StrContains(dirName, saneDirs) && strings.HasPrefix(fileName, "libsane-") {
			return true, nil
		}
	}

	return false, nil
}

func DevelMultilib(pkg *yummeta.PrimaryPackage, pkgFiles []*yummeta.FilelistsFile, excludeFilter []string, includeList []string) (bool, error) {
	// If excluded, return false
	for _, filter := range excludeFilter {
		g, err := glob.Compile(filter)
		if err != nil {
			return false, err
		}
		if g.Match(pkg.Name) {
			return false, nil
		}
	}
	if utils.StrContains(pkg.Name, excludeFilter) {
		return false, nil
	}
	// If forcefully included, return true
	if utils.StrContains(pkg.Name, includeList) {
		return true, nil
	}
	// If allowed in RuntimeMultilib, return true
	runtimeOk, err := RuntimeMultilib(pkg, pkgFiles, excludeFilter, includeList)
	if err != nil {
		return false, nil
	}
	if runtimeOk {
		return true, nil
	}
	if strings.HasPrefix(pkg.Name, "ghc-") {
		return false, nil
	}
	if strings.HasPrefix(pkg.Name, "kernel") {
		for _, prov := range pkg.Format.RpmProvides.RpmEntries {
			if prov.Name == "kernel-devel" {
				return false, nil
			}
			if strings.HasSuffix(prov.Name, "-devel") || strings.HasSuffix(prov.Name, "-static") {
				return true, nil
			}
		}
	}
	if strings.HasSuffix(pkg.Name, "-devel") {
		return true, nil
	}
	if strings.HasSuffix(pkg.Name, "-static") {
		return true, nil
	}

	return false, nil
}
