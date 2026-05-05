package setup

import (
	"context"
	"errors"
	"strings"
	"testing"
)

//nolint:paralleltest // mutates singleton setup_state row
func TestEnsureInstallToken_GeneratesOnFirstCall(t *testing.T) {
	ctx := context.Background()

	plaintext, err := ensureInstallToken(ctx)
	if err != nil {
		t.Fatalf("ensureInstallToken: %v", err)
	}

	if plaintext == "" {
		t.Error("expected non-empty token plaintext on first call")
	}

	if len(plaintext) < 32 {
		t.Errorf("token suspiciously short: %d chars", len(plaintext))
	}
}

//nolint:paralleltest // mutates singleton setup_state row
func TestEnsureInstallToken_IdempotentOnSecondCall(t *testing.T) {
	ctx := context.Background()

	if _, err := ensureInstallToken(ctx); err != nil {
		t.Fatalf("first ensureInstallToken: %v", err)
	}

	plaintext, err := ensureInstallToken(ctx)
	if err != nil {
		t.Fatalf("second ensureInstallToken: %v", err)
	}

	if plaintext != "" {
		t.Errorf("second call should return empty plaintext, got %q", plaintext)
	}
}

//nolint:paralleltest // mutates singleton setup_state row
func TestVerifyInstallToken_RejectsWrongToken(t *testing.T) {
	ctx := context.Background()

	if _, err := ensureInstallToken(ctx); err != nil {
		t.Fatalf("ensureInstallToken: %v", err)
	}

	err := verifyInstallToken(ctx, strings.Repeat("x", 32))
	if !errors.Is(err, ErrInvalidToken) && !errors.Is(err, ErrTokenAlreadyConsumed) && !errors.Is(err, ErrTokenLockedOut) {
		t.Errorf("expected ErrInvalidToken/Consumed/LockedOut, got %v", err)
	}
}
