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

package db

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	secparsepb "peridot.resf.org/secparse/proto/v1"
	"peridot.resf.org/secparse/rpmutils"
)

func DTOShortCodeToPB(sc *ShortCode) *secparseadminpb.ShortCode {
	ret := &secparseadminpb.ShortCode{
		Code: sc.Code,
		Mode: secparseadminpb.ShortCodeMode(sc.Mode),
	}

	if sc.ArchivedAt.Valid {
		ret.Archived = true
	}

	return ret
}

func DTOListShortCodesToPB(scs []*ShortCode) []*secparseadminpb.ShortCode {
	var ret []*secparseadminpb.ShortCode

	for _, v := range scs {
		ret = append(ret, DTOShortCodeToPB(v))
	}

	return ret
}

func DTOAdvisoryToPB(sc *Advisory) *secparsepb.Advisory {
	var errataType string
	switch secparsepb.Advisory_Type(sc.Type) {
	case secparsepb.Advisory_Security:
		errataType = "SA"
		break
	case secparsepb.Advisory_BugFix:
		errataType = "BA"
		break
	case secparsepb.Advisory_Enhancement:
		errataType = "EA"
		break
	default:
		errataType = "UNK"
		break
	}

	var publishedAt *timestamppb.Timestamp
	if sc.PublishedAt.Valid {
		publishedAt = timestamppb.New(sc.PublishedAt.Time)
	}

	ret := &secparsepb.Advisory{
		Type:             secparsepb.Advisory_Type(sc.Type),
		ShortCode:        sc.ShortCodeCode,
		Name:             fmt.Sprintf("%s%s-%d:%d", sc.ShortCodeCode, errataType, sc.Year, sc.Num),
		Synopsis:         sc.Synopsis,
		Severity:         secparsepb.Advisory_Severity(sc.Severity),
		Topic:            sc.Topic,
		Description:      sc.Description,
		AffectedProducts: sc.AffectedProducts,
		Fixes:            sc.Fixes,
		Cves:             sc.Cves,
		References:       sc.References,
		PublishedAt:      publishedAt,
		Rpms:             sc.RPMs,
	}
	if sc.Solution.Valid {
		ret.Solution = &wrapperspb.StringValue{Value: sc.Solution.String}
	}

	for i, rpm := range sc.RPMs {
		sc.RPMs[i] = rpmutils.Epoch().ReplaceAllString(rpm, "")
	}

	return ret
}

func DTOListAdvisoriesToPB(scs []*Advisory) []*secparsepb.Advisory {
	var ret []*secparsepb.Advisory

	for _, v := range scs {
		ret = append(ret, DTOAdvisoryToPB(v))
	}

	return ret
}

func DTOCVEToPB(cve *CVE) *secparseadminpb.CVE {
	ret := &secparseadminpb.CVE{
		Name:  cve.ID,
		State: secparseadminpb.CVEState(cve.State),
	}

	if cve.SourceBy.Valid {
		ret.SourceBy = wrapperspb.String(cve.SourceBy.String)
	}
	if cve.SourceLink.Valid {
		ret.SourceLink = wrapperspb.String(cve.SourceLink.String)
	}

	return ret
}

func DTOListCVEsToPB(cves []*CVE) []*secparseadminpb.CVE {
	var ret []*secparseadminpb.CVE

	for _, v := range cves {
		ret = append(ret, DTOCVEToPB(v))
	}

	return ret
}
