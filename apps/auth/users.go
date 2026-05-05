package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/auth/dbauth"
)

const maxFailedAttempts = 5

var (
	ErrUserNotFound       = errors.New("auth: user not found")
	ErrUserLocked         = errors.New("auth: user locked")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrEmailAlreadyExists = errors.New("auth: email already exists")
	ErrInvalidUserInput   = errors.New("auth: invalid user input")
)

type User struct {
	ID                 int64
	Email              string
	FullName           string
	Role               string
	MustChangePassword bool
	LockedAt           *time.Time
	FailedAttempts     int
	LastLoginAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateUserParams struct {
	Email              string
	Password           string
	FullName           string
	Role               string
	MustChangePassword bool
}

func CreateUser(ctx context.Context, p CreateUserParams) (*User, error) {
	email := strings.ToLower(strings.TrimSpace(p.Email))
	if email == "" {
		return nil, ErrInvalidUserInput
	}

	if strings.TrimSpace(p.FullName) == "" {
		return nil, ErrInvalidUserInput
	}

	if strings.TrimSpace(p.Role) == "" {
		return nil, ErrInvalidUserInput
	}

	hash, err := HashPassword(p.Password)
	if err != nil {
		return nil, err
	}

	row, err := queries.CreateUser(ctx, dbauth.CreateUserParams{
		Email:              email,
		PasswordHash:       hash,
		FullName:           p.FullName,
		Role:               p.Role,
		MustChangePassword: p.MustChangePassword,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}

		return nil, err
	}

	return userFromRow(row), nil
}

func GetUserByID(ctx context.Context, id int64) (*User, error) {
	row, err := queries.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return userFromRow(row), nil
}

func ListUsers(ctx context.Context, limit, offset int32) ([]*User, error) {
	if limit <= 0 {
		limit = 200
	}

	rows, err := queries.ListUsers(ctx, dbauth.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	out := make([]*User, 0, len(rows))
	for _, r := range rows {
		out = append(out, userFromRow(r))
	}

	return out, nil
}

func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row, err := queries.GetUserByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return userFromRow(row), nil
}

func VerifyCredentials(ctx context.Context, email, plain string) (*User, error) {
	user, err := GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if user.LockedAt != nil {
		return nil, ErrUserLocked
	}

	hash, err := queries.GetUserPasswordHash(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if err := VerifyPassword(hash, plain); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			if regErr := queries.IncrementFailedAttempts(ctx, dbauth.IncrementFailedAttemptsParams{
				ID:             user.ID,
				FailedAttempts: maxFailedAttempts,
			}); regErr != nil {
				return nil, regErr
			}

			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if err := queries.ResetFailedAttempts(ctx, user.ID); err != nil {
		return nil, err
	}

	now := time.Now()
	user.FailedAttempts = 0
	user.LastLoginAt = &now

	return user, nil
}

func userFromRow(r dbauth.User) *User {
	u := &User{
		ID:                 r.ID,
		Email:              r.Email,
		FullName:           r.FullName,
		Role:               r.Role,
		MustChangePassword: r.MustChangePassword,
		FailedAttempts:     int(r.FailedAttempts),
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}

	if r.LockedAt.Valid {
		t := r.LockedAt.Time
		u.LockedAt = &t
	}

	if r.LastLoginAt.Valid {
		t := r.LastLoginAt.Time
		u.LastLoginAt = &t
	}

	return u
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
