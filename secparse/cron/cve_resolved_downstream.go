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
	"github.com/sirupsen/logrus"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"strings"
)

func (i *Instance) CheckIfCVEResolvedDownstream() {
	if i.koji == nil {
		logrus.Infoln("Automatic build checks are disabled. Provide a Koji endpoint using --koji-endpoint")
		return
	}

	cves, err := i.db.GetAllCVEsWithAllProductsFixed()
	if err != nil {
		logrus.Errorf("could not get fixed cves: %v", err)
		return
	}

	productBuffer := map[int64]*db.Product{}
	ignoredPackagesBuffer := map[string][]string{}

	for _, cve := range cves {
		affectedProducts, err := i.db.GetAllAffectedProductsByCVE(cve.ID)
		if err != nil {
			logrus.Errorf("could not get all affected products by %s: %v", cve.ID, err)
			continue
		}

		beginTx, err := i.db.Begin()
		if err != nil {
			logrus.Errorf("could not begin transaction: %v", err)
			continue
		}
		tx := i.db.UseTransaction(beginTx)

		didSkipProduct := false
		willNotFixOnly := true
		allFixed := true

		for _, affectedProduct := range affectedProducts {
			switch affectedProduct.State {
			case
				int(secparseadminpb.AffectedProductState_WillNotFixUpstream),
				int(secparseadminpb.AffectedProductState_OutOfSupportScope):
				continue
			case
				int(secparseadminpb.AffectedProductState_UnderInvestigationUpstream),
				int(secparseadminpb.AffectedProductState_AffectedUpstream):
				allFixed = false
				willNotFixOnly = false
				continue
			}

			skipProduct := false

			if productBuffer[affectedProduct.ProductID] == nil {
				product, err := i.db.GetProductByID(affectedProduct.ProductID)
				if err != nil {
					logrus.Errorf("could not get product with id %d: %v", affectedProduct.ProductID, err)
					continue
				}
				productBuffer[affectedProduct.ProductID] = product
			}
			product := productBuffer[affectedProduct.ProductID]

			if ignoredPackagesBuffer[product.ShortCode] == nil {
				ignoredUpstreamPackages, err := i.db.GetAllIgnoredPackagesByShortCode(product.ShortCode)
				if err != nil {
					logrus.Errorf("could not get ignored packages: %v", err)
					continue
				}
				ignoredPackagesBuffer[product.ShortCode] = ignoredUpstreamPackages
			}
			ignoredUpstreamPackages := ignoredPackagesBuffer[product.ShortCode]

			nvrOnly := strings.Replace(affectedProduct.Package, ":", "-", 1)
			if i.module.MatchString(nvrOnly) {
				if !affectedProduct.Advisory.Valid {
					skipProduct = true
					break
				}

				redHatAdvisory, err := i.errata.GetErrata(affectedProduct.Advisory.String)
				if err != nil {
					logrus.Errorf("Could not get Red Hat Advisory: %v", err)
					skipProduct = true
					break
				}

				for _, arch := range product.Archs {
					redHatProductName := affectedProductNameForArchAndVersion(arch, product.RedHatMajorVersion.Int32)
					affected := redHatAdvisory.AffectedProducts[redHatProductName]
					if affected == nil {
						continue
					}
					srpms := affected.SRPMs
					for _, srpm := range srpms {
						status := i.checkKojiForBuild(tx, ignoredUpstreamPackages, srpm, affectedProduct, cve)
						if status == Skip {
							skipProduct = true
							break
						} else if status == Fixed {
							willNotFixOnly = false
						} else if status == NotFixed {
							allFixed = false
							willNotFixOnly = false
						}
					}
					break
				}
				if skipProduct {
					logrus.Errorf("%s has not been fixed for NVR %s", cve.ID, nvrOnly)
					break
				}
			} else {
				nvrOnly = i.epoch.ReplaceAllString(affectedProduct.Package, "")
				status := i.checkKojiForBuild(tx, ignoredUpstreamPackages, nvrOnly, affectedProduct, cve)
				if status == Skip {
					skipProduct = true
					break
				} else if status == Fixed {
					willNotFixOnly = false
				} else if status == NotFixed {
					allFixed = false
					willNotFixOnly = false
				}
			}

			if skipProduct {
				didSkipProduct = true
				logrus.Infof("%s: Skipping package for now", affectedProduct.Package)
				_ = beginTx.Rollback()
				break
			}
		}

		if !didSkipProduct {
			newState := secparseadminpb.CVEState_ResolvedUpstream
			if allFixed {
				newState = secparseadminpb.CVEState_ResolvedDownstream
			}
			if willNotFixOnly {
				newState = secparseadminpb.CVEState_ResolvedNoAdvisory
			}
			err := tx.UpdateCVEState(cve.ID, newState)
			if err != nil {
				logrus.Errorf("Could not save new CVE state: %v", err)
				continue
			}
			err = beginTx.Commit()
			if err != nil {
				logrus.Errorf("could not commit transaction: %v", err)
				continue
			}

			logrus.Infof("%s is now set to %s", cve.ID, newState.String())
		}
	}
}
