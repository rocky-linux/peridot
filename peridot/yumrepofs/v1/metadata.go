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
	"encoding/base64"
	"encoding/json"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"path/filepath"
	yumrepofspb "peridot.resf.org/peridot/yumrepofs/pb"
	"peridot.resf.org/utils"
)

func (s *Server) GetRepoMd(_ context.Context, req *yumrepofspb.GetRepoMdRequest) (*httpbody.HttpBody, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	latestRevision, err := s.db.GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(req.ProjectId, req.RepoName, req.Arch)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}

	repomd, err := base64.StdEncoding.DecodeString(latestRevision.RepomdXml)
	if err != nil {
		return nil, utils.InternalError
	}

	return &httpbody.HttpBody{
		ContentType: "application/xml",
		Data:        repomd,
		Extensions:  nil,
	}, nil
}

func (s *Server) GetRepoMdSignature(_ context.Context, req *yumrepofspb.GetRepoMdRequest) (*httpbody.HttpBody, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	latestRevision, err := s.db.GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(req.ProjectId, req.RepoName, req.Arch)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}

	asc, err := s.storage.ReadObject(filepath.Join("repo-signatures", latestRevision.ID.String()+".xml.asc"))
	if err != nil {
		return nil, utils.CouldNotFindObject
	}

	return &httpbody.HttpBody{
		ContentType: "application/pgp-signature",
		Data:        asc,
		Extensions:  nil,
	}, nil
}

func (s *Server) GetPublicKey(_ context.Context, req *yumrepofspb.GetPublicKeyRequest) (*httpbody.HttpBody, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	key, err := s.db.GetDefaultKeyForProject(req.ProjectId)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}

	return &httpbody.HttpBody{
		ContentType: "application/pgp-keys",
		Data:        []byte(key.PublicKey),
	}, nil
}

func (s *Server) GetUrlMappings(_ context.Context, req *yumrepofspb.GetUrlMappingsRequest) (*yumrepofspb.GetUrlMappingsResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}

	latestRevision, err := s.db.GetLatestActiveRepositoryRevisionByProjectIdAndNameAndArch(req.ProjectId, req.RepoName, req.Arch)
	if err != nil {
		return nil, utils.CouldNotFindObject
	}
	var urlMappings map[string]string
	err = json.Unmarshal(latestRevision.UrlMappings, &urlMappings)
	if err != nil {
		return nil, err
	}

	return &yumrepofspb.GetUrlMappingsResponse{
		UrlMappings: urlMappings,
	}, nil
}
