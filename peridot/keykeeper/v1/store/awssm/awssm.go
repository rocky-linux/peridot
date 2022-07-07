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

package awssm

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/spf13/viper"
	"peridot.resf.org/utils"
)

type Store struct {
	sm *secretsmanager.SecretsManager
}

func prefixed(key string) string {
	return fmt.Sprintf("%s%s", viper.GetString("awssm-prefix"), key)
}

func New() (*Store, error) {
	sess, err := utils.NewAwsSession(&aws.Config{})
	if err != nil {
		return nil, err
	}

	sm := secretsmanager.New(sess)

	return &Store{
		sm: sm,
	}, nil
}

// Create creates a secret in AWS Secrets Manager
func (s *Store) Create(key string, value string) error {
	_, err := s.sm.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:         aws.String(prefixed(key)),
		SecretString: aws.String(value),
	})
	return err
}

// Set sets the secret value for the given key.
func (s *Store) Set(key string, value string) error {
	input := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(prefixed(key)),
		SecretString: aws.String(value),
	}

	_, err := s.sm.PutSecretValue(input)
	return err
}

// Get returns the secret value for the given key.
func (s *Store) Get(key string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(prefixed(key)),
	}

	result, err := s.sm.GetSecretValue(input)
	if err != nil {
		return "", err
	}

	return *result.SecretString, nil
}
