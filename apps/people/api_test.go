package people_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"

	"encore.app/apps/auth"
	"encore.app/apps/people"
	"encore.app/apps/tenancy"
)

var counter atomic.Int64

func uniqueLabel(t *testing.T) string {
	t.Helper()

	n := counter.Add(1)

	return fmt.Sprintf("%s-%d", strings.ReplaceAll(t.Name(), "/", "-"), n)
}

func uniqueEmail(t *testing.T) string {
	t.Helper()

	n := counter.Add(1)

	return fmt.Sprintf("people-%s-%d@local.test", strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")), n)
}

func newAPIService(t *testing.T) *people.Service {
	t.Helper()
	return &people.Service{}
}

func seedAdmin(t *testing.T) *auth.User {
	t.Helper()

	user, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: uniqueEmail(t), Password: "Pass1234!", FullName: "People Tester", Role: "admin",
	})
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	if err := auth.AssignRole(context.Background(), user.ID, "admin", nil); err != nil {
		t.Fatalf("AssignRole admin: %v", err)
	}

	return user
}

func authedAdminCtx(t *testing.T) context.Context {
	t.Helper()

	user := seedAdmin(t)
	et.OverrideAuthInfo(encauth.UID(strconv.FormatInt(user.ID, 10)), &auth.AuthData{UserID: user.ID, Role: "admin"})

	return context.Background()
}

func seedInstitution(t *testing.T, ctx context.Context) int64 {
	t.Helper()

	if err := tenancy.ReloadLevels(context.Background()); err != nil {
		t.Fatalf("ReloadLevels: %v", err)
	}

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueLabel(t) + "-r", Label: "R",
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}

	pref, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID, Level: tenancy.LevelPrefecture, Code: uniqueLabel(t) + "-p", Label: "P",
	})
	if err != nil {
		t.Fatalf("prefecture: %v", err)
	}

	deleg, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &pref.ID, Level: tenancy.LevelDelegation, Code: uniqueLabel(t) + "-d", Label: "D",
	})
	if err != nil {
		t.Fatalf("delegation: %v", err)
	}

	inst, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &deleg.ID, Level: tenancy.LevelInstitution, Code: uniqueLabel(t) + "-i", Label: "School",
	})
	if err != nil {
		t.Fatalf("institution: %v", err)
	}

	return inst.ID
}

func TestCreateStudent_RequiresInstitutionAccess(t *testing.T) {

	ctx := authedAdminCtx(t)

	institutionID := seedInstitution(t, ctx)

	resp, err := newAPIService(t).CreateStudentAPI(ctx, &people.CreateStudentAPIRequest{
		InstitutionID: institutionID,
		StudentCode:   uniqueLabel(t),
		FullName:      "Sarah Test",
	})
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	if resp.Student.ID == 0 {
		t.Error("ID not assigned")
	}

	if resp.Student.Person.FullName != "Sarah Test" {
		t.Errorf("FullName mismatch")
	}
}

func TestCreateStudent_RejectsOutsideScope(t *testing.T) {

	adminCtx := authedAdminCtx(t)
	institutionA := seedInstitution(t, adminCtx)
	institutionB := seedInstitution(t, adminCtx)

	scopedUser, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: uniqueEmail(t), Password: "Pass1234!", FullName: "Scoped", Role: "inspector",
	})
	if err != nil {
		t.Fatalf("scoped user: %v", err)
	}

	if err := auth.AssignRole(context.Background(), scopedUser.ID, "inspector", &institutionA); err != nil {
		t.Fatalf("AssignRole scoped: %v", err)
	}

	et.OverrideAuthInfo(encauth.UID(strconv.FormatInt(scopedUser.ID, 10)), &auth.AuthData{UserID: scopedUser.ID, Role: "inspector"})

	_, err = newAPIService(t).CreateStudentAPI(context.Background(), &people.CreateStudentAPIRequest{
		InstitutionID: institutionB,
		StudentCode:   uniqueLabel(t),
		FullName:      "Out of Scope",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
}

func TestListStudents_FilteredByInstitution(t *testing.T) {

	ctx := authedAdminCtx(t)
	institutionID := seedInstitution(t, ctx)

	_, err := newAPIService(t).CreateStudentAPI(ctx, &people.CreateStudentAPIRequest{
		InstitutionID: institutionID,
		StudentCode:   uniqueLabel(t),
		FullName:      "Listed Student",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	resp, err := newAPIService(t).ListStudentsAPI(ctx, &people.ListStudentsRequest{
		InstitutionID: institutionID,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(resp.Students) == 0 {
		t.Fatal("expected at least one student")
	}

	found := false
	for _, st := range resp.Students {
		if st.Person.FullName == "Listed Student" {
			found = true
			break
		}
	}

	if !found {
		t.Error("created student not in listing")
	}
}

func TestGetStudent_ReturnsAndDeletes(t *testing.T) {

	ctx := authedAdminCtx(t)
	institutionID := seedInstitution(t, ctx)

	created, err := newAPIService(t).CreateStudentAPI(ctx, &people.CreateStudentAPIRequest{
		InstitutionID: institutionID,
		StudentCode:   uniqueLabel(t),
		FullName:      "Round Trip Student",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	fetched, err := newAPIService(t).GetStudentAPI(ctx, created.Student.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if fetched.Student.Person.FullName != "Round Trip Student" {
		t.Errorf("FullName mismatch")
	}

	if err := newAPIService(t).DeleteStudentAPI(ctx, created.Student.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = newAPIService(t).GetStudentAPI(ctx, created.Student.ID)
	if err == nil {
		t.Error("expected NotFound after delete")
	}
}

func TestCreateStaff_RoundTrip(t *testing.T) {

	ctx := authedAdminCtx(t)
	institutionID := seedInstitution(t, ctx)

	created, err := newAPIService(t).CreateStaffAPI(ctx, &people.CreateStaffAPIRequest{
		ScopeNodeID: institutionID,
		Position:    "Headmaster",
		FullName:    "Mr. Khan",
	})
	if err != nil {
		t.Fatalf("CreateStaff: %v", err)
	}

	if created.Staff.Position != "Headmaster" {
		t.Errorf("Position mismatch")
	}

	resp, err := newAPIService(t).ListStaffAPI(ctx, &people.ListStaffRequest{ScopeNodeID: institutionID})
	if err != nil {
		t.Fatalf("ListStaff: %v", err)
	}

	if len(resp.Staff) == 0 {
		t.Error("expected at least one staff member")
	}
}

func TestListStudents_RejectsMissingInstitution(t *testing.T) {

	ctx := authedAdminCtx(t)

	_, err := newAPIService(t).ListStudentsAPI(ctx, &people.ListStudentsRequest{InstitutionID: 0})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}
