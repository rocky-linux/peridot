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

package yumrepofsv1

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"regexp"
	"strings"
	"time"
)

var (
	ErrCouldNotFindRevision = status.Error(codes.NotFound, "revision not found")
	ErrInvalidBlob          = status.Error(codes.InvalidArgument, "invalid blob given")
	RegexBlob               = regexp.MustCompile(`^(.+)-(.+)(\.xml|\.xml\.gz|\.yaml|\.yaml\.gz)$`)
)

func (s *Server) GetBlob(ctx context.Context, req *yumrepofspb.GetBlobRequest) (*httpbody.HttpBody, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if req.Arch == "i386" {
		req.Arch = "i686"
	}

	if strings.HasSuffix(req.Blob, ".sqlite.gz") {
		s3Req, _ := s.s3.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(viper.GetString("s3-bucket")),
			Key:    utils.StringP(fmt.Sprintf("sqlite-files/%s", req.Blob)),
		})
		urlStr, err := s3Req.Presign(24 * time.Hour)
		if err != nil {
			s.log.Errorf("failed to presign s3 request: %v", err)
			return nil, status.Error(codes.Internal, "failed to presign s3 request")
		}

		header := metadata.Pairs("Location", urlStr)
		err = grpc.SetHeader(ctx, header)
		if err != nil {
			s.log.Errorf("failed to send header: %v", err)
			return nil, status.Error(codes.Internal, "failed to send header")
		}
		return &httpbody.HttpBody{
			ContentType: "text/plain",
			Data:        []byte(fmt.Sprintf("Redirecting to %s", urlStr)),
		}, nil
	}

	if !RegexBlob.MatchString(req.Blob) {
		return nil, ErrInvalidBlob
	}

	blob := RegexBlob.FindStringSubmatch(req.Blob)

	revision, err := s.db.GetRepositoryRevision(blob[1])
	if err != nil {
		return nil, ErrCouldNotFindRevision
	}

	var dataB64 string
	switch blob[2] {
	case "PRIMARY":
		dataB64 = revision.PrimaryXml
	case "FILELISTS":
		dataB64 = revision.FilelistsXml
	case "OTHER":
		dataB64 = revision.OtherXml
	case "GROUPS":
		dataB64 = revision.GroupsXml
	case "MODULES":
		dataB64 = revision.ModulesYaml
	default:
		return nil, ErrInvalidBlob
	}

	contentType := "application/xml+gzip"
	data, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return nil, err
	}

	if blob[3] == ".yaml.gz" {
		contentType = "text/yaml+gzip"
	}

	if blob[3] == ".xml" || blob[3] == ".yaml" {
		contentType = "application/xml"
		if blob[3] == ".yaml" {
			contentType = "text/yaml"
		}

		var buf bytes.Buffer
		buf.Write(data)
		r, err := gzip.NewReader(&buf)
		if err != nil {
			return nil, utils.InternalError
		}

		data, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, utils.InternalError
		}
	}

	return &httpbody.HttpBody{
		ContentType: contentType,
		Data:        data,
	}, nil
}
