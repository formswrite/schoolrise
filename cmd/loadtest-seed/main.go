package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type guineaDistrict struct {
	Name    string `json:"name"`
	Schools int    `json:"schools"`
}

const (
	regionCode  = "lt-region"
	periodCode  = "lt-period"
	niveauCode  = "lt-CE1"
	formPubID   = "lt-form-public"
	campaignPub = "lt-campaign"
	pgUser      = "schoolrise"
)

type config struct {
	host             string
	port             int
	password         string
	districts        int
	schools          int
	studentsTot      int
	purge            bool
	skipFixtures     bool
	guineaModel      bool
	guineaFile       string
	studentsPerSchl  int
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.host, "host", "localhost", "")
	flag.IntVar(&cfg.port, "port", 5433, "")
	flag.StringVar(&cfg.password, "password", os.Getenv("POSTGRES_PASSWORD"), "")
	flag.IntVar(&cfg.districts, "districts", 30, "")
	flag.IntVar(&cfg.schools, "schools", 200, "")
	flag.IntVar(&cfg.studentsTot, "students", 200_000, "")
	flag.BoolVar(&cfg.purge, "purge", false, "")
	flag.BoolVar(&cfg.skipFixtures, "skip-fixtures", false, "")
	flag.BoolVar(&cfg.guineaModel, "guinea", false, "use real Guinea district + school counts from -guinea-file")
	flag.StringVar(&cfg.guineaFile, "guinea-file", "tools/loadtest/seed/guinea_districts.json", "")
	flag.IntVar(&cfg.studentsPerSchl, "students-per-school", 200, "average students per school in -guinea mode")
	flag.Parse()

	if cfg.password == "" {
		log.Fatal("POSTGRES_PASSWORD not set")
	}

	ctx := context.Background()

	if cfg.purge {
		mustPurge(ctx, cfg)
		log.Println("purge complete")
		return
	}

	t := time.Now()
	mustPurge(ctx, cfg)
	mustSeed(ctx, cfg)
	log.Printf("seed complete in %s", time.Since(t).Round(time.Millisecond))
}

func dsn(cfg config, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", pgUser, cfg.password, cfg.host, cfg.port, db)
}

func pool(ctx context.Context, cfg config, db string) *pgxpool.Pool {
	p, err := pgxpool.New(ctx, dsn(cfg, db))
	if err != nil {
		log.Fatalf("connect %s: %v", db, err)
	}
	return p
}

func mustExec(ctx context.Context, p *pgxpool.Pool, sql string, args ...any) {
	if _, err := p.Exec(ctx, sql, args...); err != nil {
		log.Fatalf("exec failed: %v\nsql: %s", err, sql)
	}
}

func mustOne[T any](ctx context.Context, p *pgxpool.Pool, sql string, args ...any) T {
	var v T
	if err := p.QueryRow(ctx, sql, args...).Scan(&v); err != nil {
		log.Fatalf("queryone failed: %v\nsql: %s", err, sql)
	}
	return v
}

