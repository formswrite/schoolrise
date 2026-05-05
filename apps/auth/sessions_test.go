package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/auth"
)

func newTestUser(t *testing.T) *auth.User {
	t.Helper()

	u, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: uniqueEmail(t), Password: "Pass1234!", FullName: "Session Tester", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return u
}

func TestCreateSession_RoundTrip(t *testing.T) {

	u := newTestUser(t)

	token, session, err := auth.CreateSession(context.Background(), u.ID, "test-agent", "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if token == "" {
		t.Fatal("token is empty")
	}

	if len(token) < 32 {
		t.Errorf("token is suspiciously short (%d chars)", len(token))
	}

	if session.UserID != u.ID {
		t.Errorf("session.UserID = %d, want %d", session.UserID, u.ID)
	}

	if !session.ExpiresAt.After(time.Now()) {
		t.Errorf("ExpiresAt %v should be in the future", session.ExpiresAt)
	}

	looked, err := auth.LookupSession(context.Background(), token)
	if err != nil {
		t.Fatalf("LookupSession: %v", err)
	}

	if looked.UserID != u.ID {
		t.Errorf("looked.UserID = %d, want %d", looked.UserID, u.ID)
	}
}

func TestLookupSession_UnknownToken(t *testing.T) {

	_, err := auth.LookupSession(context.Background(), "no-such-token-zzz")
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestLookupSession_EmptyToken(t *testing.T) {

	_, err := auth.LookupSession(context.Background(), "")
	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound for empty token, got %v", err)
	}
}

func TestRevokeSession_BlocksFutureLookups(t *testing.T) {

	u := newTestUser(t)

	token, _, err := auth.CreateSession(context.Background(), u.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := auth.RevokeSession(context.Background(), token); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}

	_, err = auth.LookupSession(context.Background(), token)
	if !errors.Is(err, auth.ErrSessionRevoked) {
		t.Errorf("expected ErrSessionRevoked, got %v", err)
	}
}

func TestRevokeSession_Idempotent(t *testing.T) {

	u := newTestUser(t)

	token, _, err := auth.CreateSession(context.Background(), u.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if err := auth.RevokeSession(context.Background(), token); err != nil {
		t.Fatalf("first RevokeSession: %v", err)
	}

	if err := auth.RevokeSession(context.Background(), token); err != nil {
		t.Errorf("second RevokeSession should be a no-op, got %v", err)
	}
}

func TestLookupSession_ExpiredSession(t *testing.T) {

	u := newTestUser(t)

	pastExpiry := time.Now().Add(-1 * time.Hour)

	token, _, err := auth.CreateSessionWithExpiry(context.Background(), u.ID, pastExpiry, "", "")
	if err != nil {
		t.Fatalf("CreateSessionWithExpiry: %v", err)
	}

	_, err = auth.LookupSession(context.Background(), token)
	if !errors.Is(err, auth.ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestRevokeAllSessionsForUser_OnlyAffectsThatUser(t *testing.T) {

	userA := newTestUser(t)
	userB := newTestUser(t)

	tokenA1, _, err := auth.CreateSession(context.Background(), userA.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession A1: %v", err)
	}

	tokenA2, _, err := auth.CreateSession(context.Background(), userA.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession A2: %v", err)
	}

	tokenB, _, err := auth.CreateSession(context.Background(), userB.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession B: %v", err)
	}

	if err := auth.RevokeAllSessionsForUser(context.Background(), userA.ID); err != nil {
		t.Fatalf("RevokeAllSessionsForUser: %v", err)
	}

	if _, err := auth.LookupSession(context.Background(), tokenA1); !errors.Is(err, auth.ErrSessionRevoked) {
		t.Errorf("A1: expected ErrSessionRevoked, got %v", err)
	}

	if _, err := auth.LookupSession(context.Background(), tokenA2); !errors.Is(err, auth.ErrSessionRevoked) {
		t.Errorf("A2: expected ErrSessionRevoked, got %v", err)
	}

	if _, err := auth.LookupSession(context.Background(), tokenB); err != nil {
		t.Errorf("B should still be valid, got %v", err)
	}
}

func TestCreateSession_ProducesUniqueTokens(t *testing.T) {

	u := newTestUser(t)

	tokens := map[string]struct{}{}

	for range 5 {
		tok, _, err := auth.CreateSession(context.Background(), u.ID, "", "")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		if _, dup := tokens[tok]; dup {
			t.Fatalf("duplicate token generated: %q", tok)
		}

		tokens[tok] = struct{}{}
	}
}
