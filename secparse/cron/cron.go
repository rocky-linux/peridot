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

package cron

import (
	"database/sql"
	"fmt"
	"github.com/gobwas/glob"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"peridot.resf.org/koji"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"peridot.resf.org/secparse/rherrata"
	"peridot.resf.org/secparse/rhsecurity"
	"peridot.resf.org/secparse/rpmutils"
	"regexp"
	"strconv"
	"strings"
)

type Instance struct {
	db     db.Access
	api    rhsecurity.DefaultApi
	errata rherrata.APIService

	koji              koji.API
	kojiCompose       string
	kojiModuleCompose string

	nvr             *regexp.Regexp
	epoch           *regexp.Regexp
	module          *regexp.Regexp
	dist            *regexp.Regexp
	moduleDist      *regexp.Regexp
	advisoryIdRegex *regexp.Regexp
}

type BuildStatus int

const (
	Fixed BuildStatus = iota
	NotFixed
	WillNotFix
	Skip
)

func New(access db.Access) (*Instance, error) {
	instance := &Instance{
		db:              access,
		api:             rhsecurity.NewAPIClient(rhsecurity.NewConfiguration()).DefaultApi,
		errata:          rherrata.NewClient(),
		nvr:             rpmutils.NVR(),
		epoch:           rpmutils.Epoch(),
		module:          rpmutils.Module(),
		dist:            rpmutils.Dist(),
		moduleDist:      rpmutils.ModuleDist(),
		advisoryIdRegex: rpmutils.AdvisoryId(),
	}

	if kojiEndpoint := viper.GetString("koji-endpoint"); kojiEndpoint != "" {
		var err error
		instance.koji, err = koji.New(kojiEndpoint)
		if err != nil {
			return nil, err
		}
		instance.kojiCompose = viper.GetString("koji-compose")
		instance.kojiModuleCompose = viper.GetString("koji-module-compose")
	}

	return instance, nil
}

// productName simply appends major version to `Red Hat Enterprise Linux`
func productName(majorVersion int32) string {
	return fmt.Sprintf("Red Hat Enterprise Linux %d", majorVersion)
}

// affectedProductNameForArchAndVersion creates appropriate upstream product names for arch and version
// This is then used to parse affected packages
func affectedProductNameForArchAndVersion(arch string, majorVersion int32) string {
	var archString string
	switch arch {
	case "x86_64":
		archString = "x86_64"
		break
	case "aarch64":
		archString = "ARM 64"
		break
	case "ppc64le":
		archString = "Power, little endian"
		break
	case "s390x":
		archString = "IBM z Systems"
		break
	default:
		archString = "UnknownBreakOnPurpose"
		break
	}
	return fmt.Sprintf("Red Hat Enterprise Linux for %s 8", archString)
}

// productState returns appropriate proto type for string states
func productState(state string) secparseadminpb.AffectedProductState {
	switch state {
	case "Under investigation":
		return secparseadminpb.AffectedProductState_UnderInvestigationUpstream
	case "Not affected":
		return secparseadminpb.AffectedProductState_UnknownProductState
	case "Will not fix":
		return secparseadminpb.AffectedProductState_WillNotFixUpstream
	case "Out of support scope":
		return secparseadminpb.AffectedProductState_OutOfSupportScope
	case "Affected":
		return secparseadminpb.AffectedProductState_AffectedUpstream
	default:
		return secparseadminpb.AffectedProductState_UnderInvestigationUpstream
	}
}

