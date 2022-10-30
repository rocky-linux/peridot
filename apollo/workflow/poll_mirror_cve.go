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
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

type ShortCodesRes struct {
	ShortCodes []*apollodb.ShortCode `json:"short_codes"`
}

func (c *Controller) pollCVEProcessProduct(ctx context.Context, product *apollodb.Product, shortCode *apollodb.ShortCode) error {
	// Skip if the product doesn't define a valid Red Hat version
	if !product.RedHatMajorVersion.Valid {
		return nil
	}
	// Skip if product doesn't have correct Red Hat prefix
	if !strings.HasPrefix(product.Name, product.RedHatProductPrefix.String) {
		return nil
	}

	var lastSync *time.Time
	mirrorState, err := c.db.GetMirrorState(shortCode.Code)
	if err != nil {
		if err != sql.ErrNoRows {
			c.log.Errorf("could not get last sync for code %s: %v", shortCode.Code, err)
			// The cron will retry this
			return nil
		}
	} else {
		if mirrorState != nil && mirrorState.LastSync.Valid {
			lastSync = &mirrorState.LastSync.Time
		}
	}
	if lastSync == nil {
		now := time.Now()
		if product.MirrorFromDate.Valid {
			now = product.MirrorFromDate.Time
		}
		lastSync = &now
	}

	req := c.security.GetCves(ctx)
	req = req.Product(productName(product.RedHatMajorVersion.Int32))
	if lastSync != nil {
		req = req.After(lastSync.Format("2006-01-02"))
	}

	page := 1
	for {
		reqNew := req.Page(float32(page))
		cves, _, err := c.security.GetCvesExecute(reqNew)
		if err != nil {
			c.log.Errorf("could not get cves: %v", err)
			return fmt.Errorf("could not get cves")
		}
		if len(cves) == 0 {
			break
		}

		for _, cve := range cves {
			_, err := c.db.GetCVEByID(cve.CVE)
			if err == nil {
				continue
			}
			if err != sql.ErrNoRows {
				c.log.Errorf("an unknown error occurred: %v", err)
				return fmt.Errorf("an unknown error occurred")
			}

			cveRh, _, err := c.security.GetCveExecute(c.security.GetCve(ctx, cve.CVE))
			if err != nil {
				return errors.Wrap(err, "could not get cve")
			}
			cveBytes, err := json.Marshal(cveRh)
			if err != nil {
				return fmt.Errorf("could not marshal cve: %v", err)
			}

			sourceBy := "Red Hat"
			_, err = c.db.CreateCVE(cve.CVE, shortCode.Code, &sourceBy, &cve.ResourceUrl, types.NullJSONText{Valid: true, JSONText: cveBytes})
			if err != nil {
				c.log.Errorf("could not create cve: %v", err)
				return fmt.Errorf("could not create cve")
			}
			c.log.Infof("Added %s to %s with state NewFromUpstream", cve.CVE, shortCode.Code)
		}
		page++
	}

	err = c.db.UpdateMirrorState(shortCode.Code, utils.Pointer[time.Time](time.Now()))
	if err != nil {
		c.log.Errorf("could not update mirroring state: %v", err)
	}

	return nil
}

func (c *Controller) PollCVEProcessShortCodeActivity(ctx context.Context, shortCode *apollodb.ShortCode) error {
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
		c.log.Errorf("could not get all products for code %s: %v", shortCode.Code, err)
		// Returning nil since the cron will retry this
		// We can set up an alert on the Grafana side to alert us
		// if this happens too often
		return nil
	}

	for _, product := range allProducts {
		err := c.pollCVEProcessProduct(ctx, product, shortCode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) GetAllShortCodesActivity() (*ShortCodesRes, error) {
	s, err := c.db.GetAllShortCodes()
	if err != nil {
		return nil, err
	}

	return &ShortCodesRes{
		ShortCodes: s,
	}, nil
}

func (c *Controller) PollRedHatCVEsWorkflow(ctx workflow.Context) error {
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
		futures = append(futures, workflow.ExecuteActivity(activityCtx, c.PollCVEProcessShortCodeActivity, shortCode))
	}

	for _, future := range futures {
		err := future.Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
