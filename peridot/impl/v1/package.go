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

package peridotimplv1

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

func (s *Server) ListPackages(ctx context.Context, req *peridotpb.ListPackagesRequest) (*peridotpb.ListPackagesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	pkgs, err := s.db.GetPackagesInProject(req.Filters, req.ProjectId, page, limit)
	if err != nil {
		s.log.Errorf("could not list packages: %v", err)
		return nil, utils.CouldNotRetrieveObjects
	}
	var total int64
	if len(pkgs) > 0 {
		total = pkgs[0].Total
	} else {
		total, err = s.db.PackageCountInProject(req.ProjectId)
		if err != nil {
			s.log.Errorf("could not count packages: %v", err)
			return nil, utils.CouldNotRetrieveObjects
		}
	}

	return &peridotpb.ListPackagesResponse{
		Packages: pkgs.ToProto(),
		Total:    total,
		Size:     limit,
		Page:     page,
	}, nil
}

func (s *Server) GetPackage(ctx context.Context, req *peridotpb.GetPackageRequest) (*peridotpb.GetPackageResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if err := s.checkPermission(ctx, ObjectProject, req.ProjectId, PermissionView); err != nil {
		return nil, err
	}

	filters := &peridotpb.PackageFilters{}
	switch req.Field {
	case "id":
		filters.Id = wrapperspb.String(req.Value)
		break
	case "name":
		filters.NameExact = wrapperspb.String(req.Value)
		break
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid field")
	}

	pkgs, err := s.db.GetPackagesInProject(filters, req.ProjectId, 0, 1)
	if err != nil {
		s.log.Errorf("could not get package: %v", err)
		return nil, status.Error(codes.Internal, "could not get package")
	}
	if len(pkgs) != 1 {
		return nil, status.Error(codes.NotFound, "package not found")
	}

	if req.Field == "name" {
		if pkgs[0].Name != req.Value {
			return nil, status.Error(codes.NotFound, "package not found")
		}
	}

	return &peridotpb.GetPackageResponse{
		Package: pkgs[0].ToProto(),
	}, nil
}