func mustSeed(ctx context.Context, cfg config) {
	tenancy := pool(ctx, cfg, "tenancy")
	defer tenancy.Close()

	mustExec(ctx, tenancy, `
        INSERT INTO hierarchy_levels (code, label, parent_level_code, depth, sort_order) VALUES
            ('region',      'Region',      NULL,         0,  0),
            ('district',    'District',    'region',     1, 10),
            ('institution', 'Institution', 'district',   2, 20)
        ON CONFLICT (code) DO NOTHING`)

	regionID := mustOne[int64](ctx, tenancy, `
        INSERT INTO hierarchy_nodes (parent_id, level, code, label)
        VALUES (NULL, 'region', $1, 'Load Test Region')
        RETURNING id`, regionCode)
	log.Printf("region: id=%d", regionID)

	if cfg.guineaModel {
		dists := loadGuineaDistricts(cfg.guineaFile)
		log.Printf("loaded %d Guinea districts, %d total schools, %d students/school avg",
			len(dists), totalGuineaSchools(dists), cfg.studentsPerSchl)
		seedGuineaHierarchy(ctx, tenancy, regionID, dists)
		cfg.districts = len(dists)
		cfg.schools = totalGuineaSchools(dists)
		cfg.studentsTot = cfg.schools * cfg.studentsPerSchl
	} else {
		mustExec(ctx, tenancy, `
            INSERT INTO hierarchy_nodes (parent_id, level, code, label)
            SELECT $1, 'district', 'lt-d' || lpad(g::text, 3, '0'), 'LT District ' || g
            FROM generate_series(1, $2) g`, regionID, cfg.districts)

		mustExec(ctx, tenancy, `
            WITH dists AS (
                SELECT id, row_number() OVER (ORDER BY id) AS n
                FROM hierarchy_nodes WHERE level = 'district' AND parent_id = $1 AND deleted_at IS NULL
            )
            INSERT INTO hierarchy_nodes (parent_id, level, code, label)
            SELECT d.id, 'institution', 'lt-s' || lpad(g::text, 5, '0'), 'LT School ' || g
            FROM generate_series(1, $2) g
            CROSS JOIN dists d
            WHERE d.n = 1 + ((g - 1) % $3)`, regionID, cfg.schools, cfg.districts)
	}

	log.Printf("districts + schools inserted")

	mustExec(ctx, tenancy, `
        INSERT INTO hierarchy_closure (ancestor_id, descendant_id, depth)
        WITH RECURSIVE walk AS (
            SELECT id AS ancestor_id, id AS descendant_id, 0 AS depth
              FROM hierarchy_nodes WHERE deleted_at IS NULL
            UNION ALL
            SELECT w.ancestor_id, n.id, w.depth + 1
              FROM walk w
              JOIN hierarchy_nodes n ON n.parent_id = w.descendant_id
             WHERE n.deleted_at IS NULL
        )
        SELECT ancestor_id, descendant_id, depth FROM walk
        ON CONFLICT DO NOTHING`)
	log.Println("closure rebuilt")

	type schoolInfo struct {
		ID  int64
		Idx int64
	}
	rows, err := tenancy.Query(ctx, `
        SELECT id, row_number() OVER (ORDER BY id)
        FROM hierarchy_nodes
        WHERE level = 'institution' AND code LIKE 'lt-s%' AND deleted_at IS NULL
        ORDER BY id`)
	if err != nil {
		log.Fatalf("query schools: %v", err)
	}
	schools := []schoolInfo{}
	for rows.Next() {
		var s schoolInfo
		if err := rows.Scan(&s.ID, &s.Idx); err != nil {
			log.Fatal(err)
		}
		schools = append(schools, s)
	}
	rows.Close()
	log.Printf("schools fetched: %d", len(schools))

	academics := pool(ctx, cfg, "academics")
	defer academics.Close()

	periodID := mustOne[int64](ctx, academics, `
        INSERT INTO academic_periods (code, label, starts_on, ends_on)
        VALUES ($1, 'Load Test Period', '2025-09-01', '2026-06-30')
        RETURNING id`, periodCode)
	niveauID := mustOne[int64](ctx, academics, `
        INSERT INTO niveaux (code, label, sort_order)
        VALUES ($1, 'LT CE1', 100)
        RETURNING id`, niveauCode)
	log.Printf("period=%d niveau=%d", periodID, niveauID)

	schoolIDs := make([]int64, len(schools))
	for i, s := range schools {
		schoolIDs[i] = s.ID
	}

	mustExec(ctx, academics, `
        INSERT INTO classes (period_id, niveau_id, institution_id, code, label, capacity)
        SELECT $1, $2, unnest($3::bigint[]), 'LT-CE1-A', 'LT CE1-A', 1500`,
		periodID, niveauID, schoolIDs)
	log.Println("classes inserted")

	people := pool(ctx, cfg, "people")
	defer people.Close()

	t := time.Now()
	mustExec(ctx, people, `
        INSERT INTO persons (full_name, given_name, family_name, gender)
        SELECT
            'LT Student ' || g,
            'LT' || g,
            'S' || (1 + ((g - 1) / $2)),
            (ARRAY['F','M','U'])[1 + (g % 3)]
        FROM generate_series(1, $1) g`,
		cfg.studentsTot, cfg.studentsTot/cfg.schools+1)
	log.Printf("persons inserted in %s", time.Since(t).Round(time.Millisecond))

	mustExec(ctx, people, `
        CREATE TEMP TABLE lt_school_map (n INT PRIMARY KEY, school_id BIGINT) ON COMMIT DROP`)
	if _, err := people.Exec(ctx, `BEGIN`); err != nil {
		log.Fatal(err)
	}

	conn, err := people.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE lt_school_map (n INT PRIMARY KEY, school_id BIGINT) ON COMMIT DROP`); err != nil {
		log.Fatal(err)
	}
	mapRows := make([][]any, len(schools))
	for i, s := range schools {
		mapRows[i] = []any{int(s.Idx), s.ID}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"lt_school_map"}, []string{"n", "school_id"}, pgx.CopyFromRows(mapRows)); err != nil {
		log.Fatalf("copy school map: %v", err)
	}

	t = time.Now()
	if _, err := tx.Exec(ctx, `
        INSERT INTO students (person_id, institution_id, student_code)
        SELECT
            p.id,
            m.school_id,
            'LT-' || m.school_id || '-' || row_number() OVER (PARTITION BY m.school_id ORDER BY p.id)
        FROM (
            SELECT id, row_number() OVER (ORDER BY id) AS n
            FROM persons
            WHERE full_name LIKE 'LT Student %'
        ) p
        JOIN lt_school_map m ON m.n = 1 + ((p.n - 1) % $1)`,
		len(schools)); err != nil {
		log.Fatalf("students insert: %v", err)
	}
	log.Printf("students inserted in %s", time.Since(t).Round(time.Millisecond))

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit students: %v", err)
	}

	type studentRow struct {
		ID, School int64
	}
	rows, err = people.Query(ctx, `SELECT id, institution_id FROM students WHERE student_code LIKE 'LT-%' ORDER BY id`)
	if err != nil {
		log.Fatal(err)
	}
	studentIDs := make([]int64, 0, cfg.studentsTot)
	studentToSchool := make(map[int64]int64, cfg.studentsTot)
	for rows.Next() {
		var s studentRow
		if err := rows.Scan(&s.ID, &s.School); err != nil {
			log.Fatal(err)
		}
		studentIDs = append(studentIDs, s.ID)
		studentToSchool[s.ID] = s.School
	}
	rows.Close()
	log.Printf("students fetched: %d", len(studentIDs))

	classRows, err := academics.Query(ctx, `SELECT id, institution_id FROM classes WHERE code = 'LT-CE1-A'`)
	if err != nil {
		log.Fatal(err)
	}
	schoolToClass := make(map[int64]int64, len(schools))
	classIDs := make([]int64, 0, len(schools))
	for classRows.Next() {
		var cid, sid int64
		if err := classRows.Scan(&cid, &sid); err != nil {
			log.Fatal(err)
		}
		schoolToClass[sid] = cid
		classIDs = append(classIDs, cid)
	}
	classRows.Close()

	csRows := make([][]any, 0, len(studentIDs))
	for _, sID := range studentIDs {
		csRows = append(csRows, []any{schoolToClass[studentToSchool[sID]], sID})
	}
	t = time.Now()
	cn, err := academics.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cn.Release()
	atx, err := cn.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := atx.Exec(ctx, `CREATE TEMP TABLE lt_class_students (class_id BIGINT, student_id BIGINT) ON COMMIT DROP`); err != nil {
		log.Fatal(err)
	}
	if _, err := atx.CopyFrom(ctx, pgx.Identifier{"lt_class_students"}, []string{"class_id", "student_id"}, pgx.CopyFromRows(csRows)); err != nil {
		log.Fatal(err)
	}
	if _, err := atx.Exec(ctx, `INSERT INTO class_students (class_id, student_id) SELECT class_id, student_id FROM lt_class_students`); err != nil {
		log.Fatal(err)
	}
	if err := atx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("class_students inserted in %s", time.Since(t).Round(time.Millisecond))

	forms := pool(ctx, cfg, "forms")
	defer forms.Close()

	formID := mustOne[int64](ctx, forms, `
        INSERT INTO forms (public_id, owner_id, title, description, status, published_at)
        VALUES ($1, 1, 'Load Test Assessment', 'Auto-seeded for k6', 'published', now())
        RETURNING id`, formPubID)
	mustExec(ctx, forms, `INSERT INTO questions (form_id, client_id, sort_order, title, type, required) VALUES ($1, 'q1', 10, 'Name', 'SHORT_ANSWER', true)`, formID)
	mustExec(ctx, forms, `INSERT INTO questions (form_id, client_id, sort_order, title, type, required) VALUES ($1, 'q2', 20, 'Confidence', 'LINEAR_SCALE', false)`, formID)
	formVerID := mustOne[int64](ctx, forms, `
        INSERT INTO form_versions (form_id, version_num, title, snapshot)
        VALUES ($1, 1, 'Load Test Assessment', '{"questions":[]}'::jsonb)
        RETURNING id`, formID)
	log.Printf("form=%d version=%d", formID, formVerID)

	assess := pool(ctx, cfg, "assessment")
	defer assess.Close()

	campaignID := mustOne[int64](ctx, assess, `
        INSERT INTO campaigns (public_id, title, scale_code, form_id, form_version_id, period_id, scope_node_id, status, opens_at)
        VALUES ($1, 'Load Test Campaign', 'french_5level', $2, $3, $4, $5, 'open', now())
        RETURNING id`, campaignPub, formID, formVerID, periodID, regionID)
	log.Printf("campaign: id=%d", campaignID)

	t = time.Now()
	mustExec(ctx, assess, `
        INSERT INTO assignments (campaign_id, student_id, access_token)
        SELECT $1, unnest($2::bigint[]), 'lt-' || md5(random()::text || generate_series(1, array_length($2::bigint[], 1)))`,
		campaignID, studentIDs)
	log.Printf("assignments inserted in %s", time.Since(t).Round(time.Millisecond))

	if !cfg.skipFixtures {
		writeFixtures(studentIDs, studentToSchool, schoolToClass, campaignID, regionID, periodID)
		log.Println("fixtures.csv written")
	}
}

func writeFixtures(studentIDs []int64, studentToSchool map[int64]int64, schoolToClass map[int64]int64, campaignID, regionID, periodID int64) {
	f, err := os.Create("tools/loadtest/seed/fixtures.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "student_id,class_id,school_id")
	for _, sID := range studentIDs {
		school := studentToSchool[sID]
		fmt.Fprintf(f, "%d,%d,%d\n", sID, schoolToClass[school], school)
	}
	meta, err := os.Create("tools/loadtest/seed/fixtures.meta.json")
	if err != nil {
		log.Fatal(err)
	}
	defer meta.Close()
	fmt.Fprintf(meta, `{"campaign_id":%d,"region_id":%d,"period_id":%d,"student_count":%d}`+"\n",
		campaignID, regionID, periodID, len(studentIDs))
}

func loadGuineaDistricts(path string) []guineaDistrict {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read guinea districts: %v", err)
	}
	var d []guineaDistrict
	if err := json.Unmarshal(b, &d); err != nil {
		log.Fatalf("parse guinea districts: %v", err)
	}
	return d
}

func totalGuineaSchools(dists []guineaDistrict) int {
	n := 0
	for _, d := range dists {
		n += d.Schools
	}
	return n
}

func seedGuineaHierarchy(ctx context.Context, conn *pgxpool.Pool, regionID int64, dists []guineaDistrict) {
	cn, err := conn.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer cn.Release()
	tx, err := cn.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE lt_dist_seed (idx INT PRIMARY KEY, name TEXT, schools INT) ON COMMIT DROP`); err != nil {
		log.Fatal(err)
	}
	rows := make([][]any, len(dists))
	for i, d := range dists {
		rows[i] = []any{i + 1, d.Name, d.Schools}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"lt_dist_seed"}, []string{"idx", "name", "schools"}, pgx.CopyFromRows(rows)); err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `
        INSERT INTO hierarchy_nodes (parent_id, level, code, label)
        SELECT $1, 'district', 'lt-d' || lpad(idx::text, 3, '0'), name
        FROM lt_dist_seed
        ORDER BY idx`, regionID); err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `
        WITH dist_nodes AS (
            SELECT id, code, row_number() OVER (ORDER BY code) AS n
            FROM hierarchy_nodes
            WHERE level = 'district' AND parent_id = $1 AND deleted_at IS NULL
        )
        INSERT INTO hierarchy_nodes (parent_id, level, code, label)
        SELECT
            d.id,
            'institution',
            'lt-s' || lpad((row_number() OVER (ORDER BY d.n, gs.s))::text, 6, '0'),
            'LT School ' || row_number() OVER (ORDER BY d.n, gs.s)
        FROM dist_nodes d
        JOIN lt_dist_seed s ON s.idx = d.n
        CROSS JOIN LATERAL generate_series(1, s.schools) AS gs(s)`, regionID); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	_ = strings.Builder{}
}

