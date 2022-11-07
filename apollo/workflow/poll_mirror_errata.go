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
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/utils"
	"strconv"
	"strings"
	"time"
)

func (c *Controller) processErrataShortCodeProduct(shortCode *apollodb.ShortCode, product *apollodb.Product) error {
	if !product.RedHatMajorVersion.Valid {
		return nil
	}
	if !strings.HasPrefix(product.Name, product.RedHatProductPrefix.String) {
		return nil
	}

	ignoredUpstreamPackages, err := c.db.GetAllIgnoredPackagesByProductID(product.ID)
	if err != nil {
		logrus.Errorf("could not get ignored packages: %v", err)
		return fmt.Errorf("could not get ignored packages")
	}

	var lastSync *time.Time
	mirrorState, err := c.db.GetMirrorState(shortCode.Code)
	if err == nil {
		if mirrorState.ErrataAfter.Valid {
			lastSync = &mirrorState.ErrataAfter.Time
		}
	}

	advisories, err := c.errata.GetAdvisories(product.CurrentFullVersion, lastSync)
	if err != nil {
		logrus.Errorf("Could not get Red Hat Advisories: %v", err)
		return fmt.Errorf("could not get Red Hat Advisories")
	}

	var newLastSync *time.Time

	parentBeginTx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	parentTx := c.db.UseTransaction(parentBeginTx)
	rollbackParent := true
	defer func() {
		if rollbackParent {
			_ = parentBeginTx.Rollback()
		}
	}()

	for _, advisory := range advisories {
		if newLastSync == nil {
			parsedTime, err := time.Parse(time.RFC3339, advisory.PublicationDate)
			if err == nil {
				newLastSync = &parsedTime
				_ = parentTx.UpdateMirrorStateErrata(shortCode.Code, newLastSync)
			}
		}

		advisoryId := rpmutils.AdvisoryId().FindStringSubmatch(advisory.Name)
		if len(advisoryId) < 5 {
			logrus.Errorf("Invalid advisory %s", advisory.Name)
			return nil
		}
		code := advisoryId[1]
		year, err := strconv.Atoi(advisoryId[3])
		if err != nil {
			logrus.Errorf("Invalid advisory %s", advisory.Name)
			return nil
		}
		num, err := strconv.Atoi(advisoryId[4])
		if err != nil {
			logrus.Errorf("Invalid advisory %s", advisory.Name)
			return nil
		}

		beginTx, err := c.db.Begin()
		if err != nil {
			logrus.Errorf("Could not begin tx: %v", err)
			return fmt.Errorf("could not begin tx")
		}
		tx := c.db.UseTransaction(beginTx)

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
							_ = beginTx.Rollback()
							return fmt.Errorf("an unknown error occurred")
						}

						sourceBy := "Red Hat"
						resourceUrl := fmt.Sprintf("https://access.redhat.com/hydra/rest/securitydata/cve/%s.json", cve)

						cveRh, _, err := c.security.GetCveExecute(c.security.GetCve(context.TODO(), cve))
						if err != nil {
							return errors.Wrap(err, "could not get cve")
						}
						cveBytes, err := json.Marshal(cveRh)
						if err != nil {
							return fmt.Errorf("could not marshal cve: %v", err)
						}

						_, err = tx.CreateCVE(cve, shortCode.Code, &sourceBy, &resourceUrl, types.NullJSONText{Valid: true, JSONText: cveBytes})
						if err != nil {
							logrus.Errorf("could not create cve: %v", err)
							_ = beginTx.Rollback()
							return fmt.Errorf("could not create cve")
						}
						logrus.Infof("Added %s to %s (%s)", cve, shortCode.Code, advisory.Name)
					}
				} else if strings.HasPrefix(advisory.Name, "RHBA") || strings.HasPrefix(advisory.Name, "RHEA") {
					_, err := tx.GetAffectedProductByAdvisory(advisory.Name)
					if err != nil {
						if err == sql.ErrNoRows {
							_, err := tx.GetCVEByID(advisory.Name)
							if err == nil {
								continue
							}
							if err != sql.ErrNoRows {
								logrus.Errorf("an unknown error occurred: %v", err)
								_ = beginTx.Rollback()
								return fmt.Errorf("an unknown error occurred")
							}

							sourceBy := "Red Hat"
							resourceUrl := fmt.Sprintf("https://access.redhat.com/errata/%s", advisory.Name)
							_, err = tx.CreateCVE(advisory.Name, product.ShortCode, &sourceBy, &resourceUrl, types.NullJSONText{})
							if err != nil {
								_ = beginTx.Rollback()
								return fmt.Errorf("could not create cve: %v", err)
							}

							for _, srpm := range advisory.AffectedPackages {
								if !strings.Contains(srpm, ".src.rpm") {
									continue
								}

								pkg := strings.Replace(srpm, ".src.rpm", "", 1)

								nvr := rpmutils.NVR().FindStringSubmatch(pkg)
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
								_, err := tx.CreateAffectedProduct(product.ID, advisory.Name, int(apollopb.AffectedProduct_STATE_FIXED_UPSTREAM), product.CurrentFullVersion, pkg, &advisory.Name)
								if err != nil {
									_ = beginTx.Rollback()
									return fmt.Errorf("could not create affected product for srpm: %v", err)
								}
							}
							logrus.Infof("Added %s to %s", advisory.Name, shortCode.Code)
						} else {
							_ = beginTx.Rollback()
							return fmt.Errorf("Could not get affected product by advisory: %v", err)
						}
					}
				}
			} else {
				_ = beginTx.Rollback()
				logrus.Errorf("Could not fetch advisory: %v", err)
				return err
			}
		}

		err = beginTx.Commit()
		if err != nil {
			logrus.Errorf("Could not commit new advisory tx: %v", err)
			return err
		}
	}

	rollbackParent = false
	err = parentBeginTx.Commit()
	if err != nil {
		logrus.Errorf("Could not commit parent tx: %v", err)
		return err
	}

	return nil
}

func (c *Controller) ProcessRedHatErrataShortCodeActivity(ctx context.Context, shortCode *apollodb.ShortCode) error {
	go func() {
		for {
			activity.RecordHeartbeat(ctx)
			time.Sleep(10 * time.Second)
		}
	}()

	if int32(shortCode.Mode) != int32(apollopb.ShortCode_MODE_MIRROR) {
		// This is not a mirrored short code, continue
		return nil
	}

	allProducts, err := c.db.GetProductsByShortCode(shortCode.Code)
	if err != nil {
		logrus.Errorf("could not get all products for code %s: %v", shortCode.Code, err)
		return fmt.Errorf("could not get all products for code %s", shortCode.Code)
	}

	for _, product := range allProducts {
		err := c.processErrataShortCodeProduct(shortCode, product)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) PollRedHatErrataWorkflow(ctx workflow.Context) error {
	shortCodeCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	})
	var shortCodeRes ShortCodesRes
	err := workflow.ExecuteActivity(shortCodeCtx, c.GetAllShortCodesActivity).Get(ctx, &shortCodeRes)
	if err != nil {
		return err
	}

	var futures []workflow.Future
	for _, shortCode := range shortCodeRes.ShortCodes {
		activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: 5 * time.Minute,
			StartToCloseTimeout:    12 * time.Hour,
			HeartbeatTimeout:       30 * time.Second,
		})
		futures = append(futures, workflow.ExecuteActivity(activityCtx, c.ProcessRedHatErrataShortCodeActivity, shortCode))
	}

	for _, future := range futures {
		err := future.Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
