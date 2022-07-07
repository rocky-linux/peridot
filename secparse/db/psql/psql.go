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

package psql

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	secparseadminpb "peridot.resf.org/secparse/admin/proto/v1"
	"peridot.resf.org/secparse/db"
	"peridot.resf.org/utils"
)

type Access struct {
	db    *sqlx.DB
	query utils.SqlQuery
}

func New() *Access {
	pgx := utils.PgInitx()
	return &Access{
		db:    pgx,
		query: pgx,
	}
}

func (a *Access) GetAllShortCodes() ([]*db.ShortCode, error) {
	var shortCodes []*db.ShortCode
	err := a.query.Select(
		&shortCodes,
		`
			select
				code,
				mode,
				created_at,
				archived_at,
				mirror_from_date,
				redhat_product_prefix
			from short_codes
			order by created_at desc
		`,
	)
	if err != nil {
		return nil, err
	}

	return shortCodes, nil
}

func (a *Access) GetShortCodeByCode(code string) (*db.ShortCode, error) {
	var shortCode db.ShortCode
	err := a.query.Get(&shortCode, "select code, mode, created_at, archived_at, mirror_from_date, redhat_product_prefix from short_codes where code = $1", code)
	if err != nil {
		return nil, err
	}

	return &shortCode, nil
}

func (a *Access) CreateShortCode(code string, mode secparseadminpb.ShortCodeMode) (*db.ShortCode, error) {
	var shortCode db.ShortCode
	err := a.query.Get(&shortCode, "insert into short_codes (code, mode) values ($1, $2) returning code, mode, created_at, archived_at, mirror_from_date, redhat_product_prefix", code, int(mode))
	if err != nil {
		return nil, err
	}

	return &shortCode, nil
}

func (a *Access) GetAllAdvisories(publishedOnly bool) ([]*db.Advisory, error) {
	var advisories []*db.Advisory
	err := a.query.Select(
		&advisories,
		`
		select
			a.id,
			a.created_at,
			a.year,
			a.num,
			a.synopsis,
			a.topic,
			a.severity,
			a.type,
			a.description,
			a.solution,
			a.redhat_issued_at,
			a.short_code_code,
			a.published_at,
			array_remove(array_agg(distinct p.name), NULL) as affected_products,
			array_remove(array_agg(distinct f.ticket), NULL) as fixes,
			array_remove(array_agg(distinct c.source_by || ':::' || c.source_link || ':::' || c.id), NULL) as cves,
			array_remove(array_agg(distinct r.url), NULL) as references,
			array_remove(array_agg(distinct ar.name), NULL) as rpms
		from advisories a
		left join advisory_fixes adf on adf.advisory_id = a.id
		left join fixes f on f.id = adf.fix_id
		left join advisory_cves ac on ac.advisory_id = a.id
		left join cves c on c.id = ac.cve_id
		left join affected_products ap on ap.cve_id = ac.cve_id
		left join products p on ap.product_id = p.id
		left join advisory_references r on r.advisory_id = a.id
		left join advisory_rpms ar on ar.advisory_id = a.id
		where
			($1 is false or a.published_at is not null)
		group by a.id
		order by a.created_at desc
		`,
		publishedOnly,
	)
	if err != nil {
		return nil, err
	}

	return advisories, nil
}

