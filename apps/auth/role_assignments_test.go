package auth_test

import (
	"context"
	"errors"
	"testing"

	"encore.app/apps/auth"
)

func TestAssignRole_GlobalAdmin(t *testing.T) {

	user := newTestUser(t)

	if err := auth.AssignRole(context.Background(), user.ID, "admin", nil); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	assignments, err := auth.ListRoleAssignmentsForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ListRoleAssignmentsForUser: %v", err)
	}

	if len(assignments) != 1 {
		t.Fatalf("got %d assignments, want 1", len(assignments))
	}

	if assignments[0].Role != "admin" {
		t.Errorf("Role = %q, want admin", assignments[0].Role)
	}

	if assignments[0].ScopeNodeID != nil {
		t.Errorf("ScopeNodeID should be nil for global admin, got %v", *assignments[0].ScopeNodeID)
	}
}

func TestAssignRole_ScopedInspector(t *testing.T) {

	user := newTestUser(t)
	scope := int64(42)

	if err := auth.AssignRole(context.Background(), user.ID, "inspector", &scope); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	assignments, err := auth.ListRoleAssignmentsForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ListRoleAssignmentsForUser: %v", err)
	}

	if len(assignments) != 1 {
		t.Fatalf("got %d assignments, want 1", len(assignments))
	}

	if assignments[0].ScopeNodeID == nil || *assignments[0].ScopeNodeID != 42 {
		t.Errorf("ScopeNodeID = %v, want 42", assignments[0].ScopeNodeID)
	}
}

func TestAssignRole_Idempotent(t *testing.T) {

	user := newTestUser(t)

	for i := 0; i < 3; i++ {
		if err := auth.AssignRole(context.Background(), user.ID, "admin", nil); err != nil {
			t.Fatalf("AssignRole iteration %d: %v", i, err)
		}
	}

	assignments, err := auth.ListRoleAssignmentsForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ListRoleAssignmentsForUser: %v", err)
	}

	if len(assignments) != 1 {
		t.Errorf("expected 1 deduplicated assignment, got %d", len(assignments))
	}
}

func TestAssignRole_RejectsEmptyRole(t *testing.T) {

	user := newTestUser(t)

	err := auth.AssignRole(context.Background(), user.ID, "", nil)
	if !errors.Is(err, auth.ErrInvalidRoleAssignment) {
		t.Errorf("expected ErrInvalidRoleAssignment, got %v", err)
	}
}

func TestRevokeRole_RemovesAssignment(t *testing.T) {

	user := newTestUser(t)

	if err := auth.AssignRole(context.Background(), user.ID, "admin", nil); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	assignments, err := auth.ListRoleAssignmentsForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("list before: %v", err)
	}

	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment before revoke, got %d", len(assignments))
	}

	if err := auth.RevokeRole(context.Background(), assignments[0].ID); err != nil {
		t.Fatalf("RevokeRole: %v", err)
	}

	after, err := auth.ListRoleAssignmentsForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("list after: %v", err)
	}

	if len(after) != 0 {
		t.Errorf("expected 0 assignments after revoke, got %d", len(after))
	}
}
