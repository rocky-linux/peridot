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

package rherrata

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var internalAfterDates = map[string]string{
	"8.4": "2021-04-29T00:00:00Z",
	"9.0": "2022-05-17T00:00:00Z",
}

type CompactErrata struct {
	Name             string   `json:"id"`
	Description      string   `json:"portal_description"`
	Synopsis         string   `json:"portal_synopsis"`
	Severity         string   `json:"portal_severity"`
	Type             string   `json:"portal_advisory_type"`
	AffectedPackages []string `json:"portal_package"`
	CVEs             []string `json:"portal_CVE"`
	Fixes            []string `json:"portal_BZ"`
	PublicationDate  string   `json:"portal_publication_date"`
}

type internalAdvisoriesInnerResponse struct {
	Docs []*CompactErrata `json:"docs"`
}

type internalAdvisoriesResponse struct {
	Response *internalAdvisoriesInnerResponse `json:"response"`
}

func (a *API) GetAdvisories(currentVersion string, after *time.Time) ([]*CompactErrata, error) {
	req, err := a.newRequest("GET", a.baseURLAPI, nil)
	if err != nil {
		return nil, err
	}

	fq1 := "documentKind:(%22Errata%22)"
	usableVersion := strings.Replace(currentVersion, ".", "%5C.", -1)
	fq2 := fmt.Sprintf("portal_product_filter:Red%%5C+Hat%%5C+Enterprise%%5C+Linux%%7C*%%7C%s%%7C*", usableVersion)
	var fq3 string
	if after != nil {
		fq3 = "&fq=" + url.QueryEscape(fmt.Sprintf("portal_publication_date:[%s TO NOW]", after.Format(time.RFC3339)))
	} else if afterDate := internalAfterDates[currentVersion]; afterDate != "" {
		fq3 = "&fq=" + url.QueryEscape(fmt.Sprintf("portal_publication_date:[%s TO NOW]", afterDate))
	}
	req.URL.RawQuery = fmt.Sprintf("fq=%s&fq=%s%s&q=*:*&rows=10000&sort=portal_publication_date+desc&start=0", fq1, fq2, fq3)

	req.Header.Set("Accept", "application/json")

	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var marshalBody internalAdvisoriesResponse
	err = json.NewDecoder(res.Body).Decode(&marshalBody)
	if err != nil {
		return nil, err
	}

	return marshalBody.Response.Docs, nil
}
