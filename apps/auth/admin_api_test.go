package auth_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"

	"encore.app/apps/auth"
)

func authedAdminCtx(t *testing.T) (context.Context, *auth.User) {
	t.Helper()

	user := newTestUser(t)

	if err := auth.AssignRole(context.Background(), user.ID, "admin", nil); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	et.OverrideAuthInfo(encauth.UID(strconv.FormatInt(user.ID, 10)), &auth.AuthData{UserID: user.ID, Role: "admin"})

	return context.Background(), user
}

func authedNonAdminCtx(t *testing.T) (context.Context, *auth.User) {
	t.Helper()

	user := newTestUser(t)

	et.OverrideAuthInfo(encauth.UID(strconv.FormatInt(user.ID, 10)), &auth.AuthData{UserID: user.ID, Role: "teacher"})

	return context.Background(), user
}

func newAuthService(t *testing.T) *auth.Service {
	t.Helper()
	return &auth.Service{}
}

func TestListUsers_RequiresGlobalAdmin(t *testing.T) {

	ctx, _ := authedNonAdminCtx(t)

	_, err := newAuthService(t).ListUsersAPI(ctx)

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
}

func TestListUsers_ReturnsUsers(t *testing.T) {

	ctx, admin := authedAdminCtx(t)

	resp, err := newAuthService(t).ListUsersAPI(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}

	found := false
	for _, u := range resp.Users {
		if u.ID == admin.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("created admin not in list")
	}
}

func TestCreateUser_RequiresGlobalAdmin(t *testing.T) {

	ctx, _ := authedNonAdminCtx(t)

	_, err := newAuthService(t).CreateUserAPI(ctx, &auth.CreateUserAPIRequest{
		Email: uniqueEmail(t), FullName: "X", Password: "Pass1234!", Role: "teacher",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
}

func TestCreateUser_AdminCreatesNewUser(t *testing.T) {

	ctx, _ := authedAdminCtx(t)

	resp, err := newAuthService(t).CreateUserAPI(ctx, &auth.CreateUserAPIRequest{
		Email:    uniqueEmail(t),
		FullName: "Created User",
		Password: "NewUserPass1234",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if resp.User.ID == 0 {
		t.Error("user ID not assigned")
	}

	if resp.User.Role != "teacher" {
		t.Errorf("role = %q, want teacher", resp.User.Role)
	}
}

func TestCreateAssignment_RequiresGlobalAdmin(t *testing.T) {

	ctx, target := authedNonAdminCtx(t)

	err := newAuthService(t).CreateAssignmentAPI(ctx, &auth.CreateAssignmentRequest{
		UserID: target.ID, Role: "inspector",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
}

func TestCreateAssignment_AdminAssignsScopedRole(t *testing.T) {

	ctx, _ := authedAdminCtx(t)
	target := newTestUser(t)
	scope := int64(42)

	err := newAuthService(t).CreateAssignmentAPI(ctx, &auth.CreateAssignmentRequest{
		UserID: target.ID, Role: "inspector", ScopeNodeID: &scope,
	})
	if err != nil {
		t.Fatalf("CreateAssignment: %v", err)
	}

	resp, err := newAuthService(t).ListUserAssignmentsAPI(ctx, target.ID)
	if err != nil {
		t.Fatalf("ListUserAssignments: %v", err)
	}

	if len(resp.Assignments) != 1 {
		t.Fatalf("got %d assignments, want 1", len(resp.Assignments))
	}

	if resp.Assignments[0].Role != "inspector" {
		t.Errorf("role = %q, want inspector", resp.Assignments[0].Role)
	}

	if resp.Assignments[0].ScopeNodeID == nil || *resp.Assignments[0].ScopeNodeID != 42 {
		t.Errorf("scope = %v, want 42", resp.Assignments[0].ScopeNodeID)
	}
}

func TestDeleteAssignment_AdminRevokes(t *testing.T) {

	ctx, _ := authedAdminCtx(t)
	target := newTestUser(t)

	if err := newAuthService(t).CreateAssignmentAPI(ctx, &auth.CreateAssignmentRequest{
		UserID: target.ID, Role: "inspector",
	}); err != nil {
		t.Fatalf("CreateAssignment: %v", err)
	}

	resp, _ := newAuthService(t).ListUserAssignmentsAPI(ctx, target.ID)

	if len(resp.Assignments) != 1 {
		t.Fatalf("setup: expected 1 assignment, got %d", len(resp.Assignments))
	}

	if err := newAuthService(t).DeleteAssignmentAPI(ctx, resp.Assignments[0].ID); err != nil {
		t.Fatalf("DeleteAssignment: %v", err)
	}

	after, _ := newAuthService(t).ListUserAssignmentsAPI(ctx, target.ID)
	if len(after.Assignments) != 0 {
		t.Errorf("expected 0 assignments after revoke, got %d", len(after.Assignments))
	}
}

func TestGetUser_RequiresGlobalAdmin(t *testing.T) {

	ctx, _ := authedNonAdminCtx(t)

	_, err := newAuthService(t).GetUserAPI(ctx, 1)

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
}
