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
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

func (c *Controller) DownstreamCVECheckActivity(ctx context.Context) error {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(10 * time.Second)
		}
	}()

	pendingProducts, err := c.db.GetPendingAffectedProducts()
	if err != nil {
		logrus.Errorf("could not get fixed cves: %v", err)
		return fmt.Errorf("could not get fixed cves")
	}

	for _, affectedProduct := range pendingProducts {
		if !affectedProduct.CveID.Valid {
			continue
		}

		err = func() error {
			willNotFixOnly := true
			allFixed := true

			switch affectedProduct.State {
			case
				int(apollopb.AffectedProduct_STATE_WILL_NOT_FIX_UPSTREAM),
				int(apollopb.AffectedProduct_STATE_OUT_OF_SUPPORT_SCOPE):
				return nil
			case
				int(apollopb.AffectedProduct_STATE_UNDER_INVESTIGATION_UPSTREAM),
				int(apollopb.AffectedProduct_STATE_AFFECTED_UPSTREAM):
				allFixed = false
				willNotFixOnly = false
				return nil
			}

			product, err := c.db.GetProductByID(affectedProduct.ProductID)
			if err != nil {
				logrus.Errorf("could not get product with id %d: %v", affectedProduct.ProductID, err)
				return err
			}

			ignoredUpstreamPackages, err := c.db.GetAllIgnoredPackagesByProductID(product.ID)
			if err != nil {
				logrus.Errorf("could not get ignored packages: %v", err)
				return err
			}

			beginTx, err := c.db.Begin()
			if err != nil {
				logrus.Errorf("could not begin transaction: %v", err)
				return err
			}
			tx := c.db.UseTransaction(beginTx)

			skipProduct := false
			defer func(skipProduct *bool, affectedProduct apollodb.AffectedProduct) {
				if *skipProduct {
					logrus.Infof("%s: Skipping package for now", affectedProduct.Package)
					_ = beginTx.Rollback()
				}
			}(&skipProduct, *affectedProduct)

			cve, err := c.db.GetCVEByID(affectedProduct.CveID.String)
			if err != nil {
				return err
			}

			nvrOnly := strings.Replace(affectedProduct.Package, ":", "-", 1)
			if rpmutils.Module().MatchString(nvrOnly) {
				if !affectedProduct.Advisory.Valid {
					skipProduct = true
				}

				redHatAdvisory, err := c.errata.GetErrata(affectedProduct.Advisory.String)
				if err != nil {
					logrus.Errorf("Could not get Red Hat Advisory: %v", err)
					skipProduct = true
				}

				for _, arch := range product.Archs {
					redHatProductName := affectedProductNameForArchAndVersion(arch, product.RedHatMajorVersion.Int32)
					affected := redHatAdvisory.AffectedProducts[redHatProductName]
					if affected == nil {
						continue
					}
					srpms := affected.SRPMs
					for _, srpm := range srpms {
						status := c.checkKojiForBuild(tx, ignoredUpstreamPackages, srpm, affectedProduct, cve)
						if status == apollopb.BuildStatus_BUILD_STATUS_SKIP {
							skipProduct = true
							break
						} else if status == apollopb.BuildStatus_BUILD_STATUS_FIXED {
							willNotFixOnly = false
						} else if status == apollopb.BuildStatus_BUILD_STATUS_NOT_FIXED {
							allFixed = false
							willNotFixOnly = false
						}
					}
					break
				}
				if skipProduct {
					logrus.Errorf("%s has not been fixed for NVR %s", cve.ID, nvrOnly)
				}
			} else {
				nvrOnly = rpmutils.Epoch().ReplaceAllString(affectedProduct.Package, "")
				status := c.checkKojiForBuild(tx, ignoredUpstreamPackages, nvrOnly, affectedProduct, cve)
				if status == apollopb.BuildStatus_BUILD_STATUS_SKIP {
					skipProduct = true
				} else if status == apollopb.BuildStatus_BUILD_STATUS_FIXED {
					willNotFixOnly = false
				} else if status == apollopb.BuildStatus_BUILD_STATUS_NOT_FIXED {
					allFixed = false
					willNotFixOnly = false
				}
			}

			if !skipProduct {
				newState := apollopb.AffectedProduct_STATE_FIXED_UPSTREAM
				if allFixed {
					newState = apollopb.AffectedProduct_STATE_FIXED_DOWNSTREAM
				}
				if willNotFixOnly {
					newState = apollopb.AffectedProduct_STATE_WILL_NOT_FIX_UPSTREAM
				}
				err := tx.UpdateAffectedProductStateAndPackageAndAdvisory(affectedProduct.ID, int(newState), affectedProduct.Package, utils.NullStringToPointer(affectedProduct.Advisory))
				if err != nil {
					logrus.Errorf("Could not save new CVE state: %v", err)
					return err
				}
				err = beginTx.Commit()
				if err != nil {
					logrus.Errorf("could not commit transaction: %v", err)
					return err
				}

				logrus.Infof("%s is now set to %s", cve.ID, newState.String())
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) DownstreamCVECheckWorkflow(ctx workflow.Context) error {
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 30 * time.Minute,
		StartToCloseTimeout:    6 * time.Hour,
		HeartbeatTimeout:       30 * time.Second,
	})
	return workflow.ExecuteActivity(activityCtx, c.DownstreamCVECheckActivity).Get(ctx, nil)
}
