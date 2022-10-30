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
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/workflow"
	apollodb "peridot.resf.org/apollo/db"
	"peridot.resf.org/apollo/rpmutils"
	"strconv"
	"strings"
	"time"
)

func (c *Controller) AutoCreateAdvisoryActivity() error {
	cves, err := c.db.GetAllCVEsFixedDownstream()
	if err != nil {
		c.log.Errorf("Could not get cves fixed downstream: %v", err)
		return fmt.Errorf("could not get cves fixed downstream")
	}

	rebootSuggestedPackages, err := c.db.GetAllRebootSuggestedPackages()
	if err != nil {
		c.log.Errorf("Could not get reboot suggested packages: %v", err)
		return fmt.Errorf("could not get reboot suggested packages")
	}

	for _, cve := range cves {
		beginTx, err := c.db.Begin()
		if err != nil {
			logrus.Errorf("could not begin tx: %v", err)
			continue
		}
		tx := c.db.UseTransaction(beginTx)

		affectedProducts, err := tx.GetAllAffectedProductsByCVE(cve.ID)
		if err != nil {
			logrus.Errorf("Could not get affected products for %s: %v", cve.ID, err)
			_ = beginTx.Rollback()
			continue
		}

		var existingAdvisory *apollodb.Advisory
		didSkip := false

		for _, affectedProduct := range affectedProducts {
			if !affectedProduct.Advisory.Valid {
				continue
			}
			product, err := c.db.GetProductByID(affectedProduct.ProductID)
			if err != nil {
				c.log.Errorf("could not get product by id %d: %v", affectedProduct.ProductID, err)
				return fmt.Errorf("could not get product by id %d", affectedProduct.ProductID)
			}

			advisorySplit := strings.Split(affectedProduct.Advisory.String, "-")
			numYearSplit := strings.Split(advisorySplit[1], ":")

			year, err := strconv.Atoi(numYearSplit[0])
			if err != nil {
				logrus.Errorf("invalid year %s", numYearSplit[0])
				continue
			}
			num, err := strconv.Atoi(numYearSplit[1])
			if err != nil {
				logrus.Errorf("invalid num %s", numYearSplit[1])
				continue
			}

			existingAdvisory, err = tx.GetAdvisoryByCodeAndYearAndNum(product.ShortCode, year, num)
			if err != nil {
				if err == sql.ErrNoRows {
					errata, err := c.errata.GetErrata(affectedProduct.Advisory.String)
					if err != nil {
						logrus.Errorf("could not get errata from Red Hat: %v", err)
						didSkip = true
						break
					}

					for i, _ := range errata.Topic {
						errata.Topic[i] = strings.Replace(errata.Topic[i], "Red Hat Enterprise Linux", product.RedHatProductPrefix.String, -1)
						errata.Topic[i] = strings.Replace(errata.Topic[i], "Red Hat", c.vendor, -1)
					}
					for i, _ := range errata.Description {
						errata.Description[i] = strings.Replace(errata.Description[i], "Red Hat Enterprise Linux", product.RedHatProductPrefix.String, -1)
						errata.Description[i] = strings.Replace(errata.Description[i], "Red Hat", c.vendor, -1)
					}

					var packageName string
					if !rpmutils.Module().MatchString(affectedProduct.Package) {
						packageName = affectedProduct.Package
					}

					newAdvisory := &apollodb.Advisory{
						Year:           year,
						Num:            num,
						Synopsis:       errata.Synopsis,
						Topic:          strings.Join(errata.Topic, "\n"),
						Severity:       int(errata.Severity),
						Type:           int(errata.Type),
						Description:    strings.Join(errata.Description, "\n"),
						RedHatIssuedAt: sql.NullTime{Valid: true, Time: errata.IssuedAt},
						ShortCodeCode:  product.ShortCode,
						PublishedAt:    sql.NullTime{},
					}
					if packageName != "" {
						match, err := c.checkForRebootSuggestedPackage(rebootSuggestedPackages, packageName)
						if err != nil {
							c.log.Errorf("could not check for reboot suggested package: %v", err)
							return fmt.Errorf("could not check for reboot suggested package")
						}
						if match {
							newAdvisory.RebootSuggested = true
						}
					}

					newAdvisory, err = tx.CreateAdvisory(newAdvisory)
					if err != nil {
						logrus.Errorf("Could not create new advisory: %v", err)
						didSkip = true
						break
					}
					existingAdvisory = newAdvisory

					for _, fix := range errata.Fixes {
						sourceLink := fmt.Sprintf("https://bugzilla.redhat.com/show_bug.cgi?id=%s", fix.BugzillaID)
						id, err := tx.CreateFix(fix.BugzillaID, "Red Hat", sourceLink, fix.Description)
						if err != nil {
							logrus.Errorf("Could not create fix for BZ#%s", fix.BugzillaID)
							didSkip = true
							break
						}
						err = tx.AddAdvisoryFix(existingAdvisory.ID, id)
						if err != nil {
							logrus.Errorf("Could not add fix BZ#%s to advisory %d", fix.BugzillaID, existingAdvisory.ID)
							didSkip = true
							break
						}
					}
					if didSkip {
						break
					}
					for _, reference := range errata.References {
						// Skip redhat.com references
						if strings.Contains(reference, "redhat.com") {
							continue
						}

						err := tx.CreateAdvisoryReference(existingAdvisory.ID, reference)
						if err != nil {
							logrus.Errorf("Could not reference %s", reference)
							didSkip = true
							break
						}
					}
					if didSkip {
						break
					}
				} else {
					logrus.Errorf("Could not reach database to retrieve advisory: %v", err)
					didSkip = true
					break
				}
			}

			if didSkip {
				break
			}

			err = tx.AddAdvisoryCVE(existingAdvisory.ID, cve.ID)
			if err != nil {
				logrus.Errorf("Could not add %s to advisory %d", cve.ID, existingAdvisory.ID)
				didSkip = true
				break
			}
		}

		if didSkip {
			_ = beginTx.Rollback()
			continue
		}

		err = beginTx.Commit()
		if err != nil {
			logrus.Errorf("Could not commit database transaction: %v", err)
			continue
		}

		logrus.Infof("Created/updated advisory %s-%d:%d for %s", existingAdvisory.ShortCodeCode, existingAdvisory.Year, existingAdvisory.Num, cve.ID)
	}

	return nil
}

func (c *Controller) AutoCreateAdvisoryWorkflow(ctx workflow.Context) error {
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    time.Hour,
	})
	return workflow.ExecuteActivity(activityCtx, c.AutoCreateAdvisoryActivity).Get(ctx, nil)
}
