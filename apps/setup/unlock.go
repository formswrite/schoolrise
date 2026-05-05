package setup

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"

	"encore.app/pkg/ratelimit"
)

var unlockLimiter = ratelimit.NewLimiter(5, 2)

type UnlockRequest struct {
	InstallToken string
}

type UnlockResponse struct {
	SessionToken string
	ExpiresAt    time.Time
}

//encore:api public method=POST path=/v1/setup/unlock
func (s *Service) UnlockAPI(ctx context.Context, req *UnlockRequest) (*UnlockResponse, error) {
	if err := requireUnlocked(ctx); err != nil {
		return nil, err
	}

	if req.InstallToken == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "install token required"}
	}

	if err := unlockLimiter.Allow(ctx, "setup-unlock"); err != nil {
		return nil, err
	}

	if err := verifyInstallToken(ctx, req.InstallToken); err != nil {
		switch {
		case errors.Is(err, ErrTokenLockedOut):
			return nil, &errs.Error{Code: errs.ResourceExhausted, Message: "too many failed attempts"}
		case errors.Is(err, ErrTokenAlreadyConsumed):
			return nil, &errs.Error{Code: errs.PermissionDenied, Message: "install token already consumed"}
		default:
			return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid install token"}
		}
	}

	if err := consumeInstallToken(ctx); err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not consume install token"}
	}

	token, expiresAt, err := createSetupSession(ctx)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not create setup session"}
	}

	return &UnlockResponse{SessionToken: token, ExpiresAt: expiresAt}, nil
}
