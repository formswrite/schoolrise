package setup

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
)

var ErrSetupAlreadyComplete = errors.New("setup: already complete")

func requireUnlocked(ctx context.Context) error {
	state, err := loadState(ctx)
	if err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not read setup state"}
	}

	if state.SetupCompleteAt {
		return &errs.Error{Code: errs.PermissionDenied, Message: "setup already complete"}
	}

	return nil
}

func requireSession(ctx context.Context, token string) error {
	if err := requireUnlocked(ctx); err != nil {
		return err
	}

	err := validateSetupSession(ctx, token)
	if err == nil {
		return nil
	}

	if errors.Is(err, ErrSetupSessionExpired) {
		return &errs.Error{Code: errs.Unauthenticated, Message: "setup session expired; paste the install token again"}
	}

	return &errs.Error{Code: errs.Unauthenticated, Message: "invalid or missing setup session"}
}
