package setup

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"encore.app/apps/setup/dbsetup"
	"encore.app/pkg/secrets"
)

const setupSessionTTL = 30 * time.Minute

var (
	ErrSetupSessionNotFound = errors.New("setup: session not found")
	ErrSetupSessionExpired  = errors.New("setup: session expired")
)

var sessionHMACKey = sync.OnceValue(func() []byte {
	return secrets.MustEnv("AUTH_SECRET")
})

func createSetupSession(ctx context.Context) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(setupSessionTTL)

	if err := queries.CreateSetupSession(ctx, dbsetup.CreateSetupSessionParams{
		TokenHash: hmacSetupToken(token),
		ExpiresAt: expiresAt,
	}); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func validateSetupSession(ctx context.Context, token string) error {
	if token == "" {
		return ErrSetupSessionNotFound
	}

	row, err := queries.GetSetupSessionByHash(ctx, hmacSetupToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSetupSessionNotFound
	}

	if err != nil {
		return err
	}

	if time.Now().After(row.ExpiresAt) {
		return ErrSetupSessionExpired
	}

	return nil
}

func hmacSetupToken(token string) []byte {
	mac := hmac.New(sha256.New, sessionHMACKey())
	mac.Write([]byte(token))

	return mac.Sum(nil)
}
