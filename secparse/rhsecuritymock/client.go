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

package rhsecuritymock

import (
	_context "context"
	_nethttp "net/http"
	"peridot.resf.org/secparse/rhsecurity"
	"peridot.resf.org/utils"
)

type Client struct {
	orig rhsecurity.DefaultApi

	ActiveCVE *rhsecurity.CVEDetailed
	Cves      []*rhsecurity.CVE
}

func New() *Client {
	return &Client{
		orig: rhsecurity.NewAPIClient(rhsecurity.NewConfiguration()).DefaultApi,
	}
}

/*
 * GetCve Get specific CVE
 * Retrieve full CVE details
 * @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param cVE
 * @return ApiGetCveRequest
 */
func (c *Client) GetCve(ctx _context.Context, cVE string) rhsecurity.ApiGetCveRequest {
	return c.orig.GetCve(ctx, cVE)
}

/*
 * GetCveExecute executes the request
 * @return CVEDetailed
 */
func (c *Client) GetCveExecute(_ rhsecurity.ApiGetCveRequest) (rhsecurity.CVEDetailed, *_nethttp.Response, error) {
	if c.ActiveCVE != nil {
		return *c.ActiveCVE, &_nethttp.Response{}, nil
	}

	return rhsecurity.CVEDetailed{}, nil, utils.CouldNotFindObject
}

/*
 * GetCves Get CVEs
 * List all the recent CVEs when no parameter is passed. Returns a convenience object as response with very minimum attributes.
 * @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @return ApiGetCvesRequest
 */
func (c *Client) GetCves(ctx _context.Context) rhsecurity.ApiGetCvesRequest {
	return c.orig.GetCves(ctx)
}

/*
 * GetCvesExecute executes the request
 * @return []CVE
 */
func (c *Client) GetCvesExecute(_ rhsecurity.ApiGetCvesRequest) ([]rhsecurity.CVE, *_nethttp.Response, error) {
	var cves []rhsecurity.CVE
	for _, cve := range c.Cves {
		cves = append(cves, *cve)
	}

	return cves, &_nethttp.Response{}, nil
}
