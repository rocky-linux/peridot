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

package impl

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"peridot.resf.org/secparse/db"
	secparsepb "peridot.resf.org/secparse/proto/v1"
	"peridot.resf.org/utils"
	"strconv"
)

func (s *Server) ListAdvisories(_ context.Context, _ *secparsepb.ListAdvisoriesRequest) (*secparsepb.ListAdvisoriesResponse, error) {
	advisories, err := s.db.GetAllAdvisories(true)
	if err != nil {
		logrus.Error(err)
		return nil, utils.CouldNotRetrieveObjects
	}

	return &secparsepb.ListAdvisoriesResponse{
		Advisories: db.DTOListAdvisoriesToPB(advisories),
	}, nil
}

func (s *Server) GetAdvisory(_ context.Context, req *secparsepb.GetAdvisoryRequest) (*secparsepb.GetAdvisoryResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	advisoryId := s.advisoryIdRegex.FindStringSubmatch(req.Id)
	code := advisoryId[1]
	year, err := strconv.Atoi(advisoryId[3])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid year")
	}
	num, err := strconv.Atoi(advisoryId[4])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid num")
	}

	advisory, err := s.db.GetAdvisoryByCodeAndYearAndNum(code, year, num)
	if err != nil {
		logrus.Error(err)
	}
	if err != nil || !advisory.PublishedAt.Valid {
		return nil, utils.CouldNotFindObject
	}

	return &secparsepb.GetAdvisoryResponse{
		Advisory: db.DTOAdvisoryToPB(advisory),
	}, nil
}
