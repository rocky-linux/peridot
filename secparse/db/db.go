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
	"database/sql"
	"github.com/lib/pq"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/utils"
	"time"
)

// ShortCode is the DTO struct for `resf.secparse.admin.ShortCode`
type ShortCode struct {
	Code                string         `db:"code"`
	Mode                int8           `db:"mode"`
	CreatedAt           *time.Time     `db:"created_at"`
	ArchivedAt          sql.NullTime   `db:"archived_at"`
	MirrorFromDate      sql.NullTime   `db:"mirror_from_date"`
	RedHatProductPrefix sql.NullString `db:"redhat_product_prefix"`
}

// Advisory is the DTO struct for `resf.secparse.Advisory`
type Advisory struct {
	ID        int64      `db:"id"`
	CreatedAt *time.Time `db:"created_at"`

	Year int `db:"year"`
	Num  int `db:"num"`

	Synopsis    string         `db:"synopsis"`
	Topic       string         `db:"topic"`
	Severity    int            `db:"severity"`
	Type        int            `db:"type"`
	Description string         `db:"description"`
	Solution    sql.NullString `db:"solution"`

	RedHatIssuedAt sql.NullTime `db:"redhat_issued_at"`
	ShortCodeCode  string       `db:"short_code_code"`
	PublishedAt    sql.NullTime `db:"published_at"`

	AffectedProducts pq.StringArray `db:"affected_products"`
	Fixes            pq.StringArray `db:"fixes"`
	Cves             pq.StringArray `db:"cves"`
	References       pq.StringArray `db:"references"`
	RPMs             pq.StringArray `db:"rpms"`
	BuildArtifacts   pq.StringArray `db:"build_artifacts"`
}

// CVE is the DTO struct for `resf.secparse.admin.CVE`
type CVE struct {
	ID        string     `db:"id"`
	CreatedAt *time.Time `db:"created_at"`

	State      int           `db:"state"`
	AdvisoryId sql.NullInt64 `db:"advisory_id"`
	ShortCode  string        `db:"short_code_code"`

	SourceBy   sql.NullString `db:"source_by"`
	SourceLink sql.NullString `db:"source_link"`
}

// AffectedProduct is the DTO struct for `ctlriq.secparse.admin.AffectedProduct`
type AffectedProduct struct {
	ID        int64          `db:"id"`
	ProductID int64          `db:"product_id"`
	CveID     sql.NullString `db:"cve_id"`
	State     int            `db:"state"`
	Version   string         `db:"version"`
	Package   string         `db:"package"`
	Advisory  sql.NullString `db:"advisory"`
}

// Product is the DTO struct for `ctlriq.secparse.admin.Product`
type Product struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`

	CurrentFullVersion string         `db:"current_full_version"`
	RedHatMajorVersion sql.NullInt32  `db:"redhat_major_version"`
	ShortCode          string         `db:"short_code_code"`
	Archs              pq.StringArray `db:"archs"`
}

type BuildReference struct {
	ID                int64  `db:"id"`
	AffectedProductId int64  `db:"affected_product_id"`
	Rpm               string `db:"rpm"`
	SrcRpm            string `db:"src_rpm"`
	CveID             string `db:"cve_id"`
	KojiID            string `db:"koji_id"`
}

type Fix struct {
	ID          int64          `db:"id"`
	Ticket      sql.NullString `db:"ticket"`
	Description sql.NullString `db:"description"`
}

type AdvisoryReference struct {
	ID         int64  `db:"advisory_reference"`
	URL        string `db:"url"`
	AdvisoryId int64  `db:"advisory_id"`
}

type MirrorState struct {
	ShortCode string       `db:"short_code_code"`
	LastSync  sql.NullTime `db:"last_sync"`
}

type AdvisoryCVE struct {
	AdvisoryID int64  `db:"advisory_id"`
	CveID      string `db:"cve_id"`
}

type AdvisoryFix struct {
	AdvisoryID int64 `db:"advisory_id"`
	FixID      int64 `db:"fix_id"`
}

type IgnoredUpstreamPackage struct {
	ID        int64  `db:"id"`
	ShortCode string `db:"short_code_code"`
	Package   string `db:"package"`
}

type AdvisoryRPM struct {
	AdvisoryID int64  `db:"advisory_id"`
	Name       string `db:"name"`
}

type Access interface {
	GetAllShortCodes() ([]*ShortCode, error)
	GetShortCodeByCode(code string) (*ShortCode, error)
	CreateShortCode(code string, mode secparseadminpb.ShortCodeMode) (*ShortCode, error)

	GetAllAdvisories(publishedOnly bool) ([]*Advisory, error)
	// Advisory is a broad entity with lots of fields
	// mustafa: It is in my opinion better to accept the same struct
	// to create and update it.
	// Obviously fields like ID and CreatedAt cannot be overridden
	// The Create and Update methods for advisory do not return
	// the following fields:
	//	 - AffectedProducts
	//	 - Fixes
	//	 - Cves
	//	 - References
	CreateAdvisory(advisory *Advisory) (*Advisory, error)
	// Update cannot override the RedHatIssuedAt field for mirrored advisories
	UpdateAdvisory(advisory *Advisory) (*Advisory, error)
	GetAdvisoryByCodeAndYearAndNum(code string, year int, num int) (*Advisory, error)

	GetAllUnresolvedCVEs() ([]*CVE, error)
	GetAllCVEsWithAllProductsFixed() ([]*CVE, error)
	GetAllCVEsFixedDownstream() ([]*CVE, error)
	GetCVEByID(id string) (*CVE, error)
	CreateCVE(cveId string, state secparseadminpb.CVEState, shortCode string, sourceBy *string, sourceLink *string) (*CVE, error)
	UpdateCVEState(cve string, state secparseadminpb.CVEState) error

	GetProductsByShortCode(code string) ([]*Product, error)
	GetProductByNameAndShortCode(product string, code string) (*Product, error)
	GetProductByID(id int64) (*Product, error)
	CreateProduct(name string, currentFullVersion string, redHatMajorVersion *int32, code string, archs []string) (*Product, error)

	GetAllAffectedProductsByCVE(cve string) ([]*AffectedProduct, error)
	GetAffectedProductByCVEAndPackage(cve string, pkg string) (*AffectedProduct, error)
	GetAffectedProductByAdvisory(advisory string) (*AffectedProduct, error)
	CreateAffectedProduct(productId int64, cveId string, state int, version string, pkg string, advisory *string) (*AffectedProduct, error)
	UpdateAffectedProductStateAndPackageAndAdvisory(id int64, state int, pkg string, advisory *string) error
	DeleteAffectedProduct(id int64) error

	CreateFix(ticket string, description string) (int64, error)

	// This will return nil rather than an error if no rows are found
	GetMirrorStateLastSync(code string) (*time.Time, error)
	UpdateMirrorState(code string, lastSync *time.Time) error

	CreateBuildReference(affectedProductId int64, rpm string, srcRpm string, cveId string, kojiId string) (*BuildReference, error)
	CreateAdvisoryReference(advisoryId int64, url string) error

	GetAllIgnoredPackagesByShortCode(code string) ([]string, error)

	// These add methods is treated like an upsert. They're only added if one doesn't exist
	AddAdvisoryFix(advisoryId int64, fixId int64) error
	AddAdvisoryCVE(advisoryId int64, cveId string) error
	AddAdvisoryRPM(advisoryId int64, name string) error

	Begin() (utils.Tx, error)
	UseTransaction(tx utils.Tx) Access
}
