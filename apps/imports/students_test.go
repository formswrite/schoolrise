package imports_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"encore.app/apps/imports"
)

func TestImportStudents_DryRunReportsButDoesNotInsert(t *testing.T) {
	ctx := context.Background()
	institutionID := time.Now().UnixNano()

	csv := strings.Join([]string{
		"full_name,gender,date_of_birth,student_code",
		"Alice Test," + uniq() + ",2010-04-12,S-A1",
		"Bob Test,male,2009-11-30,S-B1",
		"Cleo Test,female,2010-02-28,S-C1",
	}, "\n")

	job, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: institutionID,
		CSVData:       csv,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if job.Status != imports.StatusCompleted {
		t.Fatalf("status = %q, want completed", job.Status)
	}
	if job.TotalRows != 3 || job.Succeeded != 3 || job.Failed != 0 {
		t.Fatalf("counts wrong: total=%d succ=%d fail=%d", job.TotalRows, job.Succeeded, job.Failed)
	}
	if !job.DryRun {
		t.Fatalf("dry_run not recorded")
	}
}

func TestImportStudents_PartialSuccessRecordsRowErrors(t *testing.T) {
	ctx := context.Background()
	institutionID := time.Now().UnixNano() + 1

	csv := strings.Join([]string{
		"full_name,gender,date_of_birth",
		"Good Student,female,2010-03-15",
		",male,2009-01-01",
		"Bad Date,female,not-a-date",
		"Another Good,male,2008-08-08",
	}, "\n")

	job, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: institutionID,
		CSVData:       csv,
		DryRun:        false,
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if job.TotalRows != 4 {
		t.Fatalf("total=%d, want 4", job.TotalRows)
	}
	if job.Succeeded != 2 || job.Failed != 2 {
		t.Fatalf("succ=%d fail=%d, want 2/2", job.Succeeded, job.Failed)
	}
	if len(job.Errors) != 2 {
		t.Fatalf("errors len = %d, want 2", len(job.Errors))
	}

	rowNums := map[int]bool{}
	for _, e := range job.Errors {
		rowNums[e.RowNumber] = true
	}
	if !rowNums[3] || !rowNums[4] {
		t.Fatalf("expected errors on rows 3 and 4 (1-indexed +1 for header), got %v", rowNums)
	}
}

func TestImportStudents_RejectsMissingHeader(t *testing.T) {
	ctx := context.Background()
	csv := "name,gender\nAlice,female"
	job, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: 1, CSVData: csv,
	})
	if err != nil {
		t.Fatalf("expected job (with header error recorded), got go-err: %v", err)
	}
	if job.Status != imports.StatusFailed {
		t.Fatalf("status = %q, want failed", job.Status)
	}
	if len(job.Errors) == 0 {
		t.Fatalf("expected at least one row error")
	}
}

func TestImportStudents_EmptyCSVRejected(t *testing.T) {
	ctx := context.Background()
	_, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: 1, CSVData: "   ",
	})
	if !errors.Is(err, imports.ErrEmptyCSV) {
		t.Fatalf("err = %v, want ErrEmptyCSV", err)
	}
}

func TestImportStudents_ZeroInstitutionRejected(t *testing.T) {
	ctx := context.Background()
	_, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: 0, CSVData: "full_name\nAlice",
	})
	if !errors.Is(err, imports.ErrInvalidImportInput) {
		t.Fatalf("err = %v, want ErrInvalidImportInput", err)
	}
}

func TestImportStudents_RealRunInsertsThroughPeople(t *testing.T) {
	ctx := context.Background()
	institutionID := time.Now().UnixNano() + 2
	suffix := time.Now().Format("150405.000")

	csv := strings.Join([]string{
		"full_name,gender,student_code",
		"Real Student One " + suffix + ",female,RS-1-" + suffix,
		"Real Student Two " + suffix + ",male,RS-2-" + suffix,
	}, "\n")

	job, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: institutionID, CSVData: csv, DryRun: false,
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if job.Succeeded != 2 || job.Failed != 0 {
		t.Fatalf("expected 2/0, got %d/%d (errors: %+v)", job.Succeeded, job.Failed, job.Errors)
	}
}

func TestImportStudents_LargeBatch(t *testing.T) {
	ctx := context.Background()
	institutionID := time.Now().UnixNano() + 3
	suffix := time.Now().Format("150405.000")

	const N = 250
	var b strings.Builder
	b.WriteString("full_name,gender,student_code\n")
	for i := 0; i < N; i++ {
		b.WriteString(fmt.Sprintf("Student%05d %s,female,SC-%05d-%s\n", i, suffix, i, suffix))
	}
	job, err := imports.ImportStudents(ctx, imports.ImportStudentsParams{
		InstitutionID: institutionID, CSVData: b.String(), DryRun: false,
	})
	if err != nil {
		t.Fatalf("large batch: %v", err)
	}
	if job.Succeeded != N || job.Failed != 0 {
		t.Fatalf("succ=%d fail=%d, want %d/0", job.Succeeded, job.Failed, N)
	}
}

var uniqCounter int64

func uniq() string {
	uniqCounter++
	return fmt.Sprintf("female-%d", time.Now().UnixNano()+uniqCounter)
}
