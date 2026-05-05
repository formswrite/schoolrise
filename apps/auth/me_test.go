package auth_test

import (
	"context"
	"testing"

	"encore.app/apps/auth"
)

func TestMe_ReturnsCurrentUser(t *testing.T) {

	user := newTestUser(t)

	_, session, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	setAuthInfo(t, user, session.ID)

	resp, err := loginService(t).MeAPI(context.Background())
	if err != nil {
		t.Fatalf("Me: %v", err)
	}

	if resp.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", resp.UserID, user.ID)
	}

	if resp.Email != user.Email {
		t.Errorf("Email = %q, want %q", resp.Email, user.Email)
	}

	if resp.Role != user.Role {
		t.Errorf("Role = %q, want %q", resp.Role, user.Role)
	}
}
