package setup

import (
	"context"
)

type State struct {
	SetupCompleteAt        bool
	InstallTokenIssued     bool
	InstallTokenConsumedAt bool
	FailedUnlockAttempts   int
}

func loadState(ctx context.Context) (*State, error) {
	row, err := queries.GetSetupState(ctx)
	if err != nil {
		return nil, err
	}

	return &State{
		SetupCompleteAt:        row.SetupCompleteAt.Valid,
		InstallTokenIssued:     row.InstallTokenHash.Valid && row.InstallTokenHash.String != "",
		InstallTokenConsumedAt: row.InstallTokenConsumedAt.Valid,
		FailedUnlockAttempts:   int(row.FailedUnlockAttempts),
	}, nil
}

func IsSetupComplete(ctx context.Context) (bool, error) {
	s, err := loadState(ctx)
	if err != nil {
		return false, err
	}

	return s.SetupCompleteAt, nil
}
