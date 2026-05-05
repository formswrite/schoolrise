package academics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/academics"
)

type testFixtures struct {
	period *academics.Period
	niveau *academics.Niveau
}

func setupClassFixtures(t *testing.T) testFixtures {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")

	p, err := academics.CreatePeriod(ctx, academics.CreatePeriodParams{
		Code: "p-" + suffix, Label: "Period " + suffix,
		StartsOn: mustDate(t, "2050-09-01"),
		EndsOn:   mustDate(t, "2051-06-30"),
	})
	if err != nil {
		t.Fatalf("create period: %v", err)
	}

	n, err := academics.CreateNiveau(ctx, academics.CreateNiveauParams{
		Code: "n-" + suffix, Label: "Niveau " + suffix, SortOrder: 1,
	})
	if err != nil {
		t.Fatalf("create niveau: %v", err)
	}

	t.Cleanup(func() {
		_ = academics.DeleteNiveau(ctx, n.ID)
		_ = academics.DeletePeriod(ctx, p.ID)
	})

	return testFixtures{period: p, niveau: n}
}

func TestCreateClass_Success(t *testing.T) {
	ctx := context.Background()
	f := setupClassFixtures(t)

	c, err := academics.CreateClass(ctx, academics.CreateClassParams{
		PeriodID:      f.period.ID,
		NiveauID:      f.niveau.ID,
		InstitutionID: 1,
		Code:          "CE1-A",
		Label:         "CE1 Group A",
		Capacity:      30,
	})
	if err != nil {
		t.Fatalf("CreateClass: %v", err)
	}
	if c.ID == 0 || c.Capacity != 30 {
		t.Fatalf("bad class: %+v", c)
	}
	t.Cleanup(func() { _ = academics.DeleteClass(ctx, c.ID) })
}

