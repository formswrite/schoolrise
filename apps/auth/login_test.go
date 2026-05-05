package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.dev/beta/errs"

	"encore.app/apps/auth"
)

func loginService(t *testing.T) *auth.Service {
	t.Helper()
	return &auth.Service{}
}

func TestLogin_ValidCreds(t *testing.T) {

	email := uniqueEmail(t)

	created, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "Login Tester", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resp, err := loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
		Email: email, Password: "GoodPass1!",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if resp.SessionToken == "" {
		t.Error("SessionToken empty")
	}

	if resp.UserID != created.ID {
		t.Errorf("UserID = %d, want %d", resp.UserID, created.ID)
	}

	if resp.Email != email {
		t.Errorf("Email mismatch")
	}

	if resp.Role != "teacher" {
		t.Errorf("Role mismatch")
	}

	if !resp.ExpiresAt.After(time.Now()) {
		t.Errorf("ExpiresAt should be in the future, got %v", resp.ExpiresAt)
	}

	if _, err := auth.LookupSession(context.Background(), resp.SessionToken); err != nil {
		t.Errorf("returned token cannot be looked up: %v", err)
	}
}

func TestLogin_BadPassword(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	_, err = loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
		Email: email, Password: "WrongPass!",
	})
	if err == nil {
		t.Fatal("Login accepted wrong password")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) {
		t.Fatalf("expected *errs.Error, got %T: %v", err, err)
	}

	if errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", errsErr.Code)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {

	_, err := loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
		Email: "ghost@nope.local", Password: "anything",
	})
	if err == nil {
		t.Fatal("Login accepted unknown email")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) {
		t.Fatalf("expected *errs.Error, got %T: %v", err, err)
	}

	if errsErr.Code != errs.Unauthenticated {
		t.Errorf("unknown email should produce Unauthenticated (no account enum), got %v", errsErr.Code)
	}
}

func TestLogin_LockedUser(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	for range 5 {
		_, _ = loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
			Email: email, Password: "WrongPass!",
		})
	}

	_, err = loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
		Email: email, Password: "GoodPass1!",
	})
	if err == nil {
		t.Fatal("Login accepted credentials for a locked user")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) {
		t.Fatalf("expected *errs.Error, got %T: %v", err, err)
	}

	if errsErr.Code != errs.FailedPrecondition {
		t.Errorf("locked user should produce FailedPrecondition, got %v", errsErr.Code)
	}
}

func TestLogin_SurfacesMustChangePasswordFlag(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
		MustChangePassword: true,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	resp, err := loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
		Email: email, Password: "GoodPass1!",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if !resp.MustChangePassword {
		t.Error("MustChangePassword flag not surfaced in response")
	}
}

func TestLogin_RejectsEmptyEmail(t *testing.T) {

	_, err := loginService(t).LoginAPI(context.Background(), &auth.LoginRequest{
		Email: "", Password: "anything",
	})
	if err == nil {
		t.Fatal("Login accepted empty email")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) {
		t.Fatalf("expected *errs.Error, got %T", err)
	}

	if errsErr.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", errsErr.Code)
	}
}
