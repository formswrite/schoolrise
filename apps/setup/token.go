package setup

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	installTokenBytes      = 32
	installTokenBcryptCost = bcrypt.DefaultCost + 2
	maxFailedUnlockAttempts = 5
)

var (
	ErrTokenAlreadyConsumed = errors.New("setup: install token already consumed")
	ErrInvalidToken         = errors.New("setup: invalid install token")
	ErrTokenLockedOut       = errors.New("setup: too many failed token attempts")
)

func generateInstallToken() (plaintext string, hash string, err error) {
	b := make([]byte, installTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	plaintext = base64.RawURLEncoding.EncodeToString(b)

	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(plaintext), installTokenBcryptCost)
	if err != nil {
		return "", "", err
	}

	return plaintext, string(bcryptHash), nil
}

func verifyInstallToken(ctx context.Context, plaintext string) error {
	state, err := queries.GetSetupState(ctx)
	if err != nil {
		return err
	}

	if !state.InstallTokenHash.Valid || state.InstallTokenHash.String == "" {
		return ErrInvalidToken
	}

	if state.InstallTokenConsumedAt.Valid {
		return ErrTokenAlreadyConsumed
	}

	if state.FailedUnlockAttempts >= maxFailedUnlockAttempts {
		return ErrTokenLockedOut
	}

	if err := bcrypt.CompareHashAndPassword([]byte(state.InstallTokenHash.String), []byte(plaintext)); err != nil {
		if incErr := queries.IncrementFailedUnlockAttempts(ctx); incErr != nil {
			return incErr
		}

		return ErrInvalidToken
	}

	if err := queries.ResetFailedUnlockAttempts(ctx); err != nil {
		return err
	}

	return nil
}

func consumeInstallToken(ctx context.Context) error {
	return queries.ConsumeInstallToken(ctx)
}

func ensureInstallToken(ctx context.Context) (string, error) {
	state, err := queries.GetSetupState(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("setup: setup_state row missing — migration not applied")
		}

		return "", err
	}

	if state.InstallTokenHash.Valid && state.InstallTokenHash.String != "" {
		return "", nil
	}

	plaintext, hash, err := generateInstallToken()
	if err != nil {
		return "", err
	}

	if err := queries.SetInstallTokenHash(ctx, sql.NullString{String: hash, Valid: true}); err != nil {
		return "", err
	}

	return plaintext, nil
}