func (a *Access) GetAdvisoryByCodeAndYearAndNum(code string, year int, num int) (*db.Advisory, error) {
	var advisory db.Advisory
	err := a.query.Get(
		&advisory,
		`
		select
			a.id,
			a.created_at,
			a.year,
			a.num,
			a.synopsis,
			a.topic,
			a.severity,
			a.type,
			a.description,
			a.solution,
			a.redhat_issued_at,
			a.short_code_code,
			a.published_at,
			array_remove(array_agg(distinct p.name), NULL) as affected_products,
			array_remove(array_agg(distinct f.ticket), NULL) as fixes,
			array_remove(array_agg(distinct c.source_by || ':::' || c.source_link || ':::' || c.id), NULL) as cves,
			array_remove(array_agg(distinct r.url), NULL) as references,
			array_remove(array_agg(distinct ar.name), NULL) as rpms
		from advisories a
		left join advisory_fixes adf on adf.advisory_id = a.id
		left join fixes f on f.id = adf.fix_id
		left join advisory_cves ac on ac.advisory_id = a.id
		left join cves c on c.id = ac.cve_id
		left join affected_products ap on ap.cve_id = ac.cve_id
		left join products p on ap.product_id = p.id
		left join advisory_references r on r.advisory_id = a.id
		left join advisory_rpms ar on ar.advisory_id = a.id
		where
			a.year = $1
			and a.num = $2
			and a.short_code_code = $3
		group by a.id
		`,
		year,
		num,
		code,
	)
	if err != nil {
		return nil, err
	}

	return &advisory, nil
}

