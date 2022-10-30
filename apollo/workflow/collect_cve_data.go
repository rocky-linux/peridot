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
	"encoding/json"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

func (c *Controller) CollectCVEDataActivity(ctx context.Context) error {
	cves, err := c.db.GetAllCVEs()
	if err != nil {
		return errors.Wrap(err, "could not get cves")
	}

	// Go through each CVE and set CVE content by fetching from rhsecurity
	for _, cve := range cves {
		if cve.Content.Valid {
			continue
		}
		if !strings.HasPrefix(cve.ID, "CVE") {
			continue
		}

		cveRh, _, err := c.security.GetCveExecute(c.security.GetCve(ctx, cve.ID))
		if err != nil {
			return errors.Wrap(err, "could not get cve")
		}

		cveBytes, err := json.Marshal(cveRh)
		if err != nil {
			return errors.Wrap(err, "could not marshal cve")
		}
		err = c.db.SetCVEContent(cve.ID, cveBytes)
		if err != nil {
			return errors.Wrap(err, "could not set cve content")
		}
	}

	return nil
}

func (c *Controller) CollectCVEDataWorkflow(ctx workflow.Context) error {
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    12 * time.Hour,
	})
	return workflow.ExecuteActivity(activityCtx, c.CollectCVEDataActivity).Get(ctx, nil)
}
