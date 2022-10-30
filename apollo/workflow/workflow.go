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
	"fmt"
	"github.com/gobwas/glob"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rherrata"
	"peridot.resf.org/apollo/rhsecurity"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/koji"
	"peridot.resf.org/utils"
	"strconv"
	"strings"
)

var forceKoji koji.API

type Controller struct {
	log       *logrus.Logger
	temporal  client.Client
	db        apollodb.Access
	mainQueue string

	errata   rherrata.APIService
	security rhsecurity.DefaultApi

	vendor string
}

type Koji struct {
	Endpoint      string
	Compose       string
	ModuleCompose string
}

type NewControllerInput struct {
	Temporal  client.Client
	Database  apollodb.Access
	MainQueue string
}

type Option func(c *Controller)

func WithSecurityAPI(api rhsecurity.DefaultApi) Option {
	return func(c *Controller) {
		c.security = api
	}
}

func WithErrataAPI(api rherrata.APIService) Option {
	return func(c *Controller) {
		c.errata = api
	}
}

// NewController returns a new workflow controller. It is the entry point for the Temporal worker.
// Usually each project share a common controller with different workflows and activities enabled
// in the `cmd` package.
func NewController(input *NewControllerInput, opts ...Option) (*Controller, error) {
	c := &Controller{
		log:       logrus.New(),
		temporal:  input.Temporal,
		db:        input.Database,
		mainQueue: input.MainQueue,
		vendor:    viper.GetString("vendor"),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
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
func productState(state string) apollopb.AffectedProduct_State {
	switch state {
	case "Under investigation":
		return apollopb.AffectedProduct_STATE_UNDER_INVESTIGATION_UPSTREAM
	case "Not affected":
		return apollopb.AffectedProduct_STATE_UNKNOWN
	case "Will not fix":
		return apollopb.AffectedProduct_STATE_WILL_NOT_FIX_UPSTREAM
	case "Out of support scope":
		return apollopb.AffectedProduct_STATE_OUT_OF_SUPPORT_SCOPE
	case "Affected":
		return apollopb.AffectedProduct_STATE_AFFECTED_UPSTREAM
	default:
		return apollopb.AffectedProduct_STATE_UNDER_INVESTIGATION_UPSTREAM
	}
}

// checkProduct is used to check and validate CVE package states and releases
func (c *Controller) checkProduct(tx apollodb.Access, cve *apollodb.CVE, shortCode *apollodb.ShortCode, product *apollodb.Product, productState apollopb.AffectedProduct_State, packageName string, advisory *string) bool {
	// Re-create a valid product name using the short code prefix and major version.
	// Example: Red Hat Enterprise Linux 8 translates to Rocky Linux 8 for the short code `RL`.
	// Check `//apollo:seed.sql` for more info
	mirrorProductName := fmt.Sprintf("%s %d", product.RedHatProductPrefix.String, product.RedHatMajorVersion.Int32)

	// Get the affected product if exists
	affectedProduct, err := tx.GetAffectedProductByCVEAndPackage(cve.ID, packageName)
	if err != nil {
		// The affected product does not exist, so we can mark this product as affected if this product exists
		if err == sql.ErrNoRows {
			// Check if the current package name matches an NVR and if we have a non-NVR variant
			skipCreate := false
			epochlessPackage := rpmutils.Epoch().ReplaceAllString(packageName, "")
			if rpmutils.NVR().MatchString(epochlessPackage) {
				nvr := rpmutils.NVR().FindStringSubmatch(epochlessPackage)
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
				if productState != apollopb.AffectedProduct_STATE_UNKNOWN {
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
		if productState != apollopb.AffectedProduct_STATE_UNKNOWN {
			// If it's already in that state, skip
			if int(productState) == affectedProduct.State {
				return true
			}

			// If the affected product is set to FixedDownstream and we're trying to set it to FixedUpstream, skip
			if affectedProduct.State == int(apollopb.AffectedProduct_STATE_FIXED_DOWNSTREAM) && productState == apollopb.AffectedProduct_STATE_FIXED_UPSTREAM {
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

func (c *Controller) isNvrIdentical(build *koji.Build, nvr []string) bool {
	// Join all release bits and remove the dist tag (because sometimes downstream forks do not match the upstream dist tag)
	// Example: Rocky Linux 8.3 initial build did not tag updated RHEL packages as el8_3, but as el8
	joinedRelease := rpmutils.Dist().ReplaceAllString(strings.TrimSuffix(strings.Join(nvr[2:], "."), "."), "")
	// Remove all module release bits (to make it possible to actually match NVR)
	joinedRelease = rpmutils.ModuleDist().ReplaceAllString(joinedRelease, "")
	// Same operations for the build release
	buildRelease := rpmutils.Dist().ReplaceAllString(build.Release, "")
	buildRelease = rpmutils.ModuleDist().ReplaceAllString(buildRelease, "")

	// Check if package name, version matches and that the release prefix matches
	// The reason we're only checking for prefix in release is that downstream
	// builds may append `.1` or something else
	// Example: Rocky Linux appends `.rocky` to modified packages
	if build.PackageName == nvr[0] && build.Version == nvr[1] && strings.HasPrefix(buildRelease, joinedRelease) {
		return true
	}

	return false
}

func (c *Controller) checkForIgnoredPackage(ignoredPackages []string, packageName string) (bool, error) {
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

func (c *Controller) checkForRebootSuggestedPackage(pkgs []string, packageName string) (bool, error) {
	for _, p := range pkgs {
		g, err := glob.Compile(p)
		if err != nil {
			return false, err
		}

		if g.Match(packageName) {
			return true, nil
		}
	}

	return false, nil
}

func (c *Controller) checkKojiForBuild(tx apollodb.Access, ignoredPackages []string, nvrOnly string, affectedProduct *apollodb.AffectedProduct, cve *apollodb.CVE) apollopb.BuildStatus {
	product, err := tx.GetProductByID(affectedProduct.ProductID)
	if err != nil {
		c.log.Errorf("could not get product: %v", err)
		return apollopb.BuildStatus_BUILD_STATUS_SKIP
	}
	if product.BuildSystem != "koji" {
		return apollopb.BuildStatus_BUILD_STATUS_SKIP
	}

	var k koji.API
	if forceKoji != nil {
		k = forceKoji
	} else {
		k, err = koji.New(product.BuildSystemEndpoint)
		if err != nil {
			c.log.Errorf("could not create koji client: %v", err)
			return apollopb.BuildStatus_BUILD_STATUS_SKIP
		}
	}

	// Check if the submitted NVR is valid
	nvr := rpmutils.NVR().FindStringSubmatch(nvrOnly)
	if len(nvr) < 3 {
		logrus.Errorf("Invalid NVR %s", nvrOnly)
		return apollopb.BuildStatus_BUILD_STATUS_SKIP
	}
	nvr = nvr[1:]

	match, err := c.checkForIgnoredPackage(ignoredPackages, nvr[0])
	if err != nil {
		logrus.Errorf("Invalid glob: %v", err)
		return apollopb.BuildStatus_BUILD_STATUS_SKIP
	}
	if match {
		return apollopb.BuildStatus_BUILD_STATUS_WILL_NOT_FIX
	}

	var tagged []*koji.Build

	// If the package is part of a module, we have to check for valid builds
	// rather than check in the compose tag
	if strings.Contains(nvrOnly, ".module") {
		// We need to find the package id
		packageRes, err := k.GetPackage(&koji.GetPackageRequest{
			PackageName: nvr[0],
		})
		if err != nil {
			logrus.Errorf("Could not get package information from Koji: %v", err)
			return apollopb.BuildStatus_BUILD_STATUS_SKIP
		}

		// Use package id to get builds
		buildsRes, err := k.ListBuilds(&koji.ListBuildsRequest{
			PackageID: packageRes.ID,
		})
		if err != nil {
			logrus.Errorf("Could not get builds from Koji: %v", err)
			return apollopb.BuildStatus_BUILD_STATUS_SKIP
		}

		tagged = buildsRes.Builds
	} else {
		// Non-module packages can be queried using the list tagged operation.
		// We only check the compose tag
		taggedRes, err := k.ListTagged(&koji.ListTaggedRequest{
			Tag:     product.KojiCompose.String,
			Package: nvr[0],
		})
		if err != nil {
			logrus.Errorf("Could not get tagged builds for package %s: %v", nvr[0], err)
			return apollopb.BuildStatus_BUILD_STATUS_SKIP
		}

		tagged = taggedRes.Builds
	}

	// No valid builds found usually means that we don't ship that package
	if len(tagged) <= 0 {
		logrus.Errorf("No valid builds found for package %s", nvr[0])
		return apollopb.BuildStatus_BUILD_STATUS_NOT_FIXED
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
		if c.isNvrIdentical(latestBuild, nvr) {
			logrus.Infof("%s has been fixed downstream with build %d (%s)", cve.ID, latestBuild.BuildId, kojiNvr)
			err := tx.UpdateAffectedProductStateAndPackageAndAdvisory(affectedProduct.ID, int(apollopb.AffectedProduct_STATE_FIXED_DOWNSTREAM), affectedProduct.Package, &affectedProduct.Advisory.String)
			if err != nil {
				logrus.Errorf("Could not update affected product %d: %v", affectedProduct.ID, err)
				return apollopb.BuildStatus_BUILD_STATUS_SKIP
			}

			// Get all RPMs for build
			rpms, err := k.ListRPMs(&koji.ListRPMsRequest{
				BuildID: latestBuild.BuildId,
			})
			if err != nil {
				logrus.Errorf("Could not get RPMs from Koji: %v", err)
				return apollopb.BuildStatus_BUILD_STATUS_SKIP
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
				// Construct a valid rpm name (this is what the repos will contain)
				rpmStr := fmt.Sprintf("%s-%d:%s-%s.%s.rpm", rpm.Name, utils.Default[int](rpm.Epoch), rpm.Version, rpm.Release, rpm.Arch)
				_, err = tx.CreateBuildReference(affectedProduct.ID, rpmStr, srcRpm, cve.ID, "", utils.Pointer[string](strconv.Itoa(latestBuild.BuildId)), nil)
				if err != nil {
					logrus.Errorf("Could not create build reference: %v", err)
					return apollopb.BuildStatus_BUILD_STATUS_SKIP
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
		return apollopb.BuildStatus_BUILD_STATUS_NOT_FIXED
	}

	return apollopb.BuildStatus_BUILD_STATUS_FIXED
}
