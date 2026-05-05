package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/netip"
	"sync"
	"time"

	"github.com/sqlc-dev/pqtype"

	"encore.app/apps/auth/dbauth"
	"encore.app/pkg/secrets"
)

const defaultSessionTTL = 7 * 24 * time.Hour

var (
	ErrSessionNotFound = errors.New("auth: session not found")
	ErrSessionRevoked  = errors.New("auth: session revoked")
	ErrSessionExpired  = errors.New("auth: session expired")
)

var hmacKey = sync.OnceValue(func() []byte {
	return secrets.MustEnv("AUTH_SECRET")
})

type Session struct {
	ID        int64
	UserID    int64
	ExpiresAt time.Time
	RevokedAt *time.Time
	UserAgent string
	IP        string
	CreatedAt time.Time
}

func CreateSession(ctx context.Context, userID int64, userAgent, ip string) (string, *Session, error) {
	return CreateSessionWithExpiry(ctx, userID, time.Now().Add(defaultSessionTTL), userAgent, ip)
}

func CreateSessionWithExpiry(ctx context.Context, userID int64, expiresAt time.Time, userAgent, ip string) (string, *Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", nil, err
	}

	row, err := queries.CreateSession(ctx, dbauth.CreateSessionParams{
		UserID:    userID,
		TokenHash: hmacToken(token),
		ExpiresAt: expiresAt,
		Column4:   userAgent,
		Column5:   ip,
	})
	if err != nil {
		return "", nil, err
	}

	return token, sessionFromRow(row), nil
}

func LookupSession(ctx context.Context, token string) (*Session, error) {
	if token == "" {
		return nil, ErrSessionNotFound
	}

	row, err := queries.GetSessionByTokenHash(ctx, hmacToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	if row.RevokedAt.Valid {
		return nil, ErrSessionRevoked
	}

	if time.Now().After(row.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return sessionFromRow(row), nil
}

func RevokeSession(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	return queries.RevokeSessionByTokenHash(ctx, hmacToken(token))
}

func RevokeSessionByID(ctx context.Context, id int64) error {
	return queries.RevokeSessionByID(ctx, id)
}

func RevokeAllSessionsForUser(ctx context.Context, userID int64) error {
	return queries.RevokeAllSessionsForUser(ctx, userID)
}

func sessionFromRow(r dbauth.Session) *Session {
	s := &Session{
		ID:        r.ID,
		UserID:    r.UserID,
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
	}

	if r.RevokedAt.Valid {
		t := r.RevokedAt.Time
		s.RevokedAt = &t
	}

	if r.UserAgent.Valid {
		s.UserAgent = r.UserAgent.String
	}

	s.IP = inetToString(r.Ip)

	return s
}

func inetToString(ip pqtype.Inet) string {
	if !ip.Valid {
		return ""
	}

	addr, ok := netip.AddrFromSlice(ip.IPNet.IP)
	if !ok {
		return ""
	}

	return addr.String()
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hmacToken(token string) []byte {
	mac := hmac.New(sha256.New, hmacKey())
	mac.Write([]byte(token))

	return mac.Sum(nil)
}