func mustPurge(ctx context.Context, cfg config) {
	steps := []struct{ db, sql string }{
		{"assessment", `DELETE FROM scores WHERE campaign_id IN (SELECT id FROM campaigns WHERE public_id = '` + campaignPub + `')`},
		{"assessment", `DELETE FROM responses WHERE campaign_id IN (SELECT id FROM campaigns WHERE public_id = '` + campaignPub + `')`},
		{"assessment", `DELETE FROM assignments WHERE campaign_id IN (SELECT id FROM campaigns WHERE public_id = '` + campaignPub + `')`},
		{"assessment", `DELETE FROM campaigns WHERE public_id = '` + campaignPub + `'`},
		{"forms", `DELETE FROM questions WHERE form_id IN (SELECT id FROM forms WHERE public_id = '` + formPubID + `')`},
		{"forms", `DELETE FROM form_versions WHERE form_id IN (SELECT id FROM forms WHERE public_id = '` + formPubID + `')`},
		{"forms", `DELETE FROM forms WHERE public_id = '` + formPubID + `'`},
		{"academics", `DELETE FROM class_students WHERE class_id IN (SELECT id FROM classes WHERE code = 'LT-CE1-A')`},
		{"academics", `DELETE FROM class_staff WHERE class_id IN (SELECT id FROM classes WHERE code = 'LT-CE1-A')`},
		{"academics", `DELETE FROM classes WHERE code = 'LT-CE1-A'`},
		{"academics", `DELETE FROM niveaux WHERE code = '` + niveauCode + `'`},
		{"academics", `DELETE FROM academic_periods WHERE code = '` + periodCode + `'`},
		{"people", `DELETE FROM students WHERE student_code LIKE 'LT-%'`},
		{"people", `DELETE FROM persons WHERE full_name LIKE 'LT Student %'`},
		{"tenancy", `DELETE FROM hierarchy_closure WHERE descendant_id IN (SELECT id FROM hierarchy_nodes WHERE code LIKE 'lt-%')`},
		{"tenancy", `DELETE FROM hierarchy_nodes WHERE code LIKE 'lt-%'`},
	}
	for _, s := range steps {
		p, err := pgxpool.New(ctx, dsn(cfg, s.db))
		if err != nil {
			log.Fatal(err)
		}
		if _, err := p.Exec(ctx, s.sql); err != nil {
			log.Fatalf("%s: %v\nsql: %s", s.db, err, s.sql)
		}
		p.Close()
	}
}
