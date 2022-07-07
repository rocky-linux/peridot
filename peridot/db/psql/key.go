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

package serverpsql

import (
	"github.com/google/uuid"
	"peridot.resf.org/peridot/db/models"
)

func (a *Access) CreateKey(id string, name string, email string, gpgId string, encKey string, nonce string, publicKey string, extStoreType string, extStoreId string) (*models.Key, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	p := models.Key{
		ID:           uid,
		Name:         name,
		Email:        email,
		GpgId:        gpgId,
		EncKey:       encKey,
		Nonce:        nonce,
		ExtStoreType: extStoreType,
		ExtStoreId:   extStoreId,
	}

	err = a.query.Get(&p, "insert into gpg_keys (id, name, email, gpg_id, enc_key, nonce, public_key, ext_store_type, ext_store_id) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id, created_at", id, name, email, gpgId, encKey, nonce, publicKey, extStoreType, extStoreId)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) AttachKeyToProject(projectId string, keyId string, defaultKey bool) error {
	_, err := a.query.Exec("insert into project_gpg_keys (project_id, gpg_key_id, default_key) values ($1, $2, $3)", projectId, keyId, defaultKey)
	return err
}

func (a *Access) GetKeyByProjectIdAndId(projectId string, keyId string) (*models.Key, error) {
	var p models.Key
	err := a.query.Get(
		&p,
		`
		select
			gk.id,
			gk.created_at,
			gk.updated_at,
			gk.breached_at,
			gk.name,
			gk.email,
			gk.gpg_id,
			gk.enc_key,
			gk.rot_enc_key,
			gk.nonce,
			gk.rot_nonce,
			gk.public_key,
			gk.ext_store_type,
			gk.rot_ext_store_type,
			gk.ext_store_id,
			gk.rot_ext_store_id
		from gpg_keys gk
		inner join project_gpg_keys pgk on pgk.gpg_key_id = gk.id
		where
			gk.id = $1
			and pgk.project_id = $2
		`,
		keyId,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// GetDefaultKeyForProject returns the default gpg key for the project or an error if none is found
func (a *Access) GetDefaultKeyForProject(projectId string) (*models.Key, error) {
	var p models.Key
	err := a.query.Get(
		&p,
		`
		select
			gk.id,
			gk.created_at,
			gk.updated_at,
			gk.breached_at,
			gk.name,
			gk.email,
			gk.gpg_id,
			gk.enc_key,
			gk.rot_enc_key,
			gk.nonce,
			gk.rot_nonce,
			gk.public_key,
			gk.ext_store_type,
			gk.rot_ext_store_type,
			gk.ext_store_id,
			gk.rot_ext_store_id
		from gpg_keys gk
		inner join project_gpg_keys pgk on pgk.gpg_key_id = gk.id
		where
			pgk.project_id = $1
			and pgk.default_key = true
		`,
		projectId,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) GetKeyByName(name string) (*models.Key, error) {
	var p models.Key
	err := a.query.Get(
		&p,
		`
		select
			gk.id,
			gk.created_at,
			gk.updated_at,
			gk.breached_at,
			gk.name,
			gk.email,
			gk.gpg_id,
			gk.enc_key,
			gk.rot_enc_key,
			gk.nonce,
			gk.rot_nonce,
			gk.public_key,
			gk.ext_store_type,
			gk.rot_ext_store_type,
			gk.ext_store_id,
			gk.rot_ext_store_id
		from gpg_keys gk
		where
			gk.name = $1
		`,
		name,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
