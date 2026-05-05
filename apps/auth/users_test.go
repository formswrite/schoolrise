package auth_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"encore.app/apps/auth"
)

var emailCounter atomic.Uint64

func uniqueEmail(t *testing.T) string {
	t.Helper()

	n := emailCounter.Add(1)

	return fmt.Sprintf("test-%s-%d@local.test", strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")), n)
}

func TestCreateUser_RoundTrip(t *testing.T) {

	email := uniqueEmail(t)

	created, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email:              email,
		Password:           "Pass1234!",
		FullName:           "Alice Tester",
		Role:               "admin",
		MustChangePassword: true,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if created.ID == 0 {
		t.Error("ID not assigned")
	}

	if created.Email != email {
		t.Errorf("Email = %q, want %q", created.Email, email)
	}

	if created.Role != "admin" {
		t.Errorf("Role = %q, want admin", created.Role)
	}

	if !created.MustChangePassword {
		t.Error("MustChangePassword should be true")
	}

	if created.FailedAttempts != 0 {
		t.Errorf("FailedAttempts = %d, want 0", created.FailedAttempts)
	}

	fetched, err := auth.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("fetched ID = %d, want %d", fetched.ID, created.ID)
	}
}

func TestCreateUser_NormalizesEmail(t *testing.T) {

	email := uniqueEmail(t)
	mixed := strings.ToUpper(email[:5]) + email[5:]

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email:    mixed,
		Password: "Pass1234!",
		FullName: "U",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser with mixed case: %v", err)
	}

	fetched, err := auth.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("GetUserByEmail lowercase: %v", err)
	}

	if !strings.EqualFold(fetched.Email, email) {
		t.Errorf("fetched email %q does not match %q", fetched.Email, email)
	}
}

func TestCreateUser_RejectsDuplicateEmail(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "Pass1234!", FullName: "A", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("first CreateUser: %v", err)
	}

	_, err = auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "Different1!", FullName: "B", Role: "teacher",
	})
	if !errors.Is(err, auth.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestCreateUser_RejectsEmptyEmail(t *testing.T) {

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: "", Password: "Pass1234!", FullName: "X", Role: "teacher",
	})
	if err == nil {
		t.Fatal("CreateUser accepted empty email")
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {

	_, err := auth.GetUserByEmail(context.Background(), "no-such-user@nope.local")
	if !errors.Is(err, auth.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestVerifyCredentials_Valid(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	user, err := auth.VerifyCredentials(context.Background(), email, "GoodPass1!")
	if err != nil {
		t.Fatalf("VerifyCredentials: %v", err)
	}

	if user.Email != email {
		t.Errorf("got email %q, want %q", user.Email, email)
	}
}

func TestVerifyCredentials_BadPassword(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	_, err = auth.VerifyCredentials(context.Background(), email, "WrongPass!")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if errors.Is(err, auth.ErrPasswordMismatch) {
		t.Error("VerifyCredentials must not surface ErrPasswordMismatch (account enumeration risk)")
	}
}

func TestVerifyCredentials_UnknownEmail(t *testing.T) {

	_, err := auth.VerifyCredentials(context.Background(), "ghost@nope.local", "anything")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	if errors.Is(err, auth.ErrUserNotFound) {
		t.Error("VerifyCredentials must not surface ErrUserNotFound (account enumeration risk)")
	}
}

func TestVerifyCredentials_LocksAfterMaxFailures(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	for i := range 5 {
		_, attemptErr := auth.VerifyCredentials(context.Background(), email, "WrongPass!")
		if !errors.Is(attemptErr, auth.ErrInvalidCredentials) {
			t.Errorf("attempt %d: expected ErrInvalidCredentials, got %v", i+1, attemptErr)
		}
	}

	_, err = auth.VerifyCredentials(context.Background(), email, "GoodPass1!")
	if !errors.Is(err, auth.ErrUserLocked) {
		t.Errorf("expected ErrUserLocked after 5 failed attempts, got %v", err)
	}
}

func TestVerifyCredentials_ResetsFailedAttemptsOnSuccess(t *testing.T) {

	email := uniqueEmail(t)

	_, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email: email, Password: "GoodPass1!", FullName: "X", Role: "teacher",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	for range 2 {
		_, _ = auth.VerifyCredentials(context.Background(), email, "WrongPass!")
	}

	_, err = auth.VerifyCredentials(context.Background(), email, "GoodPass1!")
	if err != nil {
		t.Fatalf("VerifyCredentials after 2 fails: %v", err)
	}

	fetched, err := auth.GetUserByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}

	if fetched.FailedAttempts != 0 {
		t.Errorf("FailedAttempts = %d after successful login, want 0", fetched.FailedAttempts)
	}

	if fetched.LastLoginAt == nil {
		t.Error("LastLoginAt should be set after successful login")
	}
}
