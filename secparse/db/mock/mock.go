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

package mock

import (
	"database/sql"
	"fmt"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"peridot.resf.org/utils"
	"time"
)

type Access struct {
	ShortCodes              []*db.ShortCode
	Advisories              []*db.Advisory
	Cves                    []*db.CVE
	Fixes                   []*db.Fix
	AdvisoryReferences      []*db.AdvisoryReference
	Products                []*db.Product
	AffectedProducts        []*db.AffectedProduct
	BuildReferences         []*db.BuildReference
	MirrorStates            []*db.MirrorState
	AdvisoryCVEs            []*db.AdvisoryCVE
	AdvisoryFixes           []*db.AdvisoryFix
	IgnoredUpstreamPackages []*db.IgnoredUpstreamPackage
	AdvisoryRPMs            []*db.AdvisoryRPM
}

func New() *Access {
	return &Access{
		ShortCodes:              []*db.ShortCode{},
		Advisories:              []*db.Advisory{},
		Cves:                    []*db.CVE{},
		Fixes:                   []*db.Fix{},
		AdvisoryReferences:      []*db.AdvisoryReference{},
		Products:                []*db.Product{},
		AffectedProducts:        []*db.AffectedProduct{},
		BuildReferences:         []*db.BuildReference{},
		MirrorStates:            []*db.MirrorState{},
		AdvisoryCVEs:            []*db.AdvisoryCVE{},
		AdvisoryFixes:           []*db.AdvisoryFix{},
		IgnoredUpstreamPackages: []*db.IgnoredUpstreamPackage{},
		AdvisoryRPMs:            []*db.AdvisoryRPM{},
	}
}

func (a *Access) GetAllShortCodes() ([]*db.ShortCode, error) {
	return a.ShortCodes, nil
}

