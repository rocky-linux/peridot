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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"peridot.resf.org/peridot/db/models"
	"peridot.resf.org/utils"
	"strings"
	"sync"
)

// LoadedKey keeps the key and some other information in memory
// todo(mustafa): Add TTL, rotation check, etc.
type LoadedKey struct {
	sync.Mutex
	keyUuid uuid.UUID
	gpgId   string
}

func logCmdRun(cmd *exec.Cmd) (*bytes.Buffer, error) {
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
	return &outBuf, cmd.Run()
}

func gpgCmdEnv(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = append(cmd.Env, "GNUPGHOME=/keykeeper/gpg")
	return cmd
}

func (s *Server) importGpgKey(armoredKey string) error {
	cmd := gpgCmdEnv(exec.Command("gpg", "--batch", "--yes", "--import", "-"))
	cmd.Stdin = strings.NewReader(armoredKey)
	out, err := logCmdRun(cmd)
	if err != nil {
		s.log.Errorf("failed to import gpg key: %s", out.String())
	}
	return err
}

func (s *Server) importRpmKey(publicKey string) error {
	tmpFile, err := ioutil.TempFile("/tmp", "peridot-key-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.Write([]byte(publicKey))
	if err != nil {
		return err
	}
	cmd := gpgCmdEnv(exec.Command("rpm", "--import", tmpFile.Name()))
	out, err := logCmdRun(cmd)
	if err != nil {
		s.log.Errorf("failed to import rpm key: %s", out.String())
	}
	return err
}

// WarmGPGKey warms up a specific GPG key
// This involves shelling out to GPG to import the key
func (s *Server) WarmGPGKey(key string, armoredKey string, gpgKey *crypto.Key, db *models.Key) (*LoadedKey, error) {
	cachedKeyAny, ok := s.keys.Load(key)
	// This means that the key is already loaded
	if ok {
		return cachedKeyAny.(*LoadedKey), nil
	}

	err := s.importGpgKey(armoredKey)
	if err != nil {
		return nil, err
	}

	err = s.importRpmKey(db.PublicKey)
	if err != nil {
		return nil, err
	}

	cachedKey := &LoadedKey{
		keyUuid: db.ID,
		gpgId:   gpgKey.GetHexKeyID(),
	}
	s.keys.Store(key, cachedKey)

	return cachedKey, nil
}

// EnsureGPGKey ensures that the key is loaded
func (s *Server) EnsureGPGKey(key string) (*LoadedKey, error) {
	cachedKeyAny, ok := s.keys.Load(key)
	if ok {
		return cachedKeyAny.(*LoadedKey), nil
	}

	// Key not found in cache, fetch from database
	// Fetch the encryption key, nonce and external key ID from the database
	k, err := s.db.GetKeyByName(key)
	if err != nil {
		return nil, err
	}
	encBytes, err := hex.DecodeString(k.EncKey)
	if err != nil {
		return nil, err
	}
	nonce, err := hex.DecodeString(k.Nonce)
	if err != nil {
		return nil, err
	}

	// Fetch the encrypted key from secret store
	store := s.stores[k.ExtStoreType]
	if store == nil {
		return nil, fmt.Errorf("no store found for type %s, which key with name \"%s\" relies on. manual rotation may be required #TODO-DOCS", k.ExtStoreType, k.Name)
	}
	encryptedArmoredKeyHex, err := store.Get(k.ExtStoreId)
	if err != nil {
		return nil, err
	}
	encryptedArmoredKey, err := hex.DecodeString(encryptedArmoredKeyHex)
	if err != nil {
		return nil, err
	}

	// Decrypt the key
	block, err := aes.NewCipher(encBytes)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	armoredKey, err := gcm.Open(nil, nonce, encryptedArmoredKey, nil)
	if err != nil {
		return nil, err
	}
	keyObj, err := crypto.NewKeyFromArmored(string(armoredKey))
	if err != nil {
		s.log.Errorf("could not get key from armored string: %v", err)
		return nil, utils.InternalError
	}

	loadedKey, err := s.WarmGPGKey(key, string(armoredKey), keyObj, k)
	if err != nil {
		return nil, err
	}

	return loadedKey, nil
}
