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
	"strings"
	"time"
)

func (c *Controller) UpdateCVEStateActivity(ctx context.Context) error {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(10 * time.Second)
		}
	}()

	cves, err := c.db.GetAllUnresolvedCVEs()
	if err != nil {
		c.log.Errorf("could not get unresolved cves: %v", err)
		return fmt.Errorf("could not get unresolved cves")
	}

	shortCodeBuffer := map[string]*apollodb.ShortCode{}
	productBuffer := map[string][]*apollodb.Product{}
	ignoredPackagesBuffer := map[int64][]string{}

	for _, cve := range cves {
		if !strings.HasPrefix(cve.ID, "CVE") {
			continue
		}

		if shortCodeBuffer[cve.ShortCode] == nil {
			shortCode, err := c.db.GetShortCodeByCode(cve.ShortCode)
			if err != nil {
				logrus.Errorf("could not get short code: %v", err)
				continue
			}

			shortCodeBuffer[shortCode.Code] = shortCode
		}
		shortCode := shortCodeBuffer[cve.ShortCode]

		if productBuffer[shortCode.Code] == nil {
			products, err := c.db.GetProductsByShortCode(shortCode.Code)
			if err != nil {
				logrus.Errorf("could not get products for code: %s: %v", shortCode.Code, err)
				continue
			}
			productBuffer[shortCode.Code] = products
		}
		products := productBuffer[shortCode.Code]

		// Please do not simplify next statement
		// During testing we're mocking pagination as well, and this is the
		// easiest way to "wrap" and represent a new request restarting it from page 1
		cveRh, _, err := c.security.GetCveExecute(c.security.GetCve(ctx, cve.ID))
		if err != nil {
			logrus.Errorf("could not retrieve new state for %s from Red Hat: %v", cve.ID, err)
			continue
		}

		for _, product := range products {
			if ignoredPackagesBuffer[product.ID] == nil {
				ignoredUpstreamPackages, err := c.db.GetAllIgnoredPackagesByProductID(product.ID)
				if err != nil {
					logrus.Errorf("could not get ignored packages: %v", err)
					continue
				}
				ignoredPackagesBuffer[product.ID] = ignoredUpstreamPackages
			}
			ignoredUpstreamPackages := ignoredPackagesBuffer[product.ID]

			pName := productName(product.RedHatMajorVersion.Int32)

			beginTx, err := c.db.Begin()
			if err != nil {
				c.log.Errorf("could not begin transaction: %v", err)
				continue
			}
			tx := c.db.UseTransaction(beginTx)

			skipCve := false
			defer func() {
				if skipCve {
					_ = beginTx.Rollback()
				}
			}()

			if cveRh.AffectedRelease != nil {
				for _, state := range *cveRh.AffectedRelease {
					if (product.Cpe.Valid && state.Cpe == product.Cpe.String) || state.ProductName == pName {
						st := apollopb.AffectedProduct_STATE_FIXED_UPSTREAM
						packageName := "TBD"
						if state.Package != nil {
							packageName = *state.Package

							match, err := c.checkForIgnoredPackage(ignoredUpstreamPackages, packageName)
							if err != nil {
								c.log.Errorf("Invalid glob: %v", err)
								continue
							}
							if match {
								st = apollopb.AffectedProduct_STATE_UNKNOWN
							}
						} else {
							st = apollopb.AffectedProduct_STATE_UNKNOWN
						}
						skipCve = c.checkProduct(tx, cve, shortCode, product, st, packageName, &state.Advisory)
						if skipCve {
							break
						}
					}
				}
			}
			if cveRh.PackageState != nil {
				for _, state := range *cveRh.PackageState {
					if (product.Cpe.Valid && state.Cpe == product.Cpe.String) || state.ProductName == pName {
						pState := productState(state.FixState)
						packageName := "TBD"
						if state.PackageName != "" {
							packageName = state.PackageName

							match, err := c.checkForIgnoredPackage(ignoredUpstreamPackages, packageName)
							if err != nil {
								c.log.Errorf("Invalid glob: %v", err)
								continue
							}
							if match {
								pState = apollopb.AffectedProduct_STATE_UNKNOWN
							}
						}
						skipCve = c.checkProduct(tx, cve, shortCode, product, pState, packageName, nil)
						if skipCve {
							break
						}
					}
				}
			}

			err = beginTx.Commit()
			if err != nil {
				c.log.Errorf("could not commit transaction: %v", err)
				continue
			}
		}
	}

	return nil
}

func (c *Controller) UpdateCVEStateWorkflow(ctx workflow.Context) error {
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    12 * time.Hour,
		HeartbeatTimeout:       30 * time.Second,
	})
	return workflow.ExecuteActivity(activityCtx, c.UpdateCVEStateActivity).Get(ctx, nil)
}
