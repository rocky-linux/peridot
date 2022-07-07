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
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"peridot.resf.org/utils"
)

func (s *Server) ListShortCodes(_ context.Context, _ *secparseadminpb.ListShortCodesRequest) (*secparseadminpb.ListShortCodesResponse, error) {
	shortCodes, err := s.db.GetAllShortCodes()
	if err != nil {
		return nil, utils.CouldNotRetrieveObjects
	}

	return &secparseadminpb.ListShortCodesResponse{
		ShortCodes: db.DTOListShortCodesToPB(shortCodes),
	}, nil
}

func (s *Server) CreateShortCode(_ context.Context, req *secparseadminpb.CreateShortCodeRequest) (*secparseadminpb.CreateShortCodeResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	shortCode, err := s.db.CreateShortCode(req.Code, req.Mode)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, utils.ObjectAlreadyExists
		}
		return nil, utils.CouldNotCreateObject
	}

	return &secparseadminpb.CreateShortCodeResponse{
		ShortCode: db.DTOShortCodeToPB(shortCode),
	}, nil
}
