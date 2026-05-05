package authz_test

import (
	"context"
	"testing"

	"encore.app/pkg/authz"
)

type fakeBackend struct {
	assignments map[int64][]authz.Assignment
	descendants map[int64]map[int64]bool
}

func (f *fakeBackend) load(_ context.Context, userID int64) ([]authz.Assignment, error) {
	return f.assignments[userID], nil
}

func (f *fakeBackend) isAncestor(_ context.Context, ancestor, descendant int64) (bool, error) {
	if ancestor == descendant {
		return true, nil
	}

	return f.descendants[ancestor][descendant], nil
}

func newEnforcer(t *testing.T, b *fakeBackend) *authz.Enforcer {
	t.Helper()

	e, err := authz.New(b.load, b.isAncestor)
	if err != nil {
		t.Fatalf("authz.New: %v", err)
	}

	return e
}

func ptr(v int64) *int64 { return &v }

func TestCanAccessNode_GlobalAdminAccessesAnyNode(t *testing.T) {
	t.Parallel()

	b := &fakeBackend{
		assignments: map[int64][]authz.Assignment{
			1: {{UserID: 1, Role: "admin", ScopeNodeID: nil}},
		},
	}

	ok, err := newEnforcer(t, b).CanAccessNode(context.Background(), 1, 999)
	if err != nil {
		t.Fatalf("CanAccessNode: %v", err)
	}

	if !ok {
		t.Error("global admin should access any node")
	}
}

func TestCanAccessNode_ScopedUserAccessesDescendant(t *testing.T) {
	t.Parallel()

	b := &fakeBackend{
		assignments: map[int64][]authz.Assignment{
			2: {{UserID: 2, Role: "inspector", ScopeNodeID: ptr(10)}},
		},
		descendants: map[int64]map[int64]bool{
			10: {20: true, 30: true},
		},
	}

	ok, err := newEnforcer(t, b).CanAccessNode(context.Background(), 2, 20)
	if err != nil {
		t.Fatalf("CanAccessNode: %v", err)
	}

	if !ok {
		t.Error("inspector at scope 10 should access descendant 20")
	}
}

func TestCanAccessNode_ScopedUserAccessesOwnScope(t *testing.T) {
	t.Parallel()

	b := &fakeBackend{
		assignments: map[int64][]authz.Assignment{
			2: {{UserID: 2, Role: "inspector", ScopeNodeID: ptr(10)}},
		},
	}

	ok, err := newEnforcer(t, b).CanAccessNode(context.Background(), 2, 10)
	if err != nil {
		t.Fatalf("CanAccessNode: %v", err)
	}

	if !ok {
		t.Error("inspector should access their own scope node")
	}
}

func TestCanAccessNode_ScopedUserDeniedOutsideScope(t *testing.T) {
	t.Parallel()

	b := &fakeBackend{
		assignments: map[int64][]authz.Assignment{
			2: {{UserID: 2, Role: "inspector", ScopeNodeID: ptr(10)}},
		},
		descendants: map[int64]map[int64]bool{
			10: {20: true},
		},
	}

	ok, err := newEnforcer(t, b).CanAccessNode(context.Background(), 2, 99)
	if err != nil {
		t.Fatalf("CanAccessNode: %v", err)
	}

	if ok {
		t.Error("inspector at scope 10 should NOT access node 99")
	}
}

func TestCanAccessNode_UserWithNoAssignmentsDenied(t *testing.T) {
	t.Parallel()

	b := &fakeBackend{
		assignments: map[int64][]authz.Assignment{},
	}

	ok, err := newEnforcer(t, b).CanAccessNode(context.Background(), 99, 1)
	if err != nil {
		t.Fatalf("CanAccessNode: %v", err)
	}

	if ok {
		t.Error("user without assignments should be denied")
	}
}

func TestCanAccessNode_ZeroUserDenied(t *testing.T) {
	t.Parallel()

	b := &fakeBackend{}

	ok, err := newEnforcer(t, b).CanAccessNode(context.Background(), 0, 1)
	if err != nil {
		t.Fatalf("CanAccessNode: %v", err)
	}

	if ok {
		t.Error("zero user id should be denied")
	}
}
