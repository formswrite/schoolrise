package enrollment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/enrollment"
	"encore.app/apps/people"
	"encore.app/apps/tenancy"
)

func TestComputeCoverage_Validation(t *testing.T) {
	ctx := context.Background()
	if _, err := enrollment.ComputeCoverage(ctx, 0, 1); !errors.Is(err, enrollment.ErrInvalidCoverageInput) {
		t.Fatalf("scope=0 err=%v", err)
	}
	if _, err := enrollment.ComputeCoverage(ctx, 1, 0); !errors.Is(err, enrollment.ErrInvalidCoverageInput) {
		t.Fatalf("period=0 err=%v", err)
	}
}

type covFixture struct {
	region    *tenancy.Node
	schoolA   *tenancy.Node
	schoolB   *tenancy.Node
	periodID  int64
	studentsA []*people.Student
	studentB  *people.Student
}

func setupCoverageFixture(t *testing.T) *covFixture {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "cov-r-" + suffix, Label: "Cov Region", Level: tenancy.LevelRegion,
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}
	prefecture, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "cov-p-" + suffix, Label: "Cov Prefecture", Level: tenancy.LevelPrefecture, ParentID: &region.ID,
	})
	if err != nil {
		t.Fatalf("prefecture: %v", err)
	}
	delegation, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "cov-d-" + suffix, Label: "Cov Delegation", Level: tenancy.LevelDelegation, ParentID: &prefecture.ID,
	})
	if err != nil {
		t.Fatalf("delegation: %v", err)
	}
	schoolA, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "cov-a-" + suffix, Label: "School A", Level: tenancy.LevelInstitution, ParentID: &delegation.ID,
	})
	if err != nil {
		t.Fatalf("schoolA: %v", err)
	}
	schoolB, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "cov-b-" + suffix, Label: "School B", Level: tenancy.LevelInstitution, ParentID: &delegation.ID,
	})
	if err != nil {
		t.Fatalf("schoolB: %v", err)
	}
	t.Cleanup(func() {
		_ = tenancy.SoftDeleteNode(ctx, schoolA.ID)
		_ = tenancy.SoftDeleteNode(ctx, schoolB.ID)
		_ = tenancy.SoftDeleteNode(ctx, delegation.ID)
		_ = tenancy.SoftDeleteNode(ctx, prefecture.ID)
		_ = tenancy.SoftDeleteNode(ctx, region.ID)
	})

	periodID := time.Now().UnixNano()

	mkStudent := func(institutionID int64, name, gender string) *people.Student {
		st, _, err := people.CreateStudent(ctx, people.CreateStudentParams{
			InstitutionID: institutionID,
			Person: people.CreatePersonParams{
				FullName: name + "-" + suffix,
				Gender:   gender,
			},
		})
		if err != nil {
			t.Fatalf("student %s: %v", name, err)
		}
		t.Cleanup(func() { _ = people.SoftDeleteStudent(ctx, st.ID) })
		return st
	}

	a1 := mkStudent(schoolA.ID, "Alice", "female")
	a2 := mkStudent(schoolA.ID, "Bob", "male")
	a3 := mkStudent(schoolA.ID, "Cleo", "female")
	a4 := mkStudent(schoolA.ID, "Dakota", "")
	b1 := mkStudent(schoolB.ID, "Eli", "male")

	for _, s := range []*people.Student{a1, a2, a3, a4} {
		if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
			StudentID: s.ID, InstitutionID: schoolA.ID, PeriodID: periodID,
			EnrolledOn: mustDate(t, "2025-09-01"),
		}); err != nil {
			t.Fatalf("enroll A %d: %v", s.ID, err)
		}
	}
	if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: b1.ID, InstitutionID: schoolB.ID, PeriodID: periodID,
		EnrolledOn: mustDate(t, "2025-09-01"),
	}); err != nil {
		t.Fatalf("enroll B: %v", err)
	}

	return &covFixture{
		region: region, schoolA: schoolA, schoolB: schoolB,
		periodID:  periodID,
		studentsA: []*people.Student{a1, a2, a3, a4},
		studentB:  b1,
	}
}

