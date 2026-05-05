package auth_test

import (
	"context"
	"errors"
	"testing"

	"encore.dev/beta/errs"

	"encore.app/apps/auth"
)

func TestChangePassword_RequiresCorrectCurrentPassword(t *testing.T) {

	user := newTestUser(t)
	_, session, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	setAuthInfo(t, user, session.ID)

	err = loginService(t).ChangePasswordAPI(context.Background(), &auth.ChangePasswordRequest{
		CurrentPassword: "WrongCurrent",
		NewPassword:     "NewPass1234!",
	})
	if err == nil {
		t.Fatal("ChangePassword accepted wrong current password")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) {
		t.Fatalf("expected *errs.Error, got %T", err)
	}

	if errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", errsErr.Code)
	}
}

func TestChangePassword_SetsNewPasswordAndClearsFlag(t *testing.T) {

	user := newTestUser(t)
	_, session, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	setAuthInfo(t, user, session.ID)

	err = loginService(t).ChangePasswordAPI(context.Background(), &auth.ChangePasswordRequest{
		CurrentPassword: "Pass1234!",
		NewPassword:     "NewPass1234!",
	})
	if err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}

	verified, err := auth.VerifyCredentials(context.Background(), user.Email, "NewPass1234!")
	if err != nil {
		t.Fatalf("VerifyCredentials with new password: %v", err)
	}

	if verified.MustChangePassword {
		t.Error("MustChangePassword should be false after change")
	}

	if _, err := auth.VerifyCredentials(context.Background(), user.Email, "TestPassword123!"); err == nil {
		t.Error("old password should no longer work")
	}
}

func TestChangePassword_RejectsEmptyOrShortPassword(t *testing.T) {

	user := newTestUser(t)
	_, session, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	setAuthInfo(t, user, session.ID)

	cases := []string{"", "short"}

	for _, candidate := range cases {
		err := loginService(t).ChangePasswordAPI(context.Background(), &auth.ChangePasswordRequest{
			CurrentPassword: "Pass1234!",
			NewPassword:     candidate,
		})

		var errsErr *errs.Error
		if !errors.As(err, &errsErr) || errsErr.Code != errs.InvalidArgument {
			t.Errorf("password %q: expected InvalidArgument, got %v", candidate, err)
		}
	}
}

func TestChangePassword_RevokesOtherSessions(t *testing.T) {

	user := newTestUser(t)

	_, current, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession current: %v", err)
	}

	otherToken, _, err := auth.CreateSession(context.Background(), user.ID, "", "")
	if err != nil {
		t.Fatalf("CreateSession other: %v", err)
	}

	setAuthInfo(t, user, current.ID)

	if err := loginService(t).ChangePasswordAPI(context.Background(), &auth.ChangePasswordRequest{
		CurrentPassword: "Pass1234!",
		NewPassword:     "NewPass1234!",
	}); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}

	if _, err := auth.LookupSession(context.Background(), otherToken); !errors.Is(err, auth.ErrSessionRevoked) {
		t.Errorf("other session should be revoked, got %v", err)
	}
}