// checkProduct is used to check and validate CVE package states and releases
func (i *Instance) checkProduct(tx db.Access, cve *db.CVE, shortCode *db.ShortCode, product *db.Product, productState secparseadminpb.AffectedProductState, packageName string, advisory *string) bool {
	// Re-create a valid product name using the short code prefix and major version.
	// Example: Red Hat Enterprise Linux 8 translates to Rocky Linux 8 for the short code `RL`.
	// Check `//secparse:seed.sql` for more info
	mirrorProductName := fmt.Sprintf("%s %d", shortCode.RedHatProductPrefix.String, product.RedHatMajorVersion.Int32)

	// Get the affected product if exists
	affectedProduct, err := tx.GetAffectedProductByCVEAndPackage(cve.ID, packageName)
	if err != nil {
		// The affected product does not exist, so we can mark this product as affected if this product exists
		if err == sql.ErrNoRows {
			// Check if the current package name matches an NVR and if we have a non-NVR variant
			skipCreate := false
			epochlessPackage := i.epoch.ReplaceAllString(packageName, "")
			if i.nvr.MatchString(epochlessPackage) {
				nvr := i.nvr.FindStringSubmatch(epochlessPackage)
				affectedProduct, err = tx.GetAffectedProductByCVEAndPackage(cve.ID, nvr[1])
				if err == nil {
					skipCreate = true
				}
			}

			if !skipCreate {
				// Get the mirrored product name product if exists (this should exist if supported)
				// Example: Rocky Linux only supports 8 so we will only have `Rocky Linux 8` in our supported products
				// In the future, when we support 8 and 9 at the same time, we only need to add `Rocky Linux 9` to start
				// mirroring errata for el9 packages
				product, err := tx.GetProductByNameAndShortCode(mirrorProductName, shortCode.Code)
				if err != nil {
					// Product isn't supported so skip
					if err == sql.ErrNoRows {
						logrus.Infof("Product %s not supported", mirrorProductName)
						return true
					} else {
						logrus.Errorf("could not get product: %v", err)
						return true
					}
				}

				// If product state isn't set to unknown (usually when product isn't affected)
				// create a new affected product entry for the CVE
				if productState != secparseadminpb.AffectedProductState_UnknownProductState {
					affectedProduct, err = tx.CreateAffectedProduct(product.ID, cve.ID, int(productState), product.CurrentFullVersion, packageName, advisory)
					if err != nil {
						logrus.Errorf("could not create affected product: %v", err)
						return true
					}
					logrus.Infof("Added product %s (%s) to %s with state %s", mirrorProductName, packageName, cve.ID, productState.String())
				}
			}
		} else {
			logrus.Errorf("could not get affected product: %v", err)
			return true
		}
	}

	// We don't use else because this may change if a non-NVR variant is found
	if err == nil {
		// If the state isn't set to unknown (it is then usually queued for deletion)
		if productState != secparseadminpb.AffectedProductState_UnknownProductState {
			// If it's already in that state, skip
			if int(productState) == affectedProduct.State {
				return true
			}

			// If the affected product is set to FixedDownstream and we're trying to set it to FixedUpstream, skip
			if affectedProduct.State == int(secparseadminpb.AffectedProductState_FixedDownstream) && productState == secparseadminpb.AffectedProductState_FixedUpstream {
				return true
			}

			err := tx.UpdateAffectedProductStateAndPackageAndAdvisory(affectedProduct.ID, int(productState), packageName, advisory)
			if err != nil {
				logrus.Errorf("could not update affected product state: %v", err)
				return true
			}
			logrus.Infof("Updated product %s (%s) on %s with state %s", mirrorProductName, packageName, cve.ID, productState.String())
		} else {
			// Delete affected product if state is set to Unknown
			// That means that the product is set as NotAffected
			err = tx.DeleteAffectedProduct(affectedProduct.ID)
			if err != nil {
				logrus.Errorf("could not delete unaffected product: %v", err)
				return true
			}
			logrus.Infof("Product %s (%s) not affected by %s", mirrorProductName, packageName, cve.ID)
		}
	}

	return false
}

func (i *Instance) isNvrIdentical(build *koji.Build, nvr []string) bool {
	// Join all release bits and remove the dist tag (because sometimes downstream forks do not match the upstream dist tag)
	// Example: Rocky Linux 8.3 initial build did not tag updated RHEL packages as el8_3, but as el8
	joinedRelease := i.dist.ReplaceAllString(strings.TrimSuffix(strings.Join(nvr[2:], "."), "."), "")
	// Remove all module release bits (to make it possible to actually match NVR)
	joinedRelease = i.moduleDist.ReplaceAllString(joinedRelease, "")
	// Same operations for the build release
	buildRelease := i.dist.ReplaceAllString(build.Release, "")
	buildRelease = i.moduleDist.ReplaceAllString(buildRelease, "")

	// Check if package name, version matches and that the release prefix matches
	// The reason we're only checking for prefix in release is that downstream
	// builds may append `.1` or something else
	// Example: Rocky Linux appends `.rocky` to modified packages
	if build.PackageName == nvr[0] && build.Version == nvr[1] && strings.HasPrefix(buildRelease, joinedRelease) {
		return true
	}

	return false
}

func (i *Instance) checkForIgnoredPackage(ignoredPackages []string, packageName string) (bool, error) {
	for _, ignoredPackage := range ignoredPackages {
		g, err := glob.Compile(ignoredPackage)
		if err != nil {
			return false, err
		}

		if g.Match(packageName) {
			return true, nil
		}
	}

	return false, nil
}