func (a *Access) GetShortCodeByCode(code string) (*db.ShortCode, error) {
	for _, val := range a.ShortCodes {
		if val.Code == code {
			return val, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateShortCode(code string, mode secparseadminpb.ShortCodeMode) (*db.ShortCode, error) {
	now := time.Now()

	shortCode := db.ShortCode{
		Code:                code,
		Mode:                int8(mode),
		CreatedAt:           &now,
		ArchivedAt:          sql.NullTime{},
		MirrorFromDate:      sql.NullTime{},
		RedHatProductPrefix: sql.NullString{},
	}
	a.ShortCodes = append(a.ShortCodes, &shortCode)

	return &shortCode, nil
}

func (a *Access) getAdvisoriesWithJoin(filter func(*db.Advisory) bool) []*db.Advisory {
	var advisories []*db.Advisory
	for _, val := range a.Advisories {
		if filter(val) {
			advisories = append(advisories, val)
		}
	}

	if len(advisories) == 0 {
		return advisories
	}

	for _, advisory := range advisories {
		advisory.AffectedProducts = []string{}
		advisory.Fixes = []string{}
		advisory.Cves = []string{}
		advisory.References = []string{}
		advisory.RPMs = []string{}
		advisory.BuildArtifacts = []string{}

		for _, advisoryCve := range a.AdvisoryCVEs {
			if advisoryCve.AdvisoryID != advisory.ID {
				continue
			}

			for _, buildReference := range a.BuildReferences {
				if buildReference.CveID == advisoryCve.CveID {
					advisory.BuildArtifacts = append(advisory.BuildArtifacts, fmt.Sprintf("%s:::%s", buildReference.Rpm, buildReference.SrcRpm))
				}
			}

			for _, cve := range a.Cves {
				if cve.ID == advisoryCve.CveID {
					cveString := fmt.Sprintf("%s:::%s:::%s", cve.SourceBy.String, cve.SourceLink.String, cve.ID)
					if !utils.StrContains(cveString, advisory.Cves) {
						advisory.Cves = append(advisory.Cves, cveString)
					}
				}
			}
			for _, val := range a.AffectedProducts {
				if val.CveID.String == advisoryCve.CveID {
					for _, product := range a.Products {
						if val.ProductID == product.ID {
							if !utils.StrContains(product.Name, advisory.AffectedProducts) {
								advisory.AffectedProducts = append(advisory.AffectedProducts, product.Name)
							}
						}
					}
				}
			}
		}

		for _, advisoryFix := range a.AdvisoryFixes {
			if advisoryFix.AdvisoryID != advisory.ID {
				continue
			}

			for _, fix := range a.Fixes {
				if fix.ID == advisoryFix.FixID {
					if !utils.StrContains(fix.Ticket.String, advisory.Fixes) {
						advisory.Fixes = append(advisory.Fixes, fix.Ticket.String)
					}
				}
			}
		}

		for _, advisoryReference := range a.AdvisoryReferences {
			if advisoryReference.AdvisoryId != advisory.ID {
				continue
			}

			if !utils.StrContains(advisoryReference.URL, advisory.References) {
				advisory.References = append(advisory.References, advisoryReference.URL)
			}
		}

		for _, advisoryRPM := range a.AdvisoryRPMs {
			if advisoryRPM.AdvisoryID != advisory.ID {
				continue
			}

			if !utils.StrContains(advisoryRPM.Name, advisory.RPMs) {
				advisory.RPMs = append(advisory.RPMs, advisoryRPM.Name)
			}
		}
	}

	return advisories
}

func (a *Access) GetAllAdvisories(publishedOnly bool) ([]*db.Advisory, error) {
	return a.getAdvisoriesWithJoin(func(advisory *db.Advisory) bool {
		if publishedOnly {
			if !advisory.PublishedAt.Valid {
				return false
			}
		}

		return true
	}), nil
}

func (a *Access) GetAdvisoryByCodeAndYearAndNum(code string, year int, num int) (*db.Advisory, error) {
	advisories := a.getAdvisoriesWithJoin(func(advisory *db.Advisory) bool {
		if advisory.ShortCodeCode == code && advisory.Year == year && advisory.Num == num {
			return true
		}

		return false
	})
	if len(advisories) == 0 {
		return nil, sql.ErrNoRows
	}

	return advisories[0], nil
}

func (a *Access) CreateAdvisory(advisory *db.Advisory) (*db.Advisory, error) {
	var lastId int64 = 1
	if len(a.Advisories) > 0 {
		lastId = a.Advisories[len(a.Advisories)-1].ID + 1
	}

	now := time.Now()
	ret := &db.Advisory{
		ID:             lastId,
		CreatedAt:      &now,
		Year:           advisory.Year,
		Num:            advisory.Num,
		Synopsis:       advisory.Synopsis,
		Topic:          advisory.Topic,
		Severity:       advisory.Severity,
		Type:           advisory.Type,
		Description:    advisory.Description,
		Solution:       advisory.Solution,
		RedHatIssuedAt: advisory.RedHatIssuedAt,
		ShortCodeCode:  advisory.ShortCodeCode,
		PublishedAt:    advisory.PublishedAt,
	}

	return ret, nil
}

func (a *Access) UpdateAdvisory(advisory *db.Advisory) (*db.Advisory, error) {
	for _, val := range a.Advisories {
		if val.ID == advisory.ID {
			val.Year = advisory.Year
			val.Num = advisory.Num
			val.Synopsis = advisory.Synopsis
			val.Topic = advisory.Topic
			val.Severity = advisory.Severity
			val.Type = advisory.Type
			val.Description = advisory.Description
			val.Solution = advisory.Solution
			val.ShortCodeCode = advisory.ShortCodeCode
			val.PublishedAt = advisory.PublishedAt

			return val, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetAllUnresolvedCVEs() ([]*db.CVE, error) {
	var cves []*db.CVE

	for _, cve := range a.Cves {
		switch cve.State {
		case
			int(secparseadminpb.CVEState_NewFromUpstream),
			int(secparseadminpb.CVEState_NewOriginal):
			cves = append(cves, cve)
			break
		}
	}

	return cves, nil
}

func (a *Access) GetAllCVEsWithAllProductsFixed() ([]*db.CVE, error) {
	var cves []*db.CVE
	var fixedAffectedProducts []*db.AffectedProduct

	for _, affectedProduct := range a.AffectedProducts {
		switch affectedProduct.State {
		case
			int(secparseadminpb.AffectedProductState_FixedUpstream),
			int(secparseadminpb.AffectedProductState_WillNotFixUpstream),
			int(secparseadminpb.AffectedProductState_WillNotFixDownstream),
			int(secparseadminpb.AffectedProductState_OutOfSupportScope):
			fixedAffectedProducts = append(fixedAffectedProducts, affectedProduct)
		}
	}

	for _, cve := range a.Cves {
		switch cve.State {
		case
			int(secparseadminpb.CVEState_NewFromUpstream),
			int(secparseadminpb.CVEState_NewOriginal),
			int(secparseadminpb.CVEState_ResolvedUpstream),
			int(secparseadminpb.CVEState_ResolvedDownstream):
			for _, fixed := range fixedAffectedProducts {
				if fixed.CveID.String == cve.ID {
					cves = append(cves, cve)
					break
				}
			}
		}
	}

	return cves, nil
}

func (a *Access) GetAllCVEsFixedDownstream() ([]*db.CVE, error) {
	var cves []*db.CVE

	for _, cve := range a.Cves {
		if cve.State == int(secparseadminpb.CVEState_ResolvedDownstream) {
			cves = append(cves, cve)
		}
	}

	return cves, nil
}

func (a *Access) GetCVEByID(id string) (*db.CVE, error) {
	for _, cve := range a.Cves {
		if cve.ID == id {
			return cve, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateCVE(cveId string, state secparseadminpb.CVEState, shortCode string, sourceBy *string, sourceLink *string) (*db.CVE, error) {
	var sby sql.NullString
	var sl sql.NullString

	if sourceBy != nil {
		sby.String = *sourceBy
		sby.Valid = true
	}

	if sourceLink != nil {
		sl.String = *sourceLink
		sl.Valid = true
	}

	now := time.Now()
	cve := &db.CVE{
		ID:         cveId,
		CreatedAt:  &now,
		State:      int(state),
		AdvisoryId: sql.NullInt64{},
		ShortCode:  shortCode,
		SourceBy:   sby,
		SourceLink: sl,
	}
	a.Cves = append(a.Cves, cve)

	return cve, nil
}

func (a *Access) UpdateCVEState(cve string, state secparseadminpb.CVEState) error {
	for _, c := range a.Cves {
		if c.ID == cve {
			c.State = int(state)
		}
	}

	return nil
}

func (a *Access) GetProductsByShortCode(code string) ([]*db.Product, error) {
	var products []*db.Product

	for _, product := range a.Products {
		if product.ShortCode == code {
			products = append(products, product)
		}
	}

	return products, nil
}

func (a *Access) GetProductByNameAndShortCode(name string, code string) (*db.Product, error) {
	for _, product := range a.Products {
		if product.Name == name && product.ShortCode == code {
			return product, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetProductByID(id int64) (*db.Product, error) {
	for _, product := range a.Products {
		if product.ID == id {
			return product, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateProduct(name string, currentFullVersion string, redHatMajorVersion *int32, code string, archs []string) (*db.Product, error) {
	var lastId int64 = 1
	if len(a.Products) > 0 {
		lastId = a.Products[len(a.Products)-1].ID + 1
	}

	var rhmv sql.NullInt32
	if redHatMajorVersion != nil {
		rhmv.Int32 = *redHatMajorVersion
		rhmv.Valid = true
	}

	product := &db.Product{
		ID:                 lastId,
		Name:               name,
		CurrentFullVersion: currentFullVersion,
		RedHatMajorVersion: rhmv,
		ShortCode:          code,
		Archs:              archs,
	}
	a.Products = append(a.Products, product)

	return product, nil
}

func (a *Access) GetAllAffectedProductsByCVE(cve string) ([]*db.AffectedProduct, error) {
	var affectedProducts []*db.AffectedProduct

	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.CveID.String == cve {
			affectedProducts = append(affectedProducts, affectedProduct)
		}
	}

	return affectedProducts, nil
}

func (a *Access) GetAffectedProductByCVEAndPackage(cve string, pkg string) (*db.AffectedProduct, error) {
	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.CveID.String == cve && affectedProduct.Package == pkg {
			return affectedProduct, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetAffectedProductByAdvisory(advisory string) (*db.AffectedProduct, error) {
	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.Advisory.String == advisory {
			return affectedProduct, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateAffectedProduct(productId int64, cveId string, state int, version string, pkg string, advisory *string) (*db.AffectedProduct, error) {
	var lastId int64 = 1
	if len(a.AffectedProducts) > 0 {
		lastId = a.AffectedProducts[len(a.AffectedProducts)-1].ID + 1
	}

	var adv sql.NullString
	if advisory != nil {
		adv.String = *advisory
		adv.Valid = true
	}

	affectedProduct := &db.AffectedProduct{
		ID:        lastId,
		ProductID: productId,
		CveID:     sql.NullString{Valid: true, String: cveId},
		State:     state,
		Version:   version,
		Package:   pkg,
		Advisory:  adv,
	}
	a.AffectedProducts = append(a.AffectedProducts, affectedProduct)

	return affectedProduct, nil
}

func (a *Access) UpdateAffectedProductStateAndPackageAndAdvisory(id int64, state int, pkg string, advisory *string) error {
	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.ID == id {
			affectedProduct.State = state
			affectedProduct.Package = pkg

			var adv sql.NullString
			if advisory != nil {
				adv.String = *advisory
				adv.Valid = true
			}
			affectedProduct.Advisory = adv

			return nil
		}
	}

	return sql.ErrNoRows
}

func (a *Access) DeleteAffectedProduct(id int64) error {
	var index *int
	for i, affectedProduct := range a.AffectedProducts {
		if affectedProduct.ID == id {
			index = &i
		}
	}
	if index == nil {
		return sql.ErrNoRows
	}

	a.AffectedProducts = append(a.AffectedProducts[:*index], a.AffectedProducts[*index+1:]...)

	return nil
}

func (a *Access) CreateFix(ticket string, description string) (int64, error) {
	var lastId int64 = 1
	if len(a.Fixes) > 0 {
		lastId = a.Fixes[len(a.Fixes)-1].ID + 1
	}

	fix := &db.Fix{
		ID:          lastId,
		Ticket:      sql.NullString{Valid: true, String: ticket},
		Description: sql.NullString{Valid: true, String: description},
	}
	a.Fixes = append(a.Fixes, fix)

	return lastId, nil
}

func (a *Access) GetMirrorStateLastSync(code string) (*time.Time, error) {
	var lastSync *time.Time

	for _, mirrorState := range a.MirrorStates {
		if mirrorState.ShortCode == code {
			if mirrorState.LastSync.Valid {
				lastSync = &mirrorState.LastSync.Time
			}
		}
	}

	if lastSync == nil {
		return nil, sql.ErrNoRows
	}

	return lastSync, nil
}

func (a *Access) UpdateMirrorState(code string, lastSync *time.Time) error {
	for _, mirrorState := range a.MirrorStates {
		if mirrorState.ShortCode == code {
			mirrorState.LastSync.Time = *lastSync
			mirrorState.LastSync.Valid = true

			return nil
		}
	}

	mirrorState := &db.MirrorState{
		ShortCode: code,
		LastSync:  sql.NullTime{Valid: true, Time: *lastSync},
	}
	a.MirrorStates = append(a.MirrorStates, mirrorState)

	return nil
}

func (a *Access) CreateBuildReference(affectedProductId int64, rpm string, srcRpm string, cveId string, kojiId string) (*db.BuildReference, error) {
	var lastId int64 = 1
	if len(a.BuildReferences) > 0 {
		lastId = a.BuildReferences[len(a.BuildReferences)-1].ID + 1
	}

	buildReference := &db.BuildReference{
		ID:                lastId,
		AffectedProductId: affectedProductId,
		Rpm:               rpm,
		SrcRpm:            srcRpm,
		CveID:             cveId,
		KojiID:            kojiId,
	}

	a.BuildReferences = append(a.BuildReferences, buildReference)

	return buildReference, nil
}

func (a *Access) CreateAdvisoryReference(advisoryId int64, url string) error {
	var lastId int64 = 1
	if len(a.AdvisoryReferences) > 0 {
		lastId = a.AdvisoryReferences[len(a.AdvisoryReferences)-1].ID + 1
	}

	advisoryReference := &db.AdvisoryReference{
		ID:         lastId,
		URL:        url,
		AdvisoryId: advisoryId,
	}
	a.AdvisoryReferences = append(a.AdvisoryReferences, advisoryReference)

	return nil
}

func (a *Access) GetAllIgnoredPackagesByShortCode(code string) ([]string, error) {
	var packages []string

	for _, ignoredPackage := range a.IgnoredUpstreamPackages {
		if ignoredPackage.ShortCode == code {
			packages = append(packages, ignoredPackage.Package)
		}
	}

	return packages, nil
}

func (a *Access) AddAdvisoryFix(advisoryId int64, fixId int64) error {
	advisoryFix := &db.AdvisoryFix{
		AdvisoryID: advisoryId,
		FixID:      fixId,
	}
	a.AdvisoryFixes = append(a.AdvisoryFixes, advisoryFix)

	return nil
}

func (a *Access) AddAdvisoryCVE(advisoryId int64, cveId string) error {
	advisoryCVE := &db.AdvisoryCVE{
		AdvisoryID: advisoryId,
		CveID:      cveId,
	}
	a.AdvisoryCVEs = append(a.AdvisoryCVEs, advisoryCVE)

	return nil
}

func (a *Access) AddAdvisoryRPM(advisoryId int64, name string) error {
	advisoryRPM := &db.AdvisoryRPM{
		AdvisoryID: advisoryId,
		Name:       name,
	}
	a.AdvisoryRPMs = append(a.AdvisoryRPMs, advisoryRPM)

	return nil
}

func (a *Access) Begin() (utils.Tx, error) {
	return &utils.MockTx{}, nil
}

func (a *Access) UseTransaction(_ utils.Tx) db.Access {
	return a
}
