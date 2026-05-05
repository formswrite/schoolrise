package auth

import (
	"context"
	"strconv"
	"strings"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type AuthData struct {
	UserID             int64
	SessionID          int64
	Email              string
	Role               string
	MustChangePassword bool
}

//encore:authhandler
func AuthHandler(ctx context.Context, token string) (encauth.UID, *AuthData, error) {
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	if token == "" {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing session token"}
	}

	session, err := LookupSession(ctx, token)
	if err != nil {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid or expired session"}
	}

	user, err := GetUserByID(ctx, session.UserID)
	if err != nil {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "user not found"}
	}

	if user.LockedAt != nil {
		return "", nil, &errs.Error{Code: errs.Unauthenticated, Message: "user locked"}
	}

	uid := encauth.UID(strconv.FormatInt(user.ID, 10))

	return uid, &AuthData{
		UserID:             user.ID,
		SessionID:          session.ID,
		Email:              user.Email,
		Role:               user.Role,
		MustChangePassword: user.MustChangePassword,
	}, nil
}