func (a *Access) CreateAdvisory(advisory *db.Advisory) (*db.Advisory, error) {
	var ret db.Advisory

	var redHatIssuedAt *time.Time
	var publishedAt *time.Time

	if advisory.RedHatIssuedAt.Valid {
		redHatIssuedAt = &advisory.RedHatIssuedAt.Time
	}
	if advisory.PublishedAt.Valid {
		publishedAt = &advisory.PublishedAt.Time
	}

	err := a.query.Get(
		&ret,
		`
			insert into advisories
			(year, num, synopsis, topic, severity, type, description, solution,
			redhat_issued_at, short_code_code, published_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			returning
				id,
				created_at,
				year,
				num,
				synopsis,
				topic,
				severity,
				type,
				description,
				solution,
				redhat_issued_at,
				short_code_code,
				published_at
		`,
		advisory.Year,
		advisory.Num,
		advisory.Synopsis,
		advisory.Topic,
		advisory.Severity,
		advisory.Type,
		advisory.Description,
		advisory.Solution,
		redHatIssuedAt,
		advisory.ShortCodeCode,
		publishedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) UpdateAdvisory(advisory *db.Advisory) (*db.Advisory, error) {
	var ret db.Advisory

	var publishedAt *time.Time

	if advisory.PublishedAt.Valid {
		publishedAt = &advisory.PublishedAt.Time
	}

	err := a.query.Get(
		&ret,
		`
			update advisories
			set
				year = $1,
				num = $2,
				synopsis = $3,
				topic = $4,
				severity = $5,
				type = $6,
				description = $7,
				solution = $8,
				short_code_code = $9,
				published_at = $10
			where
				id = $11
			returning
				id,
				created_at,
				year,
				num,
				synopsis,
				topic,
				severity,
				type,
				description,
				solution,
				redhat_issued_at,
				short_code_code,
				published_at
		`,
		advisory.Year,
		advisory.Num,
		advisory.Synopsis,
		advisory.Topic,
		advisory.Severity,
		advisory.Type,
		advisory.Description,
		advisory.Solution,
		advisory.ShortCodeCode,
		publishedAt,
		advisory.ID,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetAllUnresolvedCVEs() ([]*db.CVE, error) {
	var cves []*db.CVE
	err := a.query.Select(&cves, "select id, created_at, state, short_code_code, source_by, source_link from cves where state in (1, 2, 8, 9)")
	if err != nil {
		return nil, err
	}

	return cves, nil
}

func (a *Access) GetAllCVEsWithAllProductsFixed() ([]*db.CVE, error) {
	var cves []*db.CVE
	err := a.query.Select(
		&cves,
		`
			select
				c.id,
				c.created_at,
				c.state,
				c.short_code_code,
				c.source_by,
				c.source_link
			from cves c
			where
				c.id in (select cve_id from affected_products where state = 3)
				and c.state in (1, 2, 3, 4)
		`,
	)
	if err != nil {
		return nil, err
	}

	return cves, nil
}

func (a *Access) GetAllCVEsFixedDownstream() ([]*db.CVE, error) {
	var cves []*db.CVE
	err := a.query.Select(
		&cves,
		`
			select
				c.id,
				c.created_at,
				c.state,
				c.short_code_code,
				c.source_by,
				c.source_link
			from cves c
			where
				c.state = 4
		`,
	)
	if err != nil {
		return nil, err
	}

	return cves, nil
}

func (a *Access) GetCVEByID(id string) (*db.CVE, error) {
	var cve db.CVE
	err := a.query.Get(&cve, "select id, created_at, state, short_code_code, source_by, source_link from cves where id = $1", id)
	if err != nil {
		return nil, err
	}

	return &cve, nil
}

func (a *Access) CreateCVE(cveId string, state secparseadminpb.CVEState, shortCode string, sourceBy *string, sourceLink *string) (*db.CVE, error) {
	var cve db.CVE
	err := a.query.Get(&cve, "insert into cves (id, state, short_code_code, source_by, source_link) values ($1, $2, $3, $4, $5) returning id, created_at, state, short_code_code, source_by, source_link", cveId, int(state), shortCode, sourceBy, sourceLink)
	if err != nil {
		return nil, err
	}

	return &cve, nil
}

func (a *Access) UpdateCVEState(cve string, state secparseadminpb.CVEState) error {
	_, err := a.query.Exec(
		`
			update cves
			set
				state = $1
			where id = $2
		`,
		state,
		cve,
	)
	return err
}

func (a *Access) GetProductsByShortCode(code string) ([]*db.Product, error) {
	var products []*db.Product
	err := a.query.Select(&products, "select id, name, current_full_version, redhat_major_version, short_code_code, archs from products where short_code_code = $1 and (eol_at < now() or eol_at is null)", code)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (a *Access) GetProductByNameAndShortCode(name string, code string) (*db.Product, error) {
	var product db.Product
	err := a.query.Get(&product, "select id, name, current_full_version, redhat_major_version, short_code_code, archs from products where name = $1 and short_code_code = $2", name, code)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (a *Access) GetProductByID(id int64) (*db.Product, error) {
	var product db.Product
	err := a.query.Get(&product, "select id, name, current_full_version, redhat_major_version, short_code_code, archs from products where id = $1", id)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (a *Access) CreateProduct(name string, currentFullVersion string, redHatMajorVersion *int32, code string, archs []string) (*db.Product, error) {
	var product db.Product
	err := a.query.Get(&product, "insert into products (name, current_full_version, redhat_major_version, short_code_code, archs) values ($1, $2, $3, $4) returning id, name, current_full_version, redhat_major_version, short_code_code, archs", name, currentFullVersion, redHatMajorVersion, code, archs)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (a *Access) GetAllAffectedProductsByCVE(cve string) ([]*db.AffectedProduct, error) {
	var affectedProducts []*db.AffectedProduct
	err := a.query.Select(&affectedProducts, "select id, product_id, cve_id, state, version, package, advisory from affected_products where cve_id = $1", cve)
	if err != nil {
		return nil, err
	}

	return affectedProducts, nil
}

func (a *Access) GetAffectedProductByCVEAndPackage(cve string, pkg string) (*db.AffectedProduct, error) {
	var affectedProduct db.AffectedProduct
	err := a.query.Get(&affectedProduct, "select id, product_id, cve_id, state, version, package, advisory from affected_products where cve_id = $1 and package = $2", cve, pkg)
	if err != nil {
		return nil, err
	}

	return &affectedProduct, nil
}

func (a *Access) GetAffectedProductByAdvisory(advisory string) (*db.AffectedProduct, error) {
	var affectedProduct db.AffectedProduct
	err := a.query.Get(&affectedProduct, "select id, product_id, cve_id, state, version, package, advisory from affected_products where advisory = $1", advisory)
	if err != nil {
		return nil, err
	}

	return &affectedProduct, nil
}

func (a *Access) CreateAffectedProduct(productId int64, cveId string, state int, version string, pkg string, advisory *string) (*db.AffectedProduct, error) {
	var affectedProduct db.AffectedProduct
	err := a.query.Get(&affectedProduct, "insert into affected_products (product_id, cve_id, state, version, package, advisory) values ($1, $2, $3, $4, $5, $6) returning id, product_id, cve_id, state, version, package, advisory", productId, cveId, state, version, pkg, advisory)
	if err != nil {
		return nil, err
	}

	return &affectedProduct, nil
}

func (a *Access) UpdateAffectedProductStateAndPackageAndAdvisory(id int64, state int, pkg string, advisory *string) error {
	_, err := a.query.Exec(
		`
			update affected_products
			set
				state = $1,
				package = $2,
				advisory = $3
			where id = $4
		`,
		state,
		pkg,
		advisory,
		id,
	)
	return err
}

func (a *Access) DeleteAffectedProduct(id int64) error {
	_, err := a.query.Exec(
		`
			delete from affected_products
			where id = $1
		`,
		id,
	)
	return err
}

func (a *Access) CreateFix(ticket string, description string) (int64, error) {
	var id int64
	err := a.query.Get(&id, "insert into fixes (ticket, description) values ($1, $2) returning id", ticket, description)
	return id, err
}

func (a *Access) GetMirrorStateLastSync(code string) (*time.Time, error) {
	var lastSync time.Time
	row := a.query.QueryRowx("select last_sync from mirror_state where short_code_code = $1", code)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	err := row.Scan(&lastSync)
	if err != nil {
		return nil, err
	}

	return &lastSync, nil
}

func (a *Access) UpdateMirrorState(code string, lastSync *time.Time) error {
	_, err := a.query.Exec(
		`
			insert into mirror_state (short_code_code, last_sync)
			values ($1, $2)
			on conflict (short_code_code) do
			update
				set last_sync = EXCLUDED.last_sync
		`,
		code,
		lastSync,
	)
	return err
}

func (a *Access) CreateBuildReference(affectedProductId int64, rpm string, srcRpm string, cveId string, kojiId string) (*db.BuildReference, error) {
	var buildReference db.BuildReference
	err := a.query.Get(
		&buildReference,
		`
			insert into build_references
			(affected_product_id, rpm, src_rpm, cve_id, koji_id)
			values ($1, $2, $3, $4, $5)
			returning id, affected_product_id, rpm, src_rpm, cve_id, koji_id
		`,
		affectedProductId,
		rpm,
		srcRpm,
		cveId,
		kojiId,
	)
	if err != nil {
		return nil, err
	}

	return &buildReference, nil
}

func (a *Access) CreateAdvisoryReference(advisoryId int64, url string) error {
	_, err := a.query.Exec("insert into advisory_references (advisory_id, url) values ($1, $2)", advisoryId, url)
	return err
}

func (a *Access) GetAllIgnoredPackagesByShortCode(code string) ([]string, error) {
	var packages []string
	err := a.query.Select(&packages, "select package from ignored_upstream_packages where short_code_code = $1", code)
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func (a *Access) AddAdvisoryFix(advisoryId int64, fixId int64) error {
	_, err := a.query.Exec("insert into advisory_fixes (advisory_id, fix_id) values ($1, $2) on conflict do nothing", advisoryId, fixId)
	if err != nil {
		return err
	}

	return nil
}

func (a *Access) AddAdvisoryCVE(advisoryId int64, cveId string) error {
	_, err := a.query.Exec("insert into advisory_cves (advisory_id, cve_id) values ($1, $2) on conflict do nothing", advisoryId, cveId)
	if err != nil {
		return err
	}

	return nil
}

func (a *Access) AddAdvisoryRPM(advisoryId int64, name string) error {
	_, err := a.query.Exec("insert into advisory_rpms (advisory_id, name) values ($1, $2) on conflict do nothing", advisoryId, name)
	if err != nil {
		return err
	}

	return nil
}

func (a *Access) Begin() (utils.Tx, error) {
	tx, err := a.db.Beginx()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (a *Access) UseTransaction(tx utils.Tx) db.Access {
	newAccess := *a
	newAccess.query = tx

	return &newAccess
}
