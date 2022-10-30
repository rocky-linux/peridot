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
	"encoding/json"
	"go.temporal.io/sdk/testsuite"
	"os"
	apollodb "peridot.resf.org/apollo/db"
	apollomock "peridot.resf.org/apollo/db/mock"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rherrata"
	"peridot.resf.org/apollo/rhsecuritymock"
	"peridot.resf.org/koji"
	"testing"
	"time"
)

var (
	mockDb       *apollomock.Access
	securityMock *rhsecuritymock.Client
	errataMock   *rherrata.MockInstance
	kojiMock     *koji.Mock
	testWfSuite  *testsuite.WorkflowTestSuite
	controller   *Controller
)

func resetDb() {
	*mockDb = *apollomock.New()
	now := time.Now()

	mirrorFromDate, _ := time.Parse("2006-01-02", "2021-06-01")
	mockDb.ShortCodes = append(mockDb.ShortCodes, &apollodb.ShortCode{
		Code:       "RL",
		Mode:       int8(apollopb.ShortCode_MODE_MIRROR),
		CreatedAt:  &now,
		ArchivedAt: sql.NullTime{},
	})
	mockDb.Products = append(mockDb.Products, &apollodb.Product{
		ID:                  1,
		Name:                "Rocky Linux 8",
		CurrentFullVersion:  "8.4",
		RedHatMajorVersion:  sql.NullInt32{Valid: true, Int32: 8},
		ShortCode:           "RL",
		Archs:               []string{"x86_64", "aarch64"},
		MirrorFromDate:      sql.NullTime{Valid: true, Time: mirrorFromDate},
		RedHatProductPrefix: sql.NullString{Valid: true, String: "Rocky Linux"},
		BuildSystem:         "koji", // we're testing koji only for now
		BuildSystemEndpoint: "local",
		KojiCompose:         sql.NullString{Valid: true, String: "Rocky-8.4"},
		KojiModuleCompose:   sql.NullString{Valid: true, String: "Rocky-8.4-module"},
	})
}

func readTestDataJson(file string, target interface{}) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

func TestMain(m *testing.M) {
	mockDb = apollomock.New()
	securityMock = rhsecuritymock.New()
	errataMock = rherrata.NewMock()
	kojiMock = koji.NewMock()
	forceKoji = kojiMock

	testWfSuite = &testsuite.WorkflowTestSuite{}

	input := &NewControllerInput{
		Database: mockDb,
	}
	instance, err := NewController(
		input,
		WithSecurityAPI(securityMock),
		WithErrataAPI(errataMock.API),
	)
	if err != nil {
		panic(err.(any))
	}

	controller = instance

	resetDb()

	os.Exit(m.Run())
}
