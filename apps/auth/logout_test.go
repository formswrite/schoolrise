package auth_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	encauth "encore.dev/beta/auth"
	"encore.dev/et"

	"encore.app/apps/auth"
)

func setAuthInfo(t *testing.T, user *auth.User, sessionID int64) {
	t.Helper()

	et.OverrideAuthInfo(
		encauth.UID(strconv.FormatInt(user.ID, 10)),
		&auth.AuthData{
			UserID:    user.ID,
			SessionID: sessionID,
			Email:     user.Email,
			Role:      user.Role,
		},
	)
}

func TestLogout_RevokesActiveSession(t *testing.T) {

	user := newTestUser(t)

	token, session, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	setAuthInfo(t, user, session.ID)

	if err := loginService(t).LogoutAPI(context.Background()); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	_, err = auth.LookupSession(context.Background(), token)
	if !errors.Is(err, auth.ErrSessionRevoked) {
		t.Errorf("expected ErrSessionRevoked after Logout, got %v", err)
	}
}

func TestLogout_IsIdempotent(t *testing.T) {

	user := newTestUser(t)

	_, session, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	setAuthInfo(t, user, session.ID)

	if err := loginService(t).LogoutAPI(context.Background()); err != nil {
		t.Fatalf("first Logout: %v", err)
	}

	if err := loginService(t).LogoutAPI(context.Background()); err != nil {
		t.Errorf("second Logout should be a no-op, got %v", err)
	}
}
