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
	"database/sql"
	"github.com/sirupsen/logrus"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"strings"
	"time"
)

func (i *Instance) PollRedHatForNewCVEs() {
	ctx := context.TODO()

	shortCodes, err := i.db.GetAllShortCodes()
	if err != nil {
		logrus.Errorf("could not get short codes: %v", err)
		return
	}
	for _, shortCode := range shortCodes {
		if int32(shortCode.Mode) != int32(secparseadminpb.ShortCodeMode_MirrorRedHatMode) {
			// This is not a mirrored short code, continue
			continue
		}

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

			lastSync, err := i.db.GetMirrorStateLastSync(shortCode.Code)
			if err != nil {
				if err != sql.ErrNoRows {
					logrus.Errorf("could not get last sync for code %s: %v", shortCode.Code, err)
					continue
				}

				now := time.Now()
				if shortCode.MirrorFromDate.Valid {
					now = shortCode.MirrorFromDate.Time
				}
				lastSync = &now
			}

			req := i.api.GetCves(ctx)
			req = req.Product(productName(product.RedHatMajorVersion.Int32))
			if lastSync != nil {
				req = req.After(lastSync.Format("2006-01-02"))
			}

			cves, _, err := i.api.GetCvesExecute(req)
			if err != nil {
				logrus.Errorf("could not get cves: %v", err)
				return
			}

			for _, cve := range cves {
				_, err := i.db.GetCVEByID(cve.CVE)
				if err == nil {
					continue
				}
				if err != sql.ErrNoRows {
					logrus.Errorf("an unknown error occurred: %v", err)
					return
				}

				sourceBy := "Red Hat"
				_, err = i.db.CreateCVE(cve.CVE, secparseadminpb.CVEState_NewFromUpstream, shortCode.Code, &sourceBy, &cve.ResourceUrl)
				if err != nil {
					logrus.Errorf("could not create cve: %v", err)
					return
				}
				logrus.Infof("Added %s to %s with state NewFromUpstream", cve.CVE, shortCode.Code)
			}

			now := time.Now()
			err = i.db.UpdateMirrorState(shortCode.Code, &now)
			if err != nil {
				logrus.Errorf("could not update mirroring state: %v", err)
			}
		}
	}
}
