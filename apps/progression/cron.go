package progression

import (
	"context"

	"encore.dev/cron"
	"encore.dev/rlog"
)

var _ = cron.NewJob("progression-refresh", cron.JobConfig{
	Title:    "Refresh progression snapshots for open campaigns",
	Every:    5 * cron.Minute,
	Endpoint: RefreshOpenCampaignsAPI,
})

//encore:api private method=POST path=/internal/progression/refresh-open
func (s *Service) RefreshOpenCampaignsAPI(ctx context.Context) (*RefreshSummary, error) {
	summary, err := RefreshAllOpenCampaigns(ctx)
	if err != nil {
		rlog.Error("progression-refresh failed", "error", err.Error())
		return nil, err
	}
	rlog.Info("progression-refresh completed",
		"campaigns_scanned", summary.CampaignsScanned,
		"scopes_refreshed", summary.ScopesRefreshed,
		"errors", summary.Errors,
		"duration_ms", summary.DurationMs,
	)
	return summary, nil
}
