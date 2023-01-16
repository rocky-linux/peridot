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

package apollopsql

import (
	"database/sql"
	"github.com/jmoiron/sqlx/types"
	apollodb "peridot.resf.org/apollo/db"
	apollopb "peridot.resf.org/apollo/pb"
	"time"

	"github.com/jmoiron/sqlx"
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

func (a *Access) GetAllShortCodes() ([]*apollodb.ShortCode, error) {
	var shortCodes []*apollodb.ShortCode
	err := a.query.Select(
		&shortCodes,
		`
			select
				code,
				mode,
				created_at,
				archived_at
			from short_codes
			order by created_at desc
		`,
	)
	if err != nil {
		return nil, err
	}

	return shortCodes, nil
}

func (a *Access) GetShortCodeByCode(code string) (*apollodb.ShortCode, error) {
	var shortCode apollodb.ShortCode
	err := a.query.Get(&shortCode, "select code, mode, created_at from short_codes where code = $1", code)
	if err != nil {
		return nil, err
	}

	return &shortCode, nil
}

func (a *Access) CreateShortCode(code string, mode apollopb.ShortCode_Mode) (*apollodb.ShortCode, error) {
	var shortCode apollodb.ShortCode
	err := a.query.Get(&shortCode, "insert into short_codes (code, mode) values ($1, $2) returning code, mode, created_at", code, int(mode))
	if err != nil {
		return nil, err
	}

	return &shortCode, nil
}

func (a *Access) GetAllAdvisories(filters *apollopb.AdvisoryFilters, page int32, limit int32) ([]*apollodb.Advisory, error) {
	if filters == nil {
		filters = &apollopb.AdvisoryFilters{}
	}

	var advisories []*apollodb.Advisory
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
            a.reboot_suggested,
            a.published_at,
            array_remove(array_agg(distinct p.name), NULL) as affected_products,
            (select array_agg(distinct(f.ticket || ':::' || f.source_by || ':::' || f.source_link || ':::' || f.description)) from advisory_fixes adf inner join fixes f on f.id = adf.fix_id where adf.advisory_id = a.id) as fixes,
            (select array_agg(distinct(
                case when c.content is null then c.source_by || ':::' || c.source_link || ':::' || c.id || ':::::::::'
                     else c.source_by || ':::' || c.source_link || ':::' || c.id || ':::' || jsonb_extract_path_text(c.content, 'cvss3', 'cvss3_scoring_vector') || ':::' || jsonb_extract_path_text(c.content, 'cvss3', 'cvss3_base_score') || ':::' || jsonb_extract_path_text(c.content, 'cwe')
                end
            )) from advisory_cves ac inner join cves c on c.id = ac.cve_id where ac.advisory_id = a.id) as cves,
            (select array_agg(distinct(url)) from advisory_references where advisory_id = a.id) as references,
            case when $4 :: bool = true then array(select distinct concat(rpm, ':::', src_rpm) from build_references where affected_product_id in (select id from affected_products where advisory = 'RH' || (case when a.type=1 then 'SA' when a.type=2 then 'BA' else 'EA' end) || '-' || a.year || ':' || a.num))
                 else array [] :: text[]
            end as build_artifacts,
            case when $7 :: bool = true then array(select distinct(ar.name || ':::' || p.name) from advisory_rpms ar inner join products p on p.id = ar.product_id where advisory_id = a.id)
                 else array [] :: text[]
            end as rpms,
            count(a.*) over() as total
        from advisories a
        inner join affected_products ap on ap.advisory = 'RH' || (case when a.type=1 then 'SA' when a.type=2 then 'BA' else 'EA' end) || '-' || a.year || ':' || a.num
        inner join products p on ap.product_id = p.id
        where
            ($1 :: text is null or p.name = $1 :: text)
            and ($2 :: timestamp is null or a.published_at < $2 :: timestamp)
            and ($3 :: timestamp is null or a.published_at > $3 :: timestamp)
            and (a.published_at is not null or $4 :: bool = true)
            and ($5 :: text is null or exists (select cve_id from advisory_cves where advisory_id = a.id and cve_id ilike '%' || $5 :: text || '%'))
            and ($6 :: text is null or a.synopsis ilike '%' || $6 :: text || '%')
            and ($8 :: text is null or ((a.synopsis ilike '%' || $8 :: text || '%') or (a.topic ilike '%' || $8 :: text || '%') or (a.description ilike '%' || $8 :: text || '%') or (a.solution ilike '%' || $8 :: text || '%') or exists (select cve_id from advisory_cves where advisory_id = a.id and cve_id ilike '%' || $8 :: text || '%') or (a.short_code_code || (case when a.type=1 then 'SA' when a.type=2 then 'BA' else 'EA' end) || '-' || a.year || ':' || a.num ilike '%' || $8 :: text || '%')))
            and ($9 :: numeric = 0 or a.severity = $9 :: numeric)
            and ($10 :: numeric = 0 or a.type = $10 :: numeric)
        group by a.id
        order by a.published_at desc
        limit $11 offset $12
		`,
		utils.StringValueToNullString(filters.Product),
		utils.TimestampToNullTime(filters.Before),
		utils.TimestampToNullTime(filters.After),
		utils.BoolValueP(filters.IncludeUnpublished),
		utils.StringValueToNullString(filters.Cve),
		utils.StringValueToNullString(filters.Synopsis),
		utils.BoolValueP(filters.IncludeRpms),
		utils.StringValueToNullString(filters.Keyword),
		int32(filters.Severity),
		int32(filters.Type),
		utils.UnlimitedLimit(limit),
		utils.GetOffset(page, limit),
	)
	if err != nil {
		return nil, err
	}

	return advisories, nil
}

func (a *Access) GetAdvisoryByCodeAndYearAndNum(code string, year int, num int) (*apollodb.Advisory, error) {
	var advisory apollodb.Advisory
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
            a.reboot_suggested,
			a.published_at,
			array_remove(array_agg(distinct p.name), NULL) as affected_products,
			(select array_agg(distinct(f.ticket || ':::' || f.source_by || ':::' || f.source_link || ':::' || f.description)) from advisory_fixes adf inner join fixes f on f.id = adf.fix_id where adf.advisory_id = a.id) as fixes,
            (select array_agg(distinct(
                case when c.content is null then c.source_by || ':::' || c.source_link || ':::' || c.id || ':::::::::'
                     else c.source_by || ':::' || c.source_link || ':::' || c.id || ':::' || jsonb_extract_path_text(c.content, 'cvss3', 'cvss3_scoring_vector') || ':::' || jsonb_extract_path_text(c.content, 'cvss3', 'cvss3_base_score') || ':::' || jsonb_extract_path_text(c.content, 'cwe')
                end
            )) from advisory_cves ac inner join cves c on c.id = ac.cve_id where ac.advisory_id = a.id) as cves,
			(select array_agg(distinct(url)) from advisory_references where advisory_id = a.id) as references,
			(select array_agg(distinct(ar.name || ':::' || p.name)) from advisory_rpms ar inner join products p on p.id = ar.product_id where advisory_id = a.id) as rpms
		from advisories a
		inner join affected_products ap on ap.advisory = 'RH' || (case when a.type=1 then 'SA' when a.type=2 then 'BA' else 'EA' end) || '-' || a.year || ':' || a.num
		inner join products p on ap.product_id = p.id
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

func (a *Access) CreateAdvisory(advisory *apollodb.Advisory) (*apollodb.Advisory, error) {
	var ret apollodb.Advisory

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
			redhat_issued_at, short_code_code, reboot_suggested, published_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
                reboot_suggested,
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
		advisory.RebootSuggested,
		publishedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) UpdateAdvisory(advisory *apollodb.Advisory) (*apollodb.Advisory, error) {
	var ret apollodb.Advisory

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
                reboot_suggested = $10,
				published_at = $11
			where
				id = $12
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
                reboot_suggested,
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
		advisory.RebootSuggested,
		publishedAt,
		advisory.ID,
	)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *Access) GetAllUnresolvedCVEs() ([]*apollodb.CVE, error) {
	var cves []*apollodb.CVE
	err := a.query.Select(
		&cves,
		`
        select
            c.id,
            c.created_at,
            c.short_code_code,
            c.source_by,
            c.source_link,
            c.content,
            ap.id as affected_product_id
        from cves c
        left join affected_products ap on ap.cve_id = c.id
        where (ap.state is null or ap.state in (1, 2, 8, 9))
        `,
	)
	if err != nil {
		return nil, err
	}

	return cves, nil
}

func (a *Access) GetPendingAffectedProducts() ([]*apollodb.AffectedProduct, error) {
	var ret []*apollodb.AffectedProduct
	err := a.query.Select(
		&ret,
		`
            select
                ap.id,
                ap.product_id,
                ap.cve_id,
                ap.state,
                ap.version,
                ap.package,
                ap.advisory
            from affected_products ap
            where ap.state = 3
		`,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *Access) GetAllCVEsFixedDownstream() ([]*apollodb.CVE, error) {
	var cves []*apollodb.CVE
	err := a.query.Select(
		&cves,
		`
			select
				c.id,
				c.created_at,
				c.short_code_code,
				c.source_by,
				c.source_link,
                c.content,
                ap.id as affected_product_id
			from cves c
            inner join affected_products ap on ap.cve_id = c.id
			where
				ap.state = 4
		`,
	)
	if err != nil {
		return nil, err
	}

	return cves, nil
}

func (a *Access) GetCVEByID(id string) (*apollodb.CVE, error) {
	var cve apollodb.CVE
	err := a.query.Get(&cve, "select id, created_at, short_code_code, source_by, source_link, content from cves where id = $1", id)
	if err != nil {
		return nil, err
	}

	return &cve, nil
}

func (a *Access) GetAllCVEs() ([]*apollodb.CVE, error) {
	var cves []*apollodb.CVE
	err := a.query.Select(&cves, "select id, created_at, short_code_code, source_by, source_link, content from cves")
	if err != nil {
		return nil, err
	}

	return cves, nil
}

func (a *Access) CreateCVE(cveId string, shortCode string, sourceBy *string, sourceLink *string, content types.NullJSONText) (*apollodb.CVE, error) {
	var cve apollodb.CVE
	err := a.query.Get(&cve, "insert into cves (id, short_code_code, source_by, source_link, content) values ($1, $2, $3, $4, $5) returning id, created_at, short_code_code, source_by, source_link, content", cveId, shortCode, sourceBy, sourceLink, content)
	if err != nil {
		return nil, err
	}

	return &cve, nil
}

func (a *Access) SetCVEContent(cveId string, content types.JSONText) error {
	_, err := a.query.Exec("update cves set content = $1 where id = $2", content, cveId)
	return err
}

func (a *Access) GetProductsByShortCode(code string) ([]*apollodb.Product, error) {
	var products []*apollodb.Product
	err := a.query.Select(
		&products,
		`
        select
            id,
            name,
            current_full_version,
            redhat_major_version,
            short_code_code,
            archs,
            mirror_from_date,
            redhat_product_prefix,
            cpe,
            eol_at,
            build_system,
            build_system_endpoint,
            koji_compose,
            koji_module_compose,
            peridot_project_id
        from products
        where
            short_code_code = $1
            and (eol_at < now() or eol_at is null)
        `,
		code,
	)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (a *Access) GetProductByNameAndShortCode(name string, code string) (*apollodb.Product, error) {
	var product apollodb.Product
	err := a.query.Get(
		&product,
		`
        select
            id,
            name,
            current_full_version,
            redhat_major_version,
            short_code_code,
            archs,
            mirror_from_date,
            redhat_product_prefix,
            cpe,
            eol_at,
            build_system,
            build_system_endpoint,
            koji_compose,
            koji_module_compose,
            peridot_project_id
        from products
        where
            name = $1
            and short_code_code = $2
        `,
		name,
		code,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (a *Access) GetProductByID(id int64) (*apollodb.Product, error) {
	var product apollodb.Product
	err := a.query.Get(
		&product,
		`
        select
            id,
            name,
            current_full_version,
            redhat_major_version,
            short_code_code,
            archs,
            mirror_from_date,
            redhat_product_prefix,
            cpe,
            eol_at,
            build_system,
            build_system_endpoint,
            koji_compose,
            koji_module_compose,
            peridot_project_id
        from products
        where
            id = $1
        `,
		id,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (a *Access) CreateProduct(name string, currentFullVersion string, redHatMajorVersion *int32, code string, archs []string) (*apollodb.Product, error) {
	var product apollodb.Product
	err := a.query.Get(&product, "insert into products (name, current_full_version, redhat_major_version, short_code_code, archs) values ($1, $2, $3, $4) returning id, name, current_full_version, redhat_major_version, short_code_code, archs", name, currentFullVersion, redHatMajorVersion, code, archs)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (a *Access) GetAllAffectedProductsByCVE(cve string) ([]*apollodb.AffectedProduct, error) {
	var affectedProducts []*apollodb.AffectedProduct
	err := a.query.Select(&affectedProducts, "select id, product_id, cve_id, state, version, package, advisory from affected_products where cve_id = $1", cve)
	if err != nil {
		return nil, err
	}

	return affectedProducts, nil
}

func (a *Access) GetAffectedProductByCVEAndPackage(cve string, pkg string) (*apollodb.AffectedProduct, error) {
	var affectedProduct apollodb.AffectedProduct
	err := a.query.Get(&affectedProduct, "select id, product_id, cve_id, state, version, package, advisory from affected_products where cve_id = $1 and package = $2", cve, pkg)
	if err != nil {
		return nil, err
	}

	return &affectedProduct, nil
}

func (a *Access) GetAffectedProductByAdvisory(advisory string) (*apollodb.AffectedProduct, error) {
	var affectedProduct apollodb.AffectedProduct
	err := a.query.Get(&affectedProduct, "select id, product_id, cve_id, state, version, package, advisory from affected_products where advisory = $1", advisory)
	if err != nil {
		return nil, err
	}

	return &affectedProduct, nil
}

func (a *Access) GetAffectedProductByID(id int64) (*apollodb.AffectedProduct, error) {
	var affectedProduct apollodb.AffectedProduct
	err := a.query.Get(&affectedProduct, "select id, product_id, cve_id, state, version, package, advisory from affected_products where id = $1", id)
	if err != nil {
		return nil, err
	}

	return &affectedProduct, nil
}

func (a *Access) CreateAffectedProduct(productId int64, cveId string, state int, version string, pkg string, advisory *string) (*apollodb.AffectedProduct, error) {
	var affectedProduct apollodb.AffectedProduct
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

func (a *Access) CreateFix(ticket string, sourceBy string, sourceLink, description string) (int64, error) {
	var id int64
	err := a.query.Get(&id, "insert into fixes (ticket, source_by, source_link, description) values ($1, $2, $3, $4) returning id", ticket, sourceBy, sourceLink, description)
	return id, err
}

func (a *Access) GetMirrorState(code string) (*apollodb.MirrorState, error) {
	var lastSync apollodb.MirrorState
	err := a.query.Get(&lastSync, "select short_code_code, last_sync, errata_after from mirror_state where short_code_code = $1", code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

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

func (a *Access) UpdateMirrorStateErrata(code string, lastSync *time.Time) error {
	_, err := a.query.Exec(
		`
			insert into mirror_state (short_code_code, errata_after)
			values ($1, $2)
			on conflict (short_code_code) do
			update
				set errata_after = EXCLUDED.errata_after
		`,
		code,
		lastSync,
	)
	return err
}

func (a *Access) GetMaxLastSync() (*time.Time, error) {
	var lastSync time.Time
	err := a.query.Get(&lastSync, "select max(last_sync) from mirror_state")
	if err != nil {
		return nil, err
	}

	return &lastSync, nil
}

func (a *Access) CreateBuildReference(affectedProductId int64, rpm string, srcRpm string, cveId string, sha256Sum string, kojiId *string, peridotId *string) (*apollodb.BuildReference, error) {
	var buildReference apollodb.BuildReference
	err := a.query.Get(
		&buildReference,
		`
			insert into build_references
			(affected_product_id, rpm, src_rpm, cve_id, sha256_sum, koji_id, peridot_id)
			values ($1, $2, $3, $4, $5, $6, $7)
			returning id, affected_product_id, rpm, src_rpm, cve_id, sha256_sum, koji_id, peridot_id
		`,
		affectedProductId,
		rpm,
		srcRpm,
		cveId,
		sha256Sum,
		kojiId,
		peridotId,
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

func (a *Access) GetAllIgnoredPackagesByProductID(productID int64) ([]string, error) {
	var packages []string
	err := a.query.Select(&packages, "select package from ignored_upstream_packages where product_id = $1", productID)
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func (a *Access) GetAllRebootSuggestedPackages() ([]string, error) {
	var packages []string
	err := a.query.Select(&packages, "select name from reboot_suggested_packages")
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

func (a *Access) AddAdvisoryRPM(advisoryId int64, name string, productID int64) error {
	_, err := a.query.Exec("insert into advisory_rpms (advisory_id, name, product_id) values ($1, $2, $3) on conflict do nothing", advisoryId, name, productID)
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

func (a *Access) UseTransaction(tx utils.Tx) apollodb.Access {
	newAccess := *a
	newAccess.query = tx

	return &newAccess
}
