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

package obsidianpsql

import (
	"peridot.resf.org/obsidian/db/models"
	"peridot.resf.org/utils"
)

func (a *Access) CreateUser(name *string, email string) (*models.User, error) {
	p := models.User{
		Name:  utils.StringPointerToNullString(name),
		Email: email,
	}

	err := a.query.Get(&p, "insert into users (name, email) values ($1, $2) returning id, created_at, updated_at", name, email)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) GetUserByEmail(email string) (*models.User, error) {
	p := models.User{}

	err := a.query.Get(&p, "select id, created_at, updated_at, name, email, unlock_token, locked_at from users where email = $1", email)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) GetUserByID(id string) (*models.User, error) {
	p := models.User{}

	err := a.query.Get(&p, "select id, created_at, updated_at, name, email, unlock_token, locked_at from users where id = $1", id)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) GetUserByOAuth2ProviderExternalID(providerID string, externalID string) (*models.User, error) {
	p := models.User{}

	err := a.query.Get(
		&p,
		`
		select
			u.id,
			u.created_at,
			u.updated_at,
			u.name,
			u.email,
			u.unlock_token,
			u.locked_at
		from users u
		inner join user_oauth2_connections uoc on uoc.user_id = u.id
		where
			uoc.oauth2_provider_id = $1
			and uoc.external_id = $2
		`,
		providerID,
		externalID,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (a *Access) LinkUserToOAuth2Provider(userID string, providerID string, externalID string) error {
	_, err := a.query.Exec(
		"insert into user_oauth2_connections (user_id, oauth2_provider_id, external_id) values ($1, $2, $3) on conflict do nothing",
		userID,
		providerID,
		externalID,
	)
	return err
}
