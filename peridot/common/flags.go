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

package peridotcommon

import (
	"github.com/spf13/pflag"
	"os"
)

func AddFlags(pflags *pflag.FlagSet) {
	var defaultAccessKey string
	var defaultSecretKey string
	defaultRegion := "us-west-2"
	defaultBucket := "rocky-build-system"

	if os.Getenv("LOCALSTACK_ENDPOINT") != "" {
		defaultAccessKey = "AKIAIOSFODNN7EXAMPLE"
		defaultSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		defaultRegion = "us-east-2"
		defaultBucket = "peridot"
	}

	pflags.String("s3-endpoint", "", "S3 endpoint")
	pflags.String("s3-access-key", defaultAccessKey, "S3 Access Key")
	pflags.String("s3-secret-key", defaultSecretKey, "S3 Secret Key")
	pflags.String("s3-region", defaultRegion, "S3 Region")
	pflags.String("s3-bucket", defaultBucket, "S3 Bucket")

	pflags.Bool("s3-disable-ssl", false, "Disable secure S3 access")
	pflags.Bool("s3-force-path-style", false, "Disable secure S3 access")
}
