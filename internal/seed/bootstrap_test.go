package seed_test

import (
	"errors"
	"strings"
	"testing"

	"encore.app/internal/seed"
)

var requiredVars = []string{
	"POSTGRES_PASSWORD",
	"ADMIN_EMAIL",
	"ADMIN_PASSWORD",
	"AUTH_SECRET",
	"BASE_URL",
	"RESEND_API_KEY",
	"EMAIL_FROM",
	"OPENAI_API_KEY",
}

func setAllRequired(t *testing.T) {
	t.Helper()

	for _, key := range requiredVars {
		t.Setenv(key, "stub-value")
	}
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel
func TestValidateEnv_AllRequiredSet(t *testing.T) {
	setAllRequired(t)

	if err := seed.ValidateEnv(); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel
func TestValidateEnv_OneMissing(t *testing.T) {
	setAllRequired(t)
	t.Setenv("OPENAI_API_KEY", "")

	err := seed.ValidateEnv()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, seed.ErrMissingEnv) {
		t.Errorf("expected ErrMissingEnv, got %v", err)
	}

	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("error should mention OPENAI_API_KEY, got %v", err)
	}
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel
func TestValidateEnv_MultipleMissing(t *testing.T) {
	setAllRequired(t)
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("AUTH_SECRET", "")

	err := seed.ValidateEnv()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("error should mention OPENAI_API_KEY, got %v", err)
	}

	if !strings.Contains(err.Error(), "AUTH_SECRET") {
		t.Errorf("error should mention AUTH_SECRET, got %v", err)
	}
}

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel
func TestValidateEnv_NoneSet(t *testing.T) {
	for _, key := range requiredVars {
		t.Setenv(key, "")
	}

	err := seed.ValidateEnv()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	for _, key := range requiredVars {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error should mention %s, got %v", key, err)
		}
	}
}
