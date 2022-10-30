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
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httptest"
)

type MockInstance struct {
	API *API
	// Mapped to advisory id/name
	HTMLResponses    map[string]string
	Advisories       *internalAdvisoriesResponse
	TestServerErrata *httptest.Server
	TestServerMock   *httptest.Server
}

func NewMock() *MockInstance {
	mockInstance := &MockInstance{
		HTMLResponses: map[string]string{},
		Advisories: &internalAdvisoriesResponse{
			Response: &internalAdvisoriesInnerResponse{
				Docs: []*CompactErrata{},
			},
		},
	}

	muxErrata := chi.NewMux()
	muxErrata.Get("/{advisory}", func(w http.ResponseWriter, r *http.Request) {
		advisory := chi.URLParam(r, "advisory")

		if response := mockInstance.HTMLResponses[advisory]; response != "" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	muxAPI := chi.NewMux()
	muxAPI.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(mockInstance.Advisories)
	})

	api := NewClient()
	tsErrata := httptest.NewServer(muxErrata)
	tsAPI := httptest.NewServer(muxAPI)
	api.baseURLErrata = tsErrata.URL
	api.baseURLAPI = tsAPI.URL

	mockInstance.API = api
	mockInstance.TestServerErrata = tsErrata

	return mockInstance
}