func TestCreateClass_Validation(t *testing.T) {
	ctx := context.Background()
	f := setupClassFixtures(t)

	cases := []struct {
		name string
		p    academics.CreateClassParams
		want error
	}{
		{"missing code", academics.CreateClassParams{PeriodID: f.period.ID, NiveauID: f.niveau.ID, InstitutionID: 1, Label: "x"}, academics.ErrInvalidClassInput},
		{"missing label", academics.CreateClassParams{PeriodID: f.period.ID, NiveauID: f.niveau.ID, InstitutionID: 1, Code: "x"}, academics.ErrInvalidClassInput},
		{"missing institution", academics.CreateClassParams{PeriodID: f.period.ID, NiveauID: f.niveau.ID, Code: "x", Label: "y"}, academics.ErrInvalidClassInput},
		{"missing period", academics.CreateClassParams{NiveauID: f.niveau.ID, InstitutionID: 1, Code: "x", Label: "y"}, academics.ErrInvalidClassInput},
		{"missing niveau", academics.CreateClassParams{PeriodID: f.period.ID, InstitutionID: 1, Code: "x", Label: "y"}, academics.ErrInvalidClassInput},
		{"non-existent period", academics.CreateClassParams{PeriodID: 999999, NiveauID: f.niveau.ID, InstitutionID: 1, Code: "x", Label: "y"}, academics.ErrPeriodNotFound},
		{"non-existent niveau", academics.CreateClassParams{PeriodID: f.period.ID, NiveauID: 999999, InstitutionID: 1, Code: "x", Label: "y"}, academics.ErrNiveauNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := academics.CreateClass(ctx, tc.p)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCreateClass_DuplicateCodePerInstitutionPeriod(t *testing.T) {
	ctx := context.Background()
	f := setupClassFixtures(t)

	base := academics.CreateClassParams{
		PeriodID: f.period.ID, NiveauID: f.niveau.ID, InstitutionID: 42,
		Code: "CE1-DUP", Label: "Dup",
	}
	first, err := academics.CreateClass(ctx, base)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeleteClass(ctx, first.ID) })

	if _, err := academics.CreateClass(ctx, base); !errors.Is(err, academics.ErrClassCodeTaken) {
		t.Fatalf("err = %v, want ErrClassCodeTaken", err)
	}

	otherInst := base
	otherInst.InstitutionID = 99
	c, err := academics.CreateClass(ctx, otherInst)
	if err != nil {
		t.Fatalf("same code different institution: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeleteClass(ctx, c.ID) })
}

func TestListClassesByInstitution(t *testing.T) {
	ctx := context.Background()
	f := setupClassFixtures(t)

	mk := func(institutionID int64, code string) *academics.Class {
		c, err := academics.CreateClass(ctx, academics.CreateClassParams{
			PeriodID: f.period.ID, NiveauID: f.niveau.ID,
			InstitutionID: institutionID, Code: code, Label: code,
		})
		if err != nil {
			t.Fatalf("create %s@%d: %v", code, institutionID, err)
		}
		t.Cleanup(func() { _ = academics.DeleteClass(ctx, c.ID) })
		return c
	}

	a1 := mk(101, "A1")
	a2 := mk(101, "A2")
	_ = mk(202, "B1")

	got, err := academics.ListClassesByInstitution(ctx, 101)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	have := map[int64]bool{}
	for _, c := range got {
		if c.InstitutionID != 101 {
			t.Fatalf("class %d has institutionID=%d, want 101", c.ID, c.InstitutionID)
		}
		have[c.ID] = true
	}
	if !have[a1.ID] || !have[a2.ID] {
		t.Fatalf("missing expected classes: have=%v want={%d,%d}", have, a1.ID, a2.ID)
	}
}

func TestRoster_Students(t *testing.T) {
	ctx := context.Background()
	f := setupClassFixtures(t)
	c, err := academics.CreateClass(ctx, academics.CreateClassParams{
		PeriodID: f.period.ID, NiveauID: f.niveau.ID, InstitutionID: 1, Code: "RS-1", Label: "Roster Students",
	})
	if err != nil {
		t.Fatalf("create class: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeleteClass(ctx, c.ID) })

	if err := academics.AddStudentToClass(ctx, c.ID, 11); err != nil {
		t.Fatalf("add 11: %v", err)
	}
	if err := academics.AddStudentToClass(ctx, c.ID, 22); err != nil {
		t.Fatalf("add 22: %v", err)
	}
	if err := academics.AddStudentToClass(ctx, c.ID, 11); err != nil {
		t.Fatalf("add 11 again (idempotent): %v", err)
	}

	ids, err := academics.ListClassStudents(ctx, c.ID)
	if err != nil {
		t.Fatalf("list students: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("len = %d, want 2 (idempotency); got %v", len(ids), ids)
	}

	classes, err := academics.ListClassesForStudent(ctx, 22)
	if err != nil {
		t.Fatalf("classes for student: %v", err)
	}
	found := false
	for _, cl := range classes {
		if cl.ID == c.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("student 22 not linked to class %d", c.ID)
	}

	if err := academics.RemoveStudentFromClass(ctx, c.ID, 11); err != nil {
		t.Fatalf("remove 11: %v", err)
	}
	ids, _ = academics.ListClassStudents(ctx, c.ID)
	if len(ids) != 1 || ids[0] != 22 {
		t.Fatalf("after remove: %v", ids)
	}
}

func TestRoster_Staff(t *testing.T) {
	ctx := context.Background()
	f := setupClassFixtures(t)
	c, err := academics.CreateClass(ctx, academics.CreateClassParams{
		PeriodID: f.period.ID, NiveauID: f.niveau.ID, InstitutionID: 1, Code: "RST-1", Label: "Roster Staff",
	})
	if err != nil {
		t.Fatalf("create class: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeleteClass(ctx, c.ID) })

	if err := academics.AddStaffToClass(ctx, c.ID, 100, ""); err != nil {
		t.Fatalf("add teacher: %v", err)
	}
	if err := academics.AddStaffToClass(ctx, c.ID, 100, "assistant"); err != nil {
		t.Fatalf("add assistant role: %v", err)
	}
	if err := academics.AddStaffToClass(ctx, c.ID, 200, "teacher"); err != nil {
		t.Fatalf("add staff 200: %v", err)
	}

	staff, err := academics.ListClassStaff(ctx, c.ID)
	if err != nil {
		t.Fatalf("list staff: %v", err)
	}
	if len(staff) != 3 {
		t.Fatalf("len = %d, want 3 (same staff in 2 roles + another); got %+v", len(staff), staff)
	}

	classes, err := academics.ListClassesForStaff(ctx, 100)
	if err != nil {
		t.Fatalf("classes for staff: %v", err)
	}
	if len(classes) < 1 {
		t.Fatalf("expected staff 100 in at least 1 class")
	}

	if err := academics.RemoveStaffFromClass(ctx, c.ID, 100, "assistant"); err != nil {
		t.Fatalf("remove assistant: %v", err)
	}
	staff, _ = academics.ListClassStaff(ctx, c.ID)
	if len(staff) != 2 {
		t.Fatalf("after remove len = %d, want 2", len(staff))
	}
}

func TestDeleteClass_NotFound(t *testing.T) {
	ctx := context.Background()
	if err := academics.DeleteClass(ctx, 999999999); !errors.Is(err, academics.ErrClassNotFound) {
		t.Fatalf("err = %v, want ErrClassNotFound", err)
	}
}

func TestAddStudentToMissingClass_Errors(t *testing.T) {
	ctx := context.Background()
	if err := academics.AddStudentToClass(ctx, 999999999, 1); !errors.Is(err, academics.ErrClassNotFound) {
		t.Fatalf("err = %v, want ErrClassNotFound", err)
	}
}
