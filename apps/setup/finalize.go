package setup

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
)

var requiredSteps = []string{"admin", "system", "levels"}

type FinalizeRequest struct {
	SessionToken string
}

//encore:api public method=POST path=/v1/setup/finalize
func (s *Service) FinalizeAPI(ctx context.Context, req *FinalizeRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	for _, step := range requiredSteps {
		row, err := queries.GetSetupProgressStep(ctx, step)
		if err != nil {
			if errors.Is(err, errSQLNoRows()) {
				return &errs.Error{Code: errs.FailedPrecondition, Message: "step " + step + " not completed"}
			}

			return &errs.Error{Code: errs.Internal, Message: "could not read progress"}
		}

		if !row.CompletedAt.Valid {
			return &errs.Error{Code: errs.FailedPrecondition, Message: "step " + step + " not completed"}
		}
	}

	if err := queries.MarkSetupComplete(ctx); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not finalize setup"}
	}

	return nil
}
