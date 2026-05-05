package auth

import (
	"context"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

//encore:api auth method=POST path=/v1/auth/logout
func (s *Service) LogoutAPI(ctx context.Context) error {
	data, ok := encauth.Data().(*AuthData)
	if !ok || data == nil {
		return &errs.Error{Code: errs.Internal, Message: "missing auth data"}
	}

	if err := RevokeSessionByID(ctx, data.SessionID); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "logout failed"}
	}

	return nil
}
