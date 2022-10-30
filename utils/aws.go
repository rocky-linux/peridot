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
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
)

func NewAwsSession(cfg *aws.Config) (*session.Session, error) {
	if accessKey := viper.GetString("s3-access-key"); accessKey != "" {
		cfg.Credentials = credentials.NewStaticCredentials(accessKey, viper.GetString("s3-secret-key"), "")
	}

	if endpoint := viper.GetString("s3-endpoint"); endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}

	if region := viper.GetString("s3-region"); region != "" {
		cfg.Region = aws.String(region)
	}

	if disableSsl := viper.GetBool("s3-disable-ssl"); disableSsl {
		cfg.DisableSSL = aws.Bool(true)
	}

	if forcePathStyle := viper.GetBool("s3-force-path-style"); forcePathStyle {
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	if os.Getenv("AWS_REGION") != "" {
		cfg.Region = aws.String(os.Getenv("AWS_REGION"))
	}

	if os.Getenv("LOCALSTACK_ENDPOINT") != "" && os.Getenv("RESF_ENV") == "dev" {
		cfg.Endpoint = aws.String(os.Getenv("LOCALSTACK_ENDPOINT"))
		cfg.Credentials = credentials.NewStaticCredentials("test", "test", "")
	}

	return session.NewSession(cfg)
}
