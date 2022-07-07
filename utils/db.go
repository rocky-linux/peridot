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

package utils

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"
)

type DbType int

const (
	DbPostgres DbType = iota
	DbUnknown
)

// GetDbType finds the database type specified in database.url
func GetDbType() DbType {
	dbURL := viper.GetString("database.url")
	if strings.HasPrefix(dbURL, "postgres") {
		return DbPostgres
	}

	return DbUnknown
}

// PgInit is the base pg init
func PgInit() *sql.DB {
	db, err := sql.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		log.Fatalln(err)
	}

	db.SetConnMaxLifetime(30 * time.Minute)

	return db
}

// PgInitx is the base pg init (sqlx)
func PgInitx() *sqlx.DB {
	db, err := sqlx.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		log.Fatalln(err)
	}

	db.SetConnMaxLifetime(30 * time.Minute)

	return db
}

func MinLimit(limit int32) int32 {
	if limit <= 0 || limit > 1000 {
		return 20
	}

	return limit
}

func MinPage(page int32) int32 {
	if page < 0 {
		return 0
	}

	return page
}

func GetOffset(page int32, limit int32) int64 {
	if page <= 0 {
		return 0
	}

	return int64(page * limit)
}

func UnlimitedLimit(limit int32) *int32 {
	if limit == -1 {
		return nil
	}

	return &limit
}
