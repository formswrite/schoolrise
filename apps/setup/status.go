package setup

import (
	"context"
)

type StatusResponse struct {
	SetupComplete    bool `json:"setupComplete"`
	InstallTokenSet  bool `json:"installTokenSet"`
	TokenConsumed    bool `json:"tokenConsumed"`
	FailedAttempts   int  `json:"failedAttempts"`
}

//encore:api public method=GET path=/v1/setup/status
func (s *Service) StatusAPI(ctx context.Context) (*StatusResponse, error) {
	state, err := loadState(ctx)
	if err != nil {
		return nil, err
	}

	return &StatusResponse{
		SetupComplete:   state.SetupCompleteAt,
		InstallTokenSet: state.InstallTokenIssued,
		TokenConsumed:   state.InstallTokenConsumedAt,
		FailedAttempts:  state.FailedUnlockAttempts,
	}, nil
}
