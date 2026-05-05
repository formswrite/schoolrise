package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"encore.dev/beta/errs"

	"encore.app/pkg/ratelimit"
)

var loginLimiter = ratelimit.NewLimiter(20, 10)

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	SessionToken       string
	ExpiresAt          time.Time
	UserID             int64
	Email              string
	FullName           string
	Role               string
	MustChangePassword bool
}

//encore:api public method=POST path=/v1/auth/login
func (s *Service) LoginAPI(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "email and password required"}
	}

	if err := loginLimiter.Allow(ctx, "login:"+email); err != nil {
		return nil, err
	}

	user, err := VerifyCredentials(ctx, email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
		case errors.Is(err, ErrUserLocked):
			return nil, &errs.Error{Code: errs.FailedPrecondition, Message: "user locked — contact administrator"}
		default:
			return nil, &errs.Error{Code: errs.Internal, Message: "login failed"}
		}
	}

	token, session, err := CreateSession(ctx, user.ID, "", "")
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "session creation failed"}
	}

	return &LoginResponse{
		SessionToken:       token,
		ExpiresAt:          session.ExpiresAt,
		UserID:             user.ID,
		Email:              user.Email,
		FullName:           user.FullName,
		Role:               user.Role,
		MustChangePassword: user.MustChangePassword,
	}, nil
}
