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
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
	"strings"
	"time"
)

func (s *Server) GetRpm(ctx context.Context, req *yumrepofspb.GetRpmRequest) (*yumrepofspb.GetRpmResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if req.Arch == "i386" {
		req.Arch = "i686"
	}

	fileName := fmt.Sprintf("%s/%s.rpm", req.ParentTaskId, strings.TrimSuffix(req.FileName, ".rpm"))
	if len(req.ParentTaskId) == 1 {
		latestRevision, err := s.db.GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(req.ProjectId, req.RepoName, req.Arch)
		if err != nil {
			return nil, utils.CouldNotFindObject
		}
		var urlMappings map[string]string
		err = json.Unmarshal(latestRevision.UrlMappings, &urlMappings)
		if err != nil {
			return nil, err
		}
		fileName = urlMappings[fileName]
	}
	s3Req, _ := s.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(viper.GetString("s3-bucket")),
		Key:    &fileName,
	})
	urlStr, err := s3Req.Presign(24 * time.Hour)
	if err != nil {
		s.log.Errorf("failed to presign s3 request: %v", err)
		return nil, status.Error(codes.Internal, "failed to presign s3 request")
	}

	header := metadata.Pairs("Location", urlStr)
	err = grpc.SendHeader(ctx, header)
	if err != nil {
		s.log.Errorf("failed to send header: %v", err)
		return nil, status.Error(codes.Internal, "failed to send header")
	}

	return &yumrepofspb.GetRpmResponse{
		RedirectUrl: urlStr,
	}, nil
}
