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

package apollomock

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"peridot.resf.org/utils"
	"time"
)

type Access struct {
	ShortCodes              []*apollodb.ShortCode
	Advisories              []*apollodb.Advisory
	Cves                    []*apollodb.CVE
	Fixes                   []*apollodb.Fix
	AdvisoryReferences      []*apollodb.AdvisoryReference
	Products                []*apollodb.Product
	AffectedProducts        []*apollodb.AffectedProduct
	BuildReferences         []*apollodb.BuildReference
	MirrorStates            []*apollodb.MirrorState
	AdvisoryCVEs            []*apollodb.AdvisoryCVE
	AdvisoryFixes           []*apollodb.AdvisoryFix
	IgnoredUpstreamPackages []*apollodb.IgnoredUpstreamPackage
	RebootSuggestedPackages []*apollodb.RebootSuggestedPackage
	AdvisoryRPMs            []*apollodb.AdvisoryRPM
}

func New() *Access {
	return &Access{
		ShortCodes:              []*apollodb.ShortCode{},
		Advisories:              []*apollodb.Advisory{},
		Cves:                    []*apollodb.CVE{},
		Fixes:                   []*apollodb.Fix{},
		AdvisoryReferences:      []*apollodb.AdvisoryReference{},
		Products:                []*apollodb.Product{},
		AffectedProducts:        []*apollodb.AffectedProduct{},
		BuildReferences:         []*apollodb.BuildReference{},
		MirrorStates:            []*apollodb.MirrorState{},
		AdvisoryCVEs:            []*apollodb.AdvisoryCVE{},
		AdvisoryFixes:           []*apollodb.AdvisoryFix{},
		IgnoredUpstreamPackages: []*apollodb.IgnoredUpstreamPackage{},
		RebootSuggestedPackages: []*apollodb.RebootSuggestedPackage{},
		AdvisoryRPMs:            []*apollodb.AdvisoryRPM{},
	}
}

func (a *Access) GetAllShortCodes() ([]*apollodb.ShortCode, error) {
	return a.ShortCodes, nil
}

