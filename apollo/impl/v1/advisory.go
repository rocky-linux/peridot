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

package apolloimpl

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/apollo/rpmutils"
	"peridot.resf.org/utils"
	"strconv"
	"time"
)

func (s *Server) ListAdvisories(_ context.Context, req *apollopb.ListAdvisoriesRequest) (*apollopb.ListAdvisoriesResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if req.Filters != nil {
		req.Filters.IncludeUnpublished = nil
	}

	page := utils.MinPage(req.Page)
	limit := utils.MinLimit(req.Limit)
	ret, err := s.db.GetAllAdvisories(req.Filters, page, limit)
	if err != nil {
		s.log.Errorf("could not get advisories, error: %s", err)
		return nil, status.Error(codes.Internal, "failed to list advisories")
	}
	total := int64(0)
	if len(ret) > 0 {
		total = ret[0].Total
	}

	var lastUpdatedPb *timestamppb.Timestamp
	lastUpdated, err := s.db.GetMaxLastSync()
	if err != nil && err != sql.ErrNoRows {
		s.log.Errorf("could not get last sync time, error: %s", err)
		return nil, status.Error(codes.Internal, "failed to get last updated")
	}
	if lastUpdated != nil {
		lastUpdatedPb = timestamppb.New(*lastUpdated)
	}

	return &apollopb.ListAdvisoriesResponse{
		Advisories:  apollodb.DTOListAdvisoriesToPB(ret),
		Total:       total,
		Page:        page,
		Size:        limit,
		LastUpdated: lastUpdatedPb,
	}, nil
}

// ListAdvisoriesRSS returns advisories in RSS format. Only returns latest 25 published advisories
func (s *Server) ListAdvisoriesRSS(_ context.Context, req *apollopb.ListAdvisoriesRSSRequest) (*httpbody.HttpBody, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	if req.Filters == nil {
		req.Filters = &apollopb.AdvisoryFilters{}
	}
	req.Filters.IncludeUnpublished = nil

	ret, err := s.db.GetAllAdvisories(req.Filters, 0, 25)
	if err != nil {
		s.log.Errorf("could not get advisories, error: %s", err)
		return nil, status.Error(codes.Internal, "failed to list advisories")
	}
	total := int64(0)
	if len(ret) > 0 {
		total = ret[0].Total
	}

	var updated time.Time
	if total != 0 {
		updated = ret[0].PublishedAt.Time
	}

	feed := &feeds.Feed{
		Title:       "Apollo Security RSS Feed",
		Link:        &feeds.Link{Href: s.homepage},
		Description: "Security advisories issued using Apollo Errata Management",
		Author: &feeds.Author{
			Name:  "Rocky Enterprise Software Foundation, Inc.",
			Email: "releng@rockylinux.org",
		},
		Updated:   updated,
		Items:     []*feeds.Item{},
		Copyright: "(C) Rocky Enterprise Software Foundation, Inc. 2022. All rights reserved. CVE sources are copyright of their respective owners.",
	}
	if s.rssFeedTitle != "" {
		feed.Title = s.rssFeedTitle
	}
	if s.rssFeedDescription != "" {
		feed.Description = s.rssFeedDescription
	}
	for _, a := range ret {
		dtoToPB := apollodb.DTOAdvisoryToPB(a)
		item := &feeds.Item{
			Title:       fmt.Sprintf("%s: %s", dtoToPB.Name, a.Synopsis),
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/%s", s.homepage, dtoToPB.Name)},
			Description: a.Topic,
			Id:          fmt.Sprintf("%d", a.ID),
			Created:     a.PublishedAt.Time,
		}
		feed.Items = append(feed.Items, item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		s.log.Errorf("could not generate RSS feed, error: %s", err)
		return nil, status.Error(codes.Internal, "failed to generate RSS feed")
	}

	return &httpbody.HttpBody{
		ContentType: "application/rss+xml",
		Data:        []byte(rss),
	}, nil
}

// GetAdvisory returns a single advisory by name
func (s *Server) GetAdvisory(_ context.Context, req *apollopb.GetAdvisoryRequest) (*apollopb.GetAdvisoryResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, err
	}
	advisoryId := rpmutils.AdvisoryId().FindStringSubmatch(req.Id)
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

	return &apollopb.GetAdvisoryResponse{
		Advisory: apollodb.DTOAdvisoryToPB(advisory),
	}, nil
}
