package imports_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"encore.app/apps/imports"
	"encore.app/apps/tenancy"
)

func TestImportStaff_DryRunCounts(t *testing.T) {
	ctx := context.Background()
	scopeID := time.Now().UnixNano() + 5000

	csv := strings.Join([]string{
		"full_name,position,gender,hire_date,staff_code",
		"Alice Teacher,Teacher,female,2020-09-01,T-A-1",
		"Bob Inspector,Inspector,male,2018-01-10,T-B-1",
	}, "\n")

	job, err := imports.ImportStaff(ctx, imports.ImportStaffParams{
		ScopeNodeID: scopeID, CSVData: csv, DryRun: true,
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if job.Succeeded != 2 || job.Failed != 0 {
		t.Fatalf("succ=%d fail=%d", job.Succeeded, job.Failed)
	}
}

func TestImportStaff_BadHireDateFails(t *testing.T) {
	ctx := context.Background()
	csv := strings.Join([]string{
		"full_name,hire_date",
		"Alice Teacher,2020-09-01",
		"Bob Bad,not-a-date",
	}, "\n")
	job, err := imports.ImportStaff(ctx, imports.ImportStaffParams{
		ScopeNodeID: time.Now().UnixNano() + 5001, CSVData: csv, DryRun: true,
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if job.Succeeded != 1 || job.Failed != 1 {
		t.Fatalf("succ=%d fail=%d, want 1/1", job.Succeeded, job.Failed)
	}
}

func TestImportSchools_HappyPath(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "imp-r-" + suffix, Label: "Imp Region", Level: tenancy.LevelRegion,
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}
	pref, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "imp-p-" + suffix, Label: "Imp Pref", Level: tenancy.LevelPrefecture, ParentID: &region.ID,
	})
	if err != nil {
		t.Fatalf("pref: %v", err)
	}
	deleg, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "imp-d-" + suffix, Label: "Imp Deleg", Level: tenancy.LevelDelegation, ParentID: &pref.ID,
	})
	if err != nil {
		t.Fatalf("deleg: %v", err)
	}
	t.Cleanup(func() {
		_ = tenancy.SoftDeleteNode(ctx, deleg.ID)
		_ = tenancy.SoftDeleteNode(ctx, pref.ID)
		_ = tenancy.SoftDeleteNode(ctx, region.ID)
	})

	csv := strings.Join([]string{
		"code,label,parent_id",
		"sch-a-" + suffix + "," + "School A," + parentStr(deleg.ID),
		"sch-b-" + suffix + "," + "School B," + parentStr(deleg.ID),
		"sch-bad-" + suffix + ",Missing Parent," + "999999999",
	}, "\n")

	job, err := imports.ImportSchools(ctx, imports.ImportSchoolsParams{
		CSVData: csv, DryRun: false,
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if job.Succeeded != 2 {
		t.Fatalf("succ=%d, want 2", job.Succeeded)
	}
	if job.Failed != 1 {
		t.Fatalf("fail=%d, want 1", job.Failed)
	}
}

func TestImportSchools_RejectsEmpty(t *testing.T) {
	if _, err := imports.ImportSchools(context.Background(), imports.ImportSchoolsParams{}); !errors.Is(err, imports.ErrEmptyCSV) {
		t.Fatalf("err=%v, want ErrEmptyCSV", err)
	}
}

func parentStr(id int64) string {
	return strings.TrimSpace((func() string {
		var b strings.Builder
		_, _ = b.WriteString("")
		fmtInt(&b, id)
		return b.String()
	})())
}

func fmtInt(b *strings.Builder, n int64) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	if n < 0 {
		b.WriteByte('-')
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	b.Write(digits)
}