func TestComputeCoverage_PerInstitution(t *testing.T) {
	ctx := context.Background()
	f := setupCoverageFixture(t)

	covA, err := enrollment.ComputeCoverage(ctx, f.schoolA.ID, f.periodID)
	if err != nil {
		t.Fatalf("schoolA coverage: %v", err)
	}
	if covA.TotalEnrolled != 4 {
		t.Fatalf("schoolA total = %d, want 4", covA.TotalEnrolled)
	}
	if covA.Female != 2 {
		t.Fatalf("schoolA female = %d, want 2", covA.Female)
	}
	if covA.Male != 1 {
		t.Fatalf("schoolA male = %d, want 1", covA.Male)
	}
	if covA.Unknown != 1 {
		t.Fatalf("schoolA unknown = %d, want 1", covA.Unknown)
	}
}

func TestComputeCoverage_RollsUpAcrossDescendants(t *testing.T) {
	ctx := context.Background()
	f := setupCoverageFixture(t)

	covRegion, err := enrollment.ComputeCoverage(ctx, f.region.ID, f.periodID)
	if err != nil {
		t.Fatalf("region coverage: %v", err)
	}
	if covRegion.TotalEnrolled != 5 {
		t.Fatalf("region total = %d, want 5 (4 in A + 1 in B)", covRegion.TotalEnrolled)
	}
	if covRegion.Male != 2 {
		t.Fatalf("region male = %d, want 2", covRegion.Male)
	}
	if covRegion.Female != 2 {
		t.Fatalf("region female = %d, want 2", covRegion.Female)
	}
	if covRegion.Unknown != 1 {
		t.Fatalf("region unknown = %d, want 1", covRegion.Unknown)
	}
}

func TestComputeCoverage_ExcludesDroppedAndTransferred(t *testing.T) {
	ctx := context.Background()
	f := setupCoverageFixture(t)

	dropEnr, err := enrollment.GetActiveEnrollment(ctx, f.studentsA[0].ID, f.periodID)
	if err != nil {
		t.Fatalf("get active: %v", err)
	}
	if _, err := enrollment.DropEnrollment(ctx, enrollment.DropParams{
		EnrollmentID: dropEnr.ID, EndedOn: mustDate(t, "2025-10-01"),
	}); err != nil {
		t.Fatalf("drop: %v", err)
	}

	if _, err := enrollment.TransferEnrollment(ctx, enrollment.TransferParams{
		StudentID: f.studentsA[1].ID, PeriodID: f.periodID, ToInstitutionID: f.schoolB.ID,
		EffectiveOn: mustDate(t, "2025-11-01"),
	}); err != nil {
		t.Fatalf("transfer: %v", err)
	}

	covA, err := enrollment.ComputeCoverage(ctx, f.schoolA.ID, f.periodID)
	if err != nil {
		t.Fatalf("schoolA coverage: %v", err)
	}
	if covA.TotalEnrolled != 2 {
		t.Fatalf("schoolA after drop+transfer total = %d, want 2", covA.TotalEnrolled)
	}

	covB, err := enrollment.ComputeCoverage(ctx, f.schoolB.ID, f.periodID)
	if err != nil {
		t.Fatalf("schoolB coverage: %v", err)
	}
	if covB.TotalEnrolled != 2 {
		t.Fatalf("schoolB after transfer = %d, want 2 (original + transferred-in)", covB.TotalEnrolled)
	}

	covRegion, err := enrollment.ComputeCoverage(ctx, f.region.ID, f.periodID)
	if err != nil {
		t.Fatalf("region coverage: %v", err)
	}
	if covRegion.TotalEnrolled != 4 {
		t.Fatalf("region after drop = %d, want 4 (5 created - 1 dropped)", covRegion.TotalEnrolled)
	}
}

func TestSnapshotCoverage_PersistsRow(t *testing.T) {
	ctx := context.Background()
	f := setupCoverageFixture(t)

	cov, err := enrollment.SnapshotCoverage(ctx, f.region.ID, f.periodID, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if cov.TotalEnrolled != 5 {
		t.Fatalf("snapshot total = %d, want 5", cov.TotalEnrolled)
	}
}
