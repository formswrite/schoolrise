package auth_test

import (
	"errors"
	"strings"
	"testing"

	"encore.app/apps/auth"
)

func TestHashPassword_RoundTrip(t *testing.T) {

	hash, err := auth.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}

	if err := auth.VerifyPassword(hash, "ChangeMe123!"); err != nil {
		t.Errorf("VerifyPassword on correct password returned: %v", err)
	}
}

func TestVerifyPassword_RejectsWrongPassword(t *testing.T) {

	hash, err := auth.HashPassword("CorrectHorseBatteryStaple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	err = auth.VerifyPassword(hash, "WrongPassword")
	if err == nil {
		t.Fatal("VerifyPassword accepted wrong password")
	}

	if !errors.Is(err, auth.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestHashPassword_NonDeterministic(t *testing.T) {

	first, err := auth.HashPassword("SamePassword!")
	if err != nil {
		t.Fatalf("first HashPassword: %v", err)
	}

	second, err := auth.HashPassword("SamePassword!")
	if err != nil {
		t.Fatalf("second HashPassword: %v", err)
	}

	if first == second {
		t.Errorf("two hashes of the same password are identical: %q", first)
	}
}

func TestHashPassword_RejectsEmptyPassword(t *testing.T) {

	_, err := auth.HashPassword("")
	if err == nil {
		t.Fatal("HashPassword accepted empty password")
	}

	if !errors.Is(err, auth.ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestVerifyPassword_RejectsCorruptHash(t *testing.T) {

	err := auth.VerifyPassword("not-a-real-bcrypt-hash", "anything")
	if err == nil {
		t.Fatal("VerifyPassword accepted a corrupt hash")
	}

	if errors.Is(err, auth.ErrPasswordMismatch) {
		t.Errorf("corrupt hash should not surface as ErrPasswordMismatch, got %v", err)
	}
}

func TestHashPassword_HashStartsWithBcryptPrefix(t *testing.T) {

	hash, err := auth.HashPassword("AnyPassword123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("hash does not look like bcrypt: %q", hash)
	}
}
