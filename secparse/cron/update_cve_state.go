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
	"context"
	"github.com/sirupsen/logrus"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
)

func (i *Instance) UpdateCVEState() {
	cves, err := i.db.GetAllUnresolvedCVEs()
	if err != nil {
		logrus.Errorf("could not get unresolved cves: %v", err)
		return
	}

	shortCodeBuffer := map[string]*db.ShortCode{}
	productBuffer := map[string][]*db.Product{}
	ignoredPackagesBuffer := map[string][]string{}

	ctx := context.TODO()
	for _, cve := range cves {
		if shortCodeBuffer[cve.ShortCode] == nil {
			shortCode, err := i.db.GetShortCodeByCode(cve.ShortCode)
			if err != nil {
				logrus.Errorf("could not get short code: %v", err)
				continue
			}

			shortCodeBuffer[shortCode.Code] = shortCode
		}
		shortCode := shortCodeBuffer[cve.ShortCode]

		if productBuffer[shortCode.Code] == nil {
			products, err := i.db.GetProductsByShortCode(shortCode.Code)
			if err != nil {
				logrus.Errorf("could not get products for code: %s: %v", shortCode.Code, err)
				continue
			}
			productBuffer[shortCode.Code] = products
		}
		products := productBuffer[shortCode.Code]

		if ignoredPackagesBuffer[shortCode.Code] == nil {
			ignoredUpstreamPackages, err := i.db.GetAllIgnoredPackagesByShortCode(shortCode.Code)
			if err != nil {
				logrus.Errorf("could not get ignored packages: %v", err)
				continue
			}
			ignoredPackagesBuffer[shortCode.Code] = ignoredUpstreamPackages
		}
		ignoredUpstreamPackages := ignoredPackagesBuffer[shortCode.Code]

		cveRh, _, err := i.api.GetCveExecute(i.api.GetCve(ctx, cve.ID))
		if err != nil {
			logrus.Errorf("could not retrieve new state for %s from Red Hat: %v", cve.ID, err)
			continue
		}

		for _, product := range products {
			pName := productName(product.RedHatMajorVersion.Int32)

			beginTx, err := i.db.Begin()
			if err != nil {
				logrus.Errorf("could not begin transaction: %v", err)
				continue
			}
			tx := i.db.UseTransaction(beginTx)

			skipCve := false
			if cveRh.AffectedRelease != nil {
				for _, state := range *cveRh.AffectedRelease {
					if state.ProductName == pName {
						st := secparseadminpb.AffectedProductState_FixedUpstream
						packageName := "TBD"
						if state.Package != nil {
							packageName = *state.Package

							match, err := i.checkForIgnoredPackage(ignoredUpstreamPackages, packageName)
							if err != nil {
								logrus.Errorf("Invalid glob: %v", err)
								continue
							}
							if match {
								st = secparseadminpb.AffectedProductState_UnknownProductState
							}
						} else {
							st = secparseadminpb.AffectedProductState_UnknownProductState
						}
						skipCve = i.checkProduct(tx, cve, shortCode, product, st, packageName, &state.Advisory)
						if skipCve {
							break
						}
					}
				}
			}
			if cveRh.PackageState != nil {
				for _, state := range *cveRh.PackageState {
					if state.ProductName == pName {
						pState := productState(state.FixState)
						packageName := "TBD"
						if state.PackageName != "" {
							packageName = state.PackageName

							match, err := i.checkForIgnoredPackage(ignoredUpstreamPackages, packageName)
							if err != nil {
								logrus.Errorf("Invalid glob: %v", err)
								continue
							}
							if match {
								pState = secparseadminpb.AffectedProductState_UnknownProductState
							}
						}
						skipCve = i.checkProduct(tx, cve, shortCode, product, pState, packageName, nil)
						if skipCve {
							break
						}
					}
				}
			}

			if skipCve {
				_ = beginTx.Rollback()
				continue
			}

			err = beginTx.Commit()
			if err != nil {
				logrus.Errorf("could not commit transaction: %v", err)
				continue
			}
		}
	}
}
