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
	"github.com/sirupsen/logrus"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/utils"
	"strconv"
	"strings"
)

func (i *Instance) ScanRedHatErrata() {
	shortCodes, err := i.db.GetAllShortCodes()
	if err != nil {
		logrus.Errorf("could not get short codes: %v", err)
		return
	}

	ignoredPackagesBuffer := map[string][]string{}

	for _, shortCode := range shortCodes {
		if int32(shortCode.Mode) != int32(secparseadminpb.ShortCodeMode_MirrorRedHatMode) {
			// This is not a mirrored short code, continue
			continue
		}

		if ignoredPackagesBuffer[shortCode.Code] == nil {
			ignoredUpstreamPackages, err := i.db.GetAllIgnoredPackagesByShortCode(shortCode.Code)
			if err != nil {
				logrus.Errorf("could not get ignored packages: %v", err)
				continue
			}
			ignoredPackagesBuffer[shortCode.Code] = ignoredUpstreamPackages
		}
		ignoredUpstreamPackages := ignoredPackagesBuffer[shortCode.Code]

		allProducts, err := i.db.GetProductsByShortCode(shortCode.Code)
		if err != nil {
			logrus.Errorf("could not get all products for code %s: %v", shortCode.Code, err)
			continue
		}

		for _, product := range allProducts {
			if !product.RedHatMajorVersion.Valid {
				continue
			}
			if !strings.HasPrefix(product.Name, shortCode.RedHatProductPrefix.String) {
				continue
			}

			advisories, err := i.errata.GetAdvisories(product.CurrentFullVersion)
			if err != nil {
				logrus.Errorf("Could not get Red Hat Advisories: %v", err)
				continue
			}

			for _, advisory := range advisories {
				advisoryId := i.advisoryIdRegex.FindStringSubmatch(advisory.Name)
				if len(advisoryId) < 5 {
					logrus.Errorf("Invalid advisory %s", advisory.Name)
					continue
				}
				code := advisoryId[1]
				year, err := strconv.Atoi(advisoryId[3])
				if err != nil {
					logrus.Errorf("Invalid advisory %s", advisory.Name)
					continue
				}
				num, err := strconv.Atoi(advisoryId[4])
				if err != nil {
					logrus.Errorf("Invalid advisory %s", advisory.Name)
					continue
				}

				beginTx, err := i.db.Begin()
				if err != nil {
					logrus.Errorf("Could not begin tx: %v", err)
					continue
				}
				tx := i.db.UseTransaction(beginTx)

				_, err = tx.GetAdvisoryByCodeAndYearAndNum(code, year, num)
				if err != nil {
					if err == sql.ErrNoRows {
						// If security then just add CVEs, the rest should be automatic
						if strings.HasPrefix(advisory.Name, "RHSA") {
							for _, cve := range advisory.CVEs {
								_, err := tx.GetCVEByID(cve)
								if err == nil {
									continue
								}
								if err != sql.ErrNoRows {
									logrus.Errorf("an unknown error occurred: %v", err)
									return
								}

								sourceBy := "Red Hat"
								resourceUrl := fmt.Sprintf("https://access.redhat.com/hydra/rest/securitydata/cve/%s.json", cve)
								_, err = tx.CreateCVE(cve, secparseadminpb.CVEState_NewFromUpstream, shortCode.Code, &sourceBy, &resourceUrl)
								if err != nil {
									logrus.Errorf("could not create cve: %v", err)
									_ = beginTx.Rollback()
									return
								}
								logrus.Infof("Added %s to %s (%s)", cve, shortCode.Code, advisory.Name)
							}
						} else if strings.HasPrefix(advisory.Name, "RHBA") || strings.HasPrefix(advisory.Name, "RHEA") {
							doRollback := false
							_, err := tx.GetAffectedProductByAdvisory(advisory.Name)
							if err != nil {
								if err == sql.ErrNoRows {
									_, err := tx.GetCVEByID(advisory.Name)
									if err == nil {
										continue
									}
									if err != sql.ErrNoRows {
										logrus.Errorf("an unknown error occurred: %v", err)
										return
									}

									sourceBy := "Red Hat"
									resourceUrl := fmt.Sprintf("https://access.redhat.com/errata/%s", advisory.Name)
									_, err = tx.CreateCVE(advisory.Name, secparseadminpb.CVEState_ResolvedUpstream, product.ShortCode, &sourceBy, &resourceUrl)
									if err != nil {
										logrus.Errorf("Could not create cve: %v", err)
										_ = beginTx.Rollback()
										continue
									}

									for _, srpm := range advisory.AffectedPackages {
										if !strings.Contains(srpm, ".src.rpm") {
											continue
										}

										pkg := strings.Replace(srpm, ".src.rpm", "", 1)

										nvr := i.nvr.FindStringSubmatch(pkg)
										var packageName string
										if len(nvr) >= 2 {
											packageName = nvr[1]
										} else {
											packageName = pkg
										}
										if utils.StrContains(packageName, ignoredUpstreamPackages) {
											continue
										}
										dist := fmt.Sprintf("el%d", product.RedHatMajorVersion.Int32)
										if !strings.Contains(pkg, dist) {
											continue
										}
										if strings.Contains(pkg, dist+"sat") {
											continue
										}
										_, err := tx.CreateAffectedProduct(product.ID, advisory.Name, int(secparseadminpb.AffectedProductState_FixedUpstream), product.CurrentFullVersion, pkg, &advisory.Name)
										if err != nil {
											logrus.Errorf("Could not create affected product for srpm: %v", err)
											doRollback = true
											break
										}
									}
									if doRollback {
										_ = beginTx.Rollback()
										continue
									}
									logrus.Infof("Added %s to %s", advisory.Name, shortCode.Code)
								} else {
									logrus.Errorf("Could not get affected product by advisory: %v", err)
									continue
								}
							}
						}
					} else {
						logrus.Errorf("Could not fetch advisory: %v", err)
						continue
					}
				}

				err = beginTx.Commit()
				if err != nil {
					logrus.Errorf("Could not commit new advisory tx: %v", err)
					continue
				}
			}
		}
	}
}
