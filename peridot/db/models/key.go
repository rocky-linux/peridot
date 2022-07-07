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

package models

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Key struct {
	ID         uuid.UUID    `json:"id" db:"id"`
	CreatedAt  time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt  sql.NullTime `json:"updatedAt" db:"updated_at"`
	BreachedAt sql.NullTime `json:"breachedAt" db:"breached_at"`

	Name      string         `json:"name" db:"name"`
	Email     string         `json:"email" db:"email"`
	GpgId     string         `json:"gpgId" db:"gpg_id"`
	EncKey    string         `json:"encKey" db:"enc_key"`
	RotEncKey sql.NullString `json:"rotEncKey" db:"rot_enc_key"`
	Nonce     string         `json:"nonce" db:"nonce"`
	RotNonce  sql.NullString `json:"rotNonce" db:"rot_nonce"`

	PublicKey string `json:"publicKey" db:"public_key"`

	ExtStoreType    string         `json:"extStoreType" db:"ext_store_type"`
	RotExtStoreType sql.NullString `json:"rotExtStoreType" db:"rot_ext_store_type"`
	ExtStoreId      string         `json:"extStoreId" db:"ext_store_id"`
	RotExtStoreId   sql.NullString `json:"rotExtStoreId" db:"rot_ext_store_id"`
}