func (a *Access) GetShortCodeByCode(code string) (*apollodb.ShortCode, error) {
	for _, val := range a.ShortCodes {
		if val.Code == code {
			return val, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateShortCode(code string, mode apollopb.ShortCode_Mode) (*apollodb.ShortCode, error) {
	now := time.Now()

	shortCode := apollodb.ShortCode{
		Code:       code,
		Mode:       int8(mode),
		CreatedAt:  &now,
		ArchivedAt: sql.NullTime{},
	}
	a.ShortCodes = append(a.ShortCodes, &shortCode)

	return &shortCode, nil
}

func (a *Access) getAdvisoriesWithJoin(filter func(*apollodb.Advisory) bool) []*apollodb.Advisory {
	var advisories []*apollodb.Advisory
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

func (a *Access) GetAllAdvisories(filters *apollopb.AdvisoryFilters, page int32, limit int32) ([]*apollodb.Advisory, error) {
	return a.getAdvisoriesWithJoin(func(advisory *apollodb.Advisory) bool {
		if filters.Product != nil {
			if !utils.StrContains(filters.Product.Value, advisory.AffectedProducts) {
				return false
			}
		}
		if advisory.PublishedAt.Valid {
			if filters.Before != nil {
				if advisory.PublishedAt.Time.After(filters.Before.AsTime()) {
					return false
				}
			}
			if filters.After != nil {
				if advisory.PublishedAt.Time.Before(filters.After.AsTime()) {
					return false
				}
			}
		}
		if filters.IncludeUnpublished != nil {
			if !filters.IncludeUnpublished.Value && !advisory.PublishedAt.Valid {
				return false
			}
		} else {
			if !advisory.PublishedAt.Valid {
				return false
			}
		}

		if advisory.Fixes != nil && len(advisory.Fixes) < 1 {
			return false
		}

		return true
	}), nil
}

func (a *Access) GetAdvisoryByCodeAndYearAndNum(code string, year int, num int) (*apollodb.Advisory, error) {
	advisories := a.getAdvisoriesWithJoin(func(advisory *apollodb.Advisory) bool {
		if advisory.ShortCodeCode == code && advisory.Year == year && advisory.Num == num {
			return true
		}

		return false
	})

	if len(advisories) == 0 {
		return nil, sql.ErrNoRows
	}

	advisory := advisories[0]

	if advisory.Fixes != nil && len(advisory.Fixes) < 1 {
		return nil, fmt.Errorf("Expected advisory fixes. Was empty.")
	}

	return advisory, nil
}

func (a *Access) CreateAdvisory(advisory *apollodb.Advisory) (*apollodb.Advisory, error) {
	var lastId int64 = 1
	if len(a.Advisories) > 0 {
		lastId = a.Advisories[len(a.Advisories)-1].ID + 1
	}

	now := time.Now()
	ret := &apollodb.Advisory{
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

func (a *Access) UpdateAdvisory(advisory *apollodb.Advisory) (*apollodb.Advisory, error) {
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

func (a *Access) GetAllUnresolvedCVEs() ([]*apollodb.CVE, error) {
	var cves []*apollodb.CVE
	var addedCVEIds []string

	for _, cve := range a.Cves {
		for _, affectedProduct := range a.AffectedProducts {
			if affectedProduct.CveID.String == cve.ID {
				switch affectedProduct.State {
				case
					int(apollopb.AffectedProduct_STATE_UNDER_INVESTIGATION_UPSTREAM),
					int(apollopb.AffectedProduct_STATE_UNDER_INVESTIGATION_DOWNSTREAM),
					int(apollopb.AffectedProduct_STATE_AFFECTED_UPSTREAM),
					int(apollopb.AffectedProduct_STATE_AFFECTED_DOWNSTREAM):
					nCve := *cve
					nCve.AffectedProductId = sql.NullInt64{Valid: true, Int64: affectedProduct.ID}
					cves = append(cves, &nCve)
					break
				}
			}
		}
	}
	for _, cve := range a.Cves {
		if !utils.StrContains(cve.ID, addedCVEIds) {
			cves = append(cves, cve)
		}
	}

	return cves, nil
}

func (a *Access) GetPendingAffectedProducts() ([]*apollodb.AffectedProduct, error) {
	var ret []*apollodb.AffectedProduct

	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.State == int(apollopb.AffectedProduct_STATE_FIXED_UPSTREAM) {
			ret = append(ret, affectedProduct)
		}
	}

	return ret, nil
}

func (a *Access) GetAllCVEsFixedDownstream() ([]*apollodb.CVE, error) {
	var cves []*apollodb.CVE

	for _, cve := range a.Cves {
		for _, affectedProduct := range a.AffectedProducts {
			if affectedProduct.CveID.String == cve.ID {
				if affectedProduct.State == int(apollopb.AffectedProduct_STATE_FIXED_DOWNSTREAM) {
					nCve := *cve
					nCve.AffectedProductId = sql.NullInt64{Valid: true, Int64: affectedProduct.ID}
					cves = append(cves, &nCve)
					break
				}
			}
		}
	}

	return cves, nil
}

func (a *Access) GetCVEByID(id string) (*apollodb.CVE, error) {
	for _, cve := range a.Cves {
		if cve.ID == id {
			return cve, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetAllCVEs() ([]*apollodb.CVE, error) {
	return a.Cves, nil
}

func (a *Access) CreateCVE(cveId string, shortCode string, sourceBy *string, sourceLink *string, content types.NullJSONText) (*apollodb.CVE, error) {
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
	cve := &apollodb.CVE{
		ID:         cveId,
		CreatedAt:  &now,
		AdvisoryId: sql.NullInt64{},
		ShortCode:  shortCode,
		SourceBy:   sby,
		SourceLink: sl,
		Content:    content,
	}
	a.Cves = append(a.Cves, cve)

	return cve, nil
}

func (a *Access) SetCVEContent(cveId string, content types.JSONText) error {
	for _, cve := range a.Cves {
		if cve.ID == cveId {
			cve.Content = types.NullJSONText{Valid: true, JSONText: content}
			return nil
		}
	}

	return sql.ErrNoRows
}

func (a *Access) GetProductsByShortCode(code string) ([]*apollodb.Product, error) {
	var products []*apollodb.Product

	for _, product := range a.Products {
		if product.ShortCode == code {
			products = append(products, product)
		}
	}

	return products, nil
}

func (a *Access) GetProductByNameAndShortCode(name string, code string) (*apollodb.Product, error) {
	for _, product := range a.Products {
		if product.Name == name && product.ShortCode == code {
			return product, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetProductByID(id int64) (*apollodb.Product, error) {
	for _, product := range a.Products {
		if product.ID == id {
			return product, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateProduct(name string, currentFullVersion string, redHatMajorVersion *int32, code string, archs []string) (*apollodb.Product, error) {
	var lastId int64 = 1
	if len(a.Products) > 0 {
		lastId = a.Products[len(a.Products)-1].ID + 1
	}

	var rhmv sql.NullInt32
	if redHatMajorVersion != nil {
		rhmv.Int32 = *redHatMajorVersion
		rhmv.Valid = true
	}

	product := &apollodb.Product{
		ID:                  lastId,
		Name:                name,
		CurrentFullVersion:  currentFullVersion,
		RedHatMajorVersion:  rhmv,
		ShortCode:           code,
		Archs:               archs,
		MirrorFromDate:      sql.NullTime{},
		RedHatProductPrefix: sql.NullString{},
	}
	a.Products = append(a.Products, product)

	return product, nil
}

func (a *Access) GetAllAffectedProductsByCVE(cve string) ([]*apollodb.AffectedProduct, error) {
	var affectedProducts []*apollodb.AffectedProduct

	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.CveID.String == cve {
			affectedProducts = append(affectedProducts, affectedProduct)
		}
	}

	return affectedProducts, nil
}

func (a *Access) GetAffectedProductByCVEAndPackage(cve string, pkg string) (*apollodb.AffectedProduct, error) {
	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.CveID.String == cve && affectedProduct.Package == pkg {
			return affectedProduct, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetAffectedProductByAdvisory(advisory string) (*apollodb.AffectedProduct, error) {
	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.Advisory.String == advisory {
			return affectedProduct, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) GetAffectedProductByID(id int64) (*apollodb.AffectedProduct, error) {
	for _, affectedProduct := range a.AffectedProducts {
		if affectedProduct.ID == id {
			return affectedProduct, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (a *Access) CreateAffectedProduct(productId int64, cveId string, state int, version string, pkg string, advisory *string) (*apollodb.AffectedProduct, error) {
	var lastId int64 = 1
	if len(a.AffectedProducts) > 0 {
		lastId = a.AffectedProducts[len(a.AffectedProducts)-1].ID + 1
	}

	var adv sql.NullString
	if advisory != nil {
		adv.String = *advisory
		adv.Valid = true
	}

	affectedProduct := &apollodb.AffectedProduct{
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

func (a *Access) CreateFix(ticket string, sourceBy string, sourceLink string, description string) (int64, error) {
	var lastId int64 = 1
	if len(a.Fixes) > 0 {
		lastId = a.Fixes[len(a.Fixes)-1].ID + 1
	}

	fix := &apollodb.Fix{
		ID:          lastId,
		Ticket:      sql.NullString{Valid: true, String: ticket},
		SourceBy:    sql.NullString{Valid: true, String: sourceBy},
		SourceLink:  sql.NullString{Valid: true, String: sourceLink},
		Description: sql.NullString{Valid: true, String: description},
	}
	a.Fixes = append(a.Fixes, fix)

	return lastId, nil
}

func (a *Access) GetMirrorState(code string) (*apollodb.MirrorState, error) {
	var lastSync *apollodb.MirrorState

	for _, mirrorState := range a.MirrorStates {
		if mirrorState.ShortCode == code {
			if mirrorState.LastSync.Valid {
				lastSync = mirrorState
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

	mirrorState := &apollodb.MirrorState{
		ShortCode: code,
		LastSync:  sql.NullTime{Valid: true, Time: *lastSync},
	}
	a.MirrorStates = append(a.MirrorStates, mirrorState)

	return nil
}

func (a *Access) UpdateMirrorStateErrata(code string, lastSync *time.Time) error {
	for _, mirrorState := range a.MirrorStates {
		if mirrorState.ShortCode == code {
			mirrorState.ErrataAfter.Time = *lastSync
			mirrorState.ErrataAfter.Valid = true

			return nil
		}
	}

	mirrorState := &apollodb.MirrorState{
		ShortCode:   code,
		ErrataAfter: sql.NullTime{Valid: true, Time: *lastSync},
	}
	a.MirrorStates = append(a.MirrorStates, mirrorState)

	return nil
}

func (a *Access) GetMaxLastSync() (*time.Time, error) {
	var maxLastSync *time.Time

	for _, mirrorState := range a.MirrorStates {
		if mirrorState.LastSync.Valid {
			if maxLastSync == nil || mirrorState.LastSync.Time.After(*maxLastSync) {
				maxLastSync = &mirrorState.LastSync.Time
			}
		}
	}

	if maxLastSync == nil {
		return nil, sql.ErrNoRows
	}

	return maxLastSync, nil
}

func (a *Access) CreateBuildReference(affectedProductId int64, rpm string, srcRpm string, cveId string, sha256Sum string, kojiId *string, peridotId *string) (*apollodb.BuildReference, error) {
	var lastId int64 = 1
	if len(a.BuildReferences) > 0 {
		lastId = a.BuildReferences[len(a.BuildReferences)-1].ID + 1
	}

	buildReference := &apollodb.BuildReference{
		ID:                lastId,
		AffectedProductId: affectedProductId,
		Rpm:               rpm,
		SrcRpm:            srcRpm,
		CveID:             cveId,
		Sha256Sum:         sha256Sum,
	}
	if kojiId != nil {
		buildReference.KojiID = sql.NullString{Valid: true, String: *kojiId}
	}
	if peridotId != nil {
		buildReference.PeridotID = sql.NullString{Valid: true, String: *peridotId}
	}

	a.BuildReferences = append(a.BuildReferences, buildReference)

	return buildReference, nil
}

func (a *Access) CreateAdvisoryReference(advisoryId int64, url string) error {
	var lastId int64 = 1
	if len(a.AdvisoryReferences) > 0 {
		lastId = a.AdvisoryReferences[len(a.AdvisoryReferences)-1].ID + 1
	}

	advisoryReference := &apollodb.AdvisoryReference{
		ID:         lastId,
		URL:        url,
		AdvisoryId: advisoryId,
	}
	a.AdvisoryReferences = append(a.AdvisoryReferences, advisoryReference)

	return nil
}

func (a *Access) GetAllIgnoredPackagesByProductID(productID int64) ([]string, error) {
	var packages []string

	for _, ignoredPackage := range a.IgnoredUpstreamPackages {
		if ignoredPackage.ProductID == productID {
			packages = append(packages, ignoredPackage.Package)
		}
	}

	return packages, nil
}

func (a *Access) GetAllRebootSuggestedPackages() ([]string, error) {
	var packages []string

	for _, p := range a.RebootSuggestedPackages {
		packages = append(packages, p.Name)
	}

	return packages, nil
}

func (a *Access) AddAdvisoryFix(advisoryId int64, fixId int64) error {
	advisoryFix := &apollodb.AdvisoryFix{
		AdvisoryID: advisoryId,
		FixID:      fixId,
	}
	a.AdvisoryFixes = append(a.AdvisoryFixes, advisoryFix)

	return nil
}

func (a *Access) AddAdvisoryCVE(advisoryId int64, cveId string) error {
	advisoryCVE := &apollodb.AdvisoryCVE{
		AdvisoryID: advisoryId,
		CveID:      cveId,
	}
	a.AdvisoryCVEs = append(a.AdvisoryCVEs, advisoryCVE)

	return nil
}

func (a *Access) AddAdvisoryRPM(advisoryId int64, name string, productID int64) error {
	advisoryRPM := &apollodb.AdvisoryRPM{
		AdvisoryID: advisoryId,
		Name:       name,
		ProductID:  productID,
	}
	a.AdvisoryRPMs = append(a.AdvisoryRPMs, advisoryRPM)

	return nil
}

func (a *Access) Begin() (utils.Tx, error) {
	return &utils.MockTx{}, nil
}

func (a *Access) UseTransaction(_ utils.Tx) apollodb.Access {
	return a
}
