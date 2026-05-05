package auth_test

import (
	"context"
	"testing"
	"time"

	"encore.app/apps/auth"
)

func TestAuthHandler_ValidToken(t *testing.T) {

	u := newTestUser(t)

	token, _, err := auth.CreateSession(context.Background(), u.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	uid, data, err := auth.AuthHandler(context.Background(), token)
	if err != nil {
		t.Fatalf("AuthHandler: %v", err)
	}

	if uid == "" {
		t.Error("uid is empty")
	}

	if data == nil {
		t.Fatal("data is nil")
	}

	if data.UserID != u.ID {
		t.Errorf("data.UserID = %d, want %d", data.UserID, u.ID)
	}

	if data.Email != u.Email {
		t.Errorf("data.Email = %q, want %q", data.Email, u.Email)
	}

	if data.Role != u.Role {
		t.Errorf("data.Role = %q, want %q", data.Role, u.Role)
	}
}

func TestAuthHandler_StripsBearerPrefix(t *testing.T) {

	u := newTestUser(t)

	token, _, err := auth.CreateSession(context.Background(), u.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	_, data, err := auth.AuthHandler(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("AuthHandler with Bearer prefix: %v", err)
	}

	if data.UserID != u.ID {
		t.Errorf("data.UserID = %d, want %d", data.UserID, u.ID)
	}
}

func TestAuthHandler_RejectsExpiredSession(t *testing.T) {

	u := newTestUser(t)

	token, _, err := auth.CreateSessionWithExpiry(context.Background(), u.ID, time.Now().Add(-1*time.Hour), "", "")
	if err != nil {
		t.Fatalf("CreateSessionWithExpiry: %v", err)
	}

	_, _, err = auth.AuthHandler(context.Background(), token)
	if err == nil {
		t.Fatal("AuthHandler accepted an expired session")
	}
}

func TestAuthHandler_RejectsRevokedSession(t *testing.T) {

	u := newTestUser(t)

	token, _, err := auth.CreateSession(context.Background(), u.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := auth.RevokeSession(context.Background(), token); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}

	_, _, err = auth.AuthHandler(context.Background(), token)
	if err == nil {
		t.Fatal("AuthHandler accepted a revoked session")
	}
}

func TestAuthHandler_RejectsUnknownToken(t *testing.T) {

	_, _, err := auth.AuthHandler(context.Background(), "some-random-token-that-was-never-issued")
	if err == nil {
		t.Fatal("AuthHandler accepted an unknown token")
	}
}

func TestAuthHandler_RejectsEmptyToken(t *testing.T) {

	_, _, err := auth.AuthHandler(context.Background(), "")
	if err == nil {
		t.Fatal("AuthHandler accepted an empty token")
	}
}
