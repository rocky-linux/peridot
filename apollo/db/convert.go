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

package apollodb

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	apollopb "peridot.resf.org/apollo/pb"
	"strings"
)

func DTOShortCodeToPB(sc *ShortCode) *apollopb.ShortCode {
	ret := &apollopb.ShortCode{
		Code: sc.Code,
	}

	if sc.ArchivedAt.Valid {
		ret.Archived = true
	}

	return ret
}

func DTOListShortCodesToPB(scs []*ShortCode) []*apollopb.ShortCode {
	var ret []*apollopb.ShortCode

	for _, v := range scs {
		ret = append(ret, DTOShortCodeToPB(v))
	}

	return ret
}

func DTOAdvisoryToPB(sc *Advisory) *apollopb.Advisory {
	var errataType string
	switch apollopb.Advisory_Type(sc.Type) {
	case apollopb.Advisory_TYPE_SECURITY:
		errataType = "SA"
		break
	case apollopb.Advisory_TYPE_BUGFIX:
		errataType = "BA"
		break
	case apollopb.Advisory_TYPE_ENHANCEMENT:
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

	ret := &apollopb.Advisory{
		Type:             apollopb.Advisory_Type(sc.Type),
		ShortCode:        sc.ShortCodeCode,
		Name:             fmt.Sprintf("%s%s-%d:%d", sc.ShortCodeCode, errataType, sc.Year, sc.Num),
		Synopsis:         sc.Synopsis,
		Severity:         apollopb.Advisory_Severity(sc.Severity),
		Topic:            sc.Topic,
		Description:      sc.Description,
		AffectedProducts: sc.AffectedProducts,
		Fixes:            nil,
		Cves:             []*apollopb.CVE{},
		References:       sc.References,
		PublishedAt:      publishedAt,
		Rpms:             nil,
		RebootSuggested:  sc.RebootSuggested,
	}
	if sc.Solution.Valid {
		ret.Solution = &wrapperspb.StringValue{Value: sc.Solution.String}
	}
	for _, cve := range sc.Cves {
		split := strings.SplitN(cve, ":::", 6)
		ret.Cves = append(ret.Cves, &apollopb.CVE{
			Name:               split[2],
			SourceBy:           wrapperspb.String(split[0]),
			SourceLink:         wrapperspb.String(split[1]),
			Cvss3ScoringVector: wrapperspb.String(split[3]),
			Cvss3BaseScore:     wrapperspb.String(split[4]),
			Cwe:                wrapperspb.String(split[5]),
		})
	}
	if len(sc.Fixes) > 0 {
		ret.Fixes = []*apollopb.Fix{}
	}
	for _, fix := range sc.Fixes {
		split := strings.SplitN(fix, ":::", 4)
		ret.Fixes = append(ret.Fixes, &apollopb.Fix{
			Ticket:      wrapperspb.String(split[0]),
			SourceBy:    wrapperspb.String(split[1]),
			SourceLink:  wrapperspb.String(split[2]),
			Description: wrapperspb.String(split[3]),
		})
	}
	if len(sc.RPMs) > 0 {
		ret.Rpms = map[string]*apollopb.RPMs{}
	}
	for _, rpm := range sc.RPMs {
		split := strings.SplitN(rpm, ":::", 2)
		nvra := split[0]
		productName := split[1]
		if ret.Rpms[productName] == nil {
			ret.Rpms[productName] = &apollopb.RPMs{}
		}

		ret.Rpms[productName].Nvras = append(ret.Rpms[productName].Nvras, nvra)
	}

	return ret
}

func DTOListAdvisoriesToPB(scs []*Advisory) []*apollopb.Advisory {
	var ret []*apollopb.Advisory

	for _, v := range scs {
		ret = append(ret, DTOAdvisoryToPB(v))
	}

	return ret
}

func DTOCVEToPB(cve *CVE) *apollopb.CVE {
	ret := &apollopb.CVE{
		Name: cve.ID,
	}

	if cve.SourceBy.Valid {
		ret.SourceBy = wrapperspb.String(cve.SourceBy.String)
	}
	if cve.SourceLink.Valid {
		ret.SourceLink = wrapperspb.String(cve.SourceLink.String)
	}

	return ret
}

func DTOListCVEsToPB(cves []*CVE) []*apollopb.CVE {
	var ret []*apollopb.CVE

	for _, v := range cves {
		ret = append(ret, DTOCVEToPB(v))
	}

	return ret
}
