package auth

import (
	"context"
	"errors"
	"strconv"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/apps/auth/dbauth"
	"encore.app/pkg/ratelimit"
)

var changePasswordLimiter = ratelimit.NewLimiter(5, 2)

const minPasswordLength = 8

type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
}

//encore:api auth method=POST path=/v1/auth/change-password
func (s *Service) ChangePasswordAPI(ctx context.Context, req *ChangePasswordRequest) error {
	if len(req.NewPassword) < minPasswordLength {
		return &errs.Error{Code: errs.InvalidArgument, Message: "new password must be at least 8 characters"}
	}

	data, ok := encauth.Data().(*AuthData)
	if !ok || data == nil {
		return &errs.Error{Code: errs.Internal, Message: "missing auth data"}
	}

	if err := changePasswordLimiter.Allow(ctx, "change-password:"+strconv.FormatInt(data.UserID, 10)); err != nil {
		return err
	}

	currentHash, err := queries.GetUserPasswordHash(ctx, data.UserID)
	if err != nil {
		return &errs.Error{Code: errs.Internal, Message: "lookup failed"}
	}

	if err := VerifyPassword(currentHash, req.CurrentPassword); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return &errs.Error{Code: errs.Unauthenticated, Message: "current password is incorrect"}
		}

		return &errs.Error{Code: errs.Internal, Message: "verification failed"}
	}

	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		return &errs.Error{Code: errs.InvalidArgument, Message: "invalid new password"}
	}

	if err := queries.UpdateUserPassword(ctx, dbauth.UpdateUserPasswordParams{
		ID:           data.UserID,
		PasswordHash: newHash,
	}); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "update failed"}
	}

	if err := RevokeAllSessionsForUser(ctx, data.UserID); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "session cleanup failed"}
	}

	return nil
}