func (i *Instance) checkKojiForBuild(tx db.Access, ignoredPackages []string, nvrOnly string, affectedProduct *db.AffectedProduct, cve *db.CVE) BuildStatus {
	// Check if the submitted NVR is valid
	nvr := i.nvr.FindStringSubmatch(nvrOnly)
	if len(nvr) < 3 {
		logrus.Errorf("Invalid NVR %s", nvrOnly)
		return Skip
	}
	nvr = nvr[1:]

	match, err := i.checkForIgnoredPackage(ignoredPackages, nvr[0])
	if err != nil {
		logrus.Errorf("Invalid glob: %v", err)
		return Skip
	}
	if match {
		return WillNotFix
	}

	var tagged []*koji.Build

	// If the package is part of a module, we have to check for valid builds
	// rather than check in the compose tag
	if strings.Contains(nvrOnly, ".module") {
		// We need to find the package id
		packageRes, err := i.koji.GetPackage(&koji.GetPackageRequest{
			PackageName: nvr[0],
		})
		if err != nil {
			logrus.Errorf("Could not get package information from Koji: %v", err)
			return Skip
		}

		// Use package id to get builds
		buildsRes, err := i.koji.ListBuilds(&koji.ListBuildsRequest{
			PackageID: packageRes.ID,
		})
		if err != nil {
			logrus.Errorf("Could not get builds from Koji: %v", err)
			return Skip
		}

		tagged = buildsRes.Builds
	} else {
		// Non-module packages can be queried using the list tagged operation.
		// We only check the compose tag
		taggedRes, err := i.koji.ListTagged(&koji.ListTaggedRequest{
			Tag:     i.kojiCompose,
			Package: nvr[0],
		})
		if err != nil {
			logrus.Errorf("Could not get tagged builds for package %s: %v", nvr[0], err)
			return Skip
		}

		tagged = taggedRes.Builds
	}

	// No valid builds found usually means that we don't ship that package
	if len(tagged) <= 0 {
		logrus.Errorf("No valid builds found for package %s", nvr[0])
		return NotFixed
	}

	// Use a top-level fixed state to track if the NVR exists (at least once for modules)
	fixed := false
	for _, build := range tagged {
		latestBuild := build
		// Skip module contents (this is content inserted by module-build-service)
		if latestBuild.Extra != nil && latestBuild.Extra.Typeinfo != nil {
			continue
		}

		// Re-construct a valid NVR
		kojiNvr := fmt.Sprintf("%s-%s-%s", latestBuild.PackageName, latestBuild.Version, latestBuild.Release)

		// If the NVR is identical, that means that the fix has been built
		if i.isNvrIdentical(latestBuild, nvr) {
			logrus.Infof("%s has been fixed downstream with build %d (%s)", cve.ID, latestBuild.BuildId, kojiNvr)
			err := tx.UpdateAffectedProductStateAndPackageAndAdvisory(affectedProduct.ID, int(secparseadminpb.AffectedProductState_FixedDownstream), affectedProduct.Package, &affectedProduct.Advisory.String)
			if err != nil {
				logrus.Errorf("Could not update affected product %d: %v", affectedProduct.ID, err)
				return Skip
			}

			// Get all RPMs for build
			rpms, err := i.koji.ListRPMs(&koji.ListRPMsRequest{
				BuildID: latestBuild.BuildId,
			})
			if err != nil {
				logrus.Errorf("Could not get RPMs from Koji: %v", err)
				return Skip
			}

			var srcRpm string
			for _, rpm := range rpms.RPMs {
				if rpm.Arch == "src" {
					epochInt := 0
					if rpm.Epoch != nil {
						epochInt = *rpm.Epoch
					}

					srcRpm = fmt.Sprintf("%s-%d:%s-%s.%s.rpm", rpm.Name, epochInt, rpm.Version, rpm.Release, rpm.Arch)
					break
				}
			}

			// Add all RPMs as a build reference to the CVE
			// This is the "Affected packages" section of an advisory
			for _, rpm := range rpms.RPMs {
				epochInt := 0
				if rpm.Epoch != nil {
					epochInt = *rpm.Epoch
				}

				// Construct a valid rpm name (this is what the repos will contain)
				rpmStr := fmt.Sprintf("%s-%d:%s-%s.%s.rpm", rpm.Name, epochInt, rpm.Version, rpm.Release, rpm.Arch)
				_, err = tx.CreateBuildReference(affectedProduct.ID, rpmStr, srcRpm, cve.ID, strconv.Itoa(latestBuild.BuildId))
				if err != nil {
					logrus.Errorf("Could not create build reference: %v", err)
					return Skip
				}
			}

			// We've seen at least one fix
			fixed = true
			// Since we've seen a fix, we don't have to keep looking
			break
		}
	}

	// No fix has been detected, will mark as FixedUpstream
	if !fixed {
		logrus.Errorf("%s has not been fixed for NVR %s", cve.ID, nvrOnly)
		return NotFixed
	}

	return Fixed
}
