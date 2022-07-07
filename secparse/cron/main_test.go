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
	"database/sql"
	"os"
	"peridot.resf.org/koji"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"peridot.resf.org/secparse/db/mock"
	"peridot.resf.org/secparse/rherrata"
	"peridot.resf.org/secparse/rhsecuritymock"
	"testing"
	"time"
)

var (
	cronInstance *Instance
	mockDb       *mock.Access
	securityMock *rhsecuritymock.Client
	errataMock   *rherrata.MockInstance
	kojiMock     *koji.Mock
)

func resetDb() {
	*mockDb = *mock.New()
	now := time.Now()

	mirrorFromDate, _ := time.Parse("2006-01-02", "2021-06-01")
	mockDb.ShortCodes = append(mockDb.ShortCodes, &db.ShortCode{
		Code:                "RL",
		Mode:                int8(secparseadminpb.ShortCodeMode_MirrorRedHatMode),
		CreatedAt:           &now,
		ArchivedAt:          sql.NullTime{},
		MirrorFromDate:      sql.NullTime{Valid: true, Time: mirrorFromDate},
		RedHatProductPrefix: sql.NullString{Valid: true, String: "Rocky Linux"},
	})
	mockDb.Products = append(mockDb.Products, &db.Product{
		ID:                 1,
		Name:               "Rocky Linux 8",
		CurrentFullVersion: "8.4",
		RedHatMajorVersion: sql.NullInt32{Valid: true, Int32: 8},
		ShortCode:          "RL",
		Archs:              []string{"x86_64", "aarch64"},
	})
}

func TestMain(m *testing.M) {
	mockDb = mock.New()
	securityMock = rhsecuritymock.New()
	errataMock = rherrata.NewMock()
	kojiMock = koji.NewMock()

	instance, _ := New(mockDb)
	instance.api = securityMock
	instance.errata = errataMock.API
	instance.koji = kojiMock

	cronInstance = instance

	resetDb()

	os.Exit(m.Run())
}
