package progression

import (
	"context"
	"errors"

	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type ProgressionRequest struct {
	ScopeNodeID int64 `query:"scope_node_id"`
	PeriodID    int64 `query:"period_id"`
	CampaignID  int64 `query:"campaign_id"`
}

//encore:api auth method=GET path=/v1/progression
func (s *Service) GetProgressionAPI(ctx context.Context, req *ProgressionRequest) (*ScopeProgression, error) {
	prog, err := GetSnapshot(ctx, req.ScopeNodeID, req.PeriodID, req.CampaignID)
	if err != nil {
		return nil, mapErr(err)
	}
	return prog, nil
}

//encore:api auth method=GET path=/v1/progression/snapshot
func (s *Service) GetSnapshotAPI(ctx context.Context, req *ProgressionRequest) (*ScopeProgression, error) {
	prog, err := GetSnapshot(ctx, req.ScopeNodeID, req.PeriodID, req.CampaignID)
	if err != nil {
		return nil, mapErr(err)
	}
	return prog, nil
}

//encore:api auth method=POST path=/v1/progression/refresh
func (s *Service) RefreshSnapshotAPI(ctx context.Context, req *ProgressionRequest) (*ScopeProgression, error) {
	prog, err := RefreshSnapshot(ctx, req.ScopeNodeID, req.PeriodID, req.CampaignID)
	if err != nil {
		return nil, mapErr(err)
	}
	return prog, nil
}

//encore:api auth method=GET path=/v1/progression/drilldown
func (s *Service) GetDrilldownAPI(ctx context.Context, req *ProgressionRequest) (*Drilldown, error) {
	d, err := DrilldownByScopeViaSnapshots(ctx, req.ScopeNodeID, req.PeriodID, req.CampaignID)
	if err != nil {
		return nil, mapErr(err)
	}
	return d, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	case errors.Is(err, ErrCampaignNotFound):
		return &errs.Error{Code: errs.NotFound, Message: err.Error()}
	}
	return apierr.WrapInternal("progression", err)
}
