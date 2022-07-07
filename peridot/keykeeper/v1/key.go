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

package keykeeperv1

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	keykeeperpb "peridot.resf.org/peridot/keykeeper/pb"
	"peridot.resf.org/utils"
)

// GenerateKey generates a new key pair.
// We're trying to do this as securely as possible and while keeping it simple.
// We're using the gopenpgp library to do this (and ProtonMail is also using it and maintaining it).
// Best practices are followed. All keys are generated with the RSA algorithm with a key size of 4096 bits.
// The private key is encrypted with a random passphrase.
// If the project doesn't have a default key, the generated key is set as the default.
func (s *Server) GenerateKey(_ context.Context, req *keykeeperpb.GenerateKeyRequest) (*keykeeperpb.GenerateKeyResponse, error) {
	_, err := s.db.GetKeyByName(req.Name)
	if err == nil {
		return nil, status.Error(codes.InvalidArgument, "key with that name already exists")
	}

	encBytes := make([]byte, 32) //generate a random 32 byte key for AES
	if _, err := rand.Read(encBytes); err != nil {
		s.log.Errorf("failed to generate random key: %s", err)
		return nil, status.Error(codes.Internal, "failed to generate random key")
	}

	block, err := aes.NewCipher(encBytes)
	if err != nil {
		s.log.Errorf("failed to generate new key: %v", err)
		return nil, status.Error(codes.Internal, "failed to generate new key")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		s.log.Errorf("failed to create new GCM: %v", err)
		return nil, status.Error(codes.Internal, "failed to create new GCM")
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		s.log.Errorf("failed to generate nonce: %v", err)
		return nil, status.Error(codes.Internal, "failed to generate nonce")
	}

	keyUuid := uuid.New()

	rsaKey, err := helper.GenerateKey(req.Name, req.Email, []byte(keyUuid.String()), "rsa", 4096)
	if err != nil {
		s.log.Errorf("could not generate key: %v", err)
		return nil, status.Error(codes.Internal, "could not generate key")
	}
	keyObj, err := crypto.NewKeyFromArmored(rsaKey)
	if err != nil {
		s.log.Errorf("could not get key from armored string: %v", err)
		return nil, utils.InternalError
	}
	keyObj.GetEntity().Subkeys = []openpgp.Subkey{}
	fingerprint := keyObj.GetFingerprint()
	publicKey, err := keyObj.GetArmoredPublicKeyWithCustomHeaders("Keykeeper", "resf.keykeeper.v1")
	if err != nil {
		s.log.Errorf("could not get armored public key: %v", err)
		return nil, utils.InternalError
	}
	rsaKey, err = keyObj.Armor()
	if err != nil {
		s.log.Errorf("could not get armored key: %v", err)
		return nil, utils.InternalError
	}

	cipherText := gcm.Seal(nil, nonce, []byte(rsaKey), nil)
	cipherHex := hex.EncodeToString(cipherText)

	store := s.stores[s.defaultStore]
	err = store.Create(keyUuid.String(), cipherHex)
	if err != nil {
		s.log.Errorf("could not store key: %v", err)
		return nil, status.Error(codes.Internal, "could not store key")
	}

	beginTx, err := s.db.Begin()
	if err != nil {
		s.log.Errorf("could not start transaction: %v", err)
		return nil, utils.InternalError
	}
	tx := s.db.UseTransaction(beginTx)

	setDefault := false
	_, err = tx.GetDefaultKeyForProject(req.ProjectId)
	if err != nil {
		if err == sql.ErrNoRows {
			setDefault = true
		} else {
			s.log.Errorf("could not get default key for project: %v", err)
			return nil, utils.InternalError
		}
	}

	k, err := tx.CreateKey(keyUuid.String(), req.Name, req.Email, keyObj.GetHexKeyID(), hex.EncodeToString(encBytes), hex.EncodeToString(nonce), publicKey, s.defaultStore, keyUuid.String())
	if err != nil {
		s.log.Errorf("could not save key: %v", err)
		return nil, status.Error(codes.Internal, "could not save key")
	}
	err = tx.AttachKeyToProject(req.ProjectId, keyUuid.String(), setDefault)
	if err != nil {
		s.log.Errorf("could not attach key to project: %v", err)
		return nil, utils.InternalError
	}
	err = beginTx.Commit()
	if err != nil {
		s.log.Errorf("could not commit transaction: %v", err)
		return nil, utils.InternalError
	}

	// Insert into cache
	_, err = s.WarmGPGKey(req.Name, rsaKey, keyObj, k)
	if err != nil {
		// We don't have to fail, we can just log the error
		// and a future request will warm the key
		s.log.Errorf("could not warm key: %v", err)
	}

	return &keykeeperpb.GenerateKeyResponse{
		Name:        req.Name,
		Email:       req.Email,
		Fingerprint: fingerprint,
	}, nil
}

func (s *Server) GetPublicKey(_ context.Context, req *keykeeperpb.GetPublicKeyRequest) (*keykeeperpb.GetPublicKeyResponse, error) {
	key, err := s.db.GetKeyByName(req.KeyName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "key not found")
		}
		s.log.Errorf("could not get key: %v", err)
		return nil, status.Error(codes.Internal, "could not get key")
	}

	return &keykeeperpb.GetPublicKeyResponse{
		PublicKey: key.PublicKey,
	}, nil
}
