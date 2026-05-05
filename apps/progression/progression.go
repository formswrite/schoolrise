package progression

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"encore.app/apps/assessment"
	"encore.app/apps/enrollment"
	"encore.app/apps/people"
	"encore.app/apps/progression/dbprogression"
	"encore.app/apps/tenancy"
)

const RefreshAllOpenCampaignsLimit = 1000

var (
	ErrInvalidInput      = errors.New("progression: invalid input")
	ErrCampaignNotFound  = errors.New("progression: campaign not found")
)

type BandRow struct {
	BandCode     string `json:"band_code"`
	BandOrdinal  int32  `json:"band_ordinal"`
	BandLabel    string `json:"band_label"`
	StudentCount int32  `json:"student_count"`
	Percentage   int32  `json:"percentage"`
}

type ScopeProgression struct {
	ScopeNodeID  int64     `json:"scope_node_id"`
	PeriodID     int64     `json:"period_id"`
	CampaignID   int64     `json:"campaign_id"`
	TotalScored  int32     `json:"total_scored"`
	Bands        []BandRow `json:"bands"`
	GeneratedAt  time.Time `json:"generated_at"`
}

type DrilldownChild struct {
	NodeID    int64     `json:"node_id"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	Level     string    `json:"level"`
	Bands     []BandRow `json:"bands"`
	Total     int32     `json:"total"`
}

type Drilldown struct {
	Scope    ScopeProgression `json:"scope"`
	Children []DrilldownChild `json:"children"`
}

func ComputeProgression(ctx context.Context, scopeNodeID, periodID, campaignID int64) (*ScopeProgression, error) {
	if scopeNodeID <= 0 || periodID <= 0 || campaignID <= 0 {
		return nil, ErrInvalidInput
	}

	camp, err := assessment.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, assessment.ErrCampaignNotFound) {
			return nil, ErrCampaignNotFound
		}
		return nil, err
	}

	bands, err := assessment.ListBands(ctx, camp.ScaleCode)
	if err != nil {
		return nil, err
	}

	scores, err := assessment.ListScores(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	descendantIDs, err := tenancy.ListDescendantIDs(ctx, scopeNodeID)
	if err != nil {
		return nil, err
	}
	if len(descendantIDs) == 0 {
		descendantIDs = []int64{scopeNodeID}
	}
	descendantSet := make(map[int64]struct{}, len(descendantIDs))
	for _, id := range descendantIDs {
		descendantSet[id] = struct{}{}
	}

	institutionIDs := make([]int64, 0, len(descendantIDs))
	for _, id := range descendantIDs {
		institutionIDs = append(institutionIDs, id)
	}
	enrolled, err := enrollment.ListActiveStudentsInScope(ctx, institutionIDs, periodID)
	if err != nil {
		return nil, err
	}
	studentInScope := make(map[int64]bool, len(enrolled))
	for studentID := range enrolled {
		studentInScope[studentID] = true
	}

	bandCounts := make(map[string]int32, len(bands))
	for _, s := range scores {
		if !studentInScope[s.StudentID] {
			continue
		}
		bandCounts[s.BandCode]++
	}

	totalScored := int32(0)
	for _, c := range bandCounts {
		totalScored += c
	}

	out := &ScopeProgression{
		ScopeNodeID: scopeNodeID, PeriodID: periodID, CampaignID: campaignID,
		TotalScored: totalScored, GeneratedAt: time.Now(),
		Bands: make([]BandRow, 0, len(bands)),
	}
	for _, b := range bands {
		count := bandCounts[b.Code]
		pct := int32(0)
		if totalScored > 0 {
			pct = int32(float64(count) / float64(totalScored) * 100)
		}
		out.Bands = append(out.Bands, BandRow{
			BandCode: b.Code, BandOrdinal: b.Ordinal, BandLabel: b.Label,
			StudentCount: count, Percentage: pct,
		})
	}

	return out, nil
}

func RefreshSnapshot(ctx context.Context, scopeNodeID, periodID, campaignID int64) (*ScopeProgression, error) {
	logRow, err := queries.CreateRefreshLog(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	prog, computeErr := ComputeProgression(ctx, scopeNodeID, periodID, campaignID)
	if computeErr != nil {
		_ = queries.CompleteRefreshLog(ctx, dbprogression.CompleteRefreshLogParams{
			ID: logRow.ID, RowsWritten: 0,
			Error: sql.NullString{String: computeErr.Error(), Valid: true},
		})
		return nil, computeErr
	}

	for _, band := range prog.Bands {
		if err := queries.UpsertSnapshot(ctx, dbprogression.UpsertSnapshotParams{
			ScopeNodeID:  scopeNodeID,
			PeriodID:     periodID,
			CampaignID:   campaignID,
			BandCode:     band.BandCode,
			BandOrdinal:  band.BandOrdinal,
			StudentCount: band.StudentCount,
		}); err != nil {
			_ = queries.CompleteRefreshLog(ctx, dbprogression.CompleteRefreshLogParams{
				ID: logRow.ID, RowsWritten: 0,
				Error: sql.NullString{String: err.Error(), Valid: true},
			})
			return nil, err
		}
	}

	_ = queries.CompleteRefreshLog(ctx, dbprogression.CompleteRefreshLogParams{
		ID: logRow.ID, RowsWritten: int32(len(prog.Bands)),
	})

	return prog, nil
}

func GetSnapshot(ctx context.Context, scopeNodeID, periodID, campaignID int64) (*ScopeProgression, error) {
	rows, err := queries.ListSnapshotsForScope(ctx, dbprogression.ListSnapshotsForScopeParams{
		ScopeNodeID: scopeNodeID, PeriodID: periodID, CampaignID: campaignID,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return ComputeProgression(ctx, scopeNodeID, periodID, campaignID)
	}

	camp, err := assessment.GetCampaignByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	bands, err := assessment.ListBands(ctx, camp.ScaleCode)
	if err != nil {
		return nil, err
	}
	bandLabels := make(map[string]string, len(bands))
	for _, b := range bands {
		bandLabels[b.Code] = b.Label
	}

	out := &ScopeProgression{
		ScopeNodeID: scopeNodeID, PeriodID: periodID, CampaignID: campaignID,
		Bands: make([]BandRow, 0, len(rows)),
	}
	for _, r := range rows {
		out.TotalScored += r.StudentCount
		out.GeneratedAt = r.SnapshotAt
		out.Bands = append(out.Bands, BandRow{
			BandCode: r.BandCode, BandOrdinal: r.BandOrdinal,
			BandLabel: bandLabels[r.BandCode], StudentCount: r.StudentCount,
		})
	}
	for i := range out.Bands {
		if out.TotalScored > 0 {
			out.Bands[i].Percentage = int32(float64(out.Bands[i].StudentCount) / float64(out.TotalScored) * 100)
		}
	}
	return out, nil
}

func DrilldownByScope(ctx context.Context, scopeNodeID, periodID, campaignID int64) (*Drilldown, error) {
	scopeProg, err := ComputeProgression(ctx, scopeNodeID, periodID, campaignID)
	if err != nil {
		return nil, err
	}

	parentID := scopeNodeID
	children, err := tenancy.ListChildren(ctx, &parentID)
	if err != nil {
		return nil, err
	}

	out := &Drilldown{Scope: *scopeProg, Children: make([]DrilldownChild, 0, len(children))}
	for _, child := range children {
		childProg, err := ComputeProgression(ctx, child.ID, periodID, campaignID)
		if err != nil {
			return nil, err
		}
		out.Children = append(out.Children, DrilldownChild{
			NodeID: child.ID, Code: child.Code, Label: child.Label, Level: child.Level,
			Bands: childProg.Bands, Total: childProg.TotalScored,
		})
	}
	return out, nil
}

type RefreshSummary struct {
	CampaignsScanned int
	ScopesRefreshed  int
	Errors           int
	DurationMs       int64
}

func RefreshAllOpenCampaigns(ctx context.Context) (*RefreshSummary, error) {
	start := time.Now()
	summary := &RefreshSummary{}

	campaigns, err := assessment.ListOpenCampaignsWithScores(ctx)
	if err != nil {
		return nil, fmt.Errorf("list open campaigns: %w", err)
	}
	summary.CampaignsScanned = len(campaigns)

	for _, camp := range campaigns {
		studentIDs, err := assessment.ListScoredStudentIDs(ctx, camp.ID)
		if err != nil {
			summary.Errors++
			continue
		}
		if len(studentIDs) == 0 {
			continue
		}

		schools, err := people.GetSchoolsForStudents(ctx, studentIDs)
		if err != nil {
			summary.Errors++
			continue
		}
		schoolSet := make(map[int64]struct{}, len(schools))
		for _, schoolID := range schools {
			schoolSet[schoolID] = struct{}{}
		}
		schoolIDs := make([]int64, 0, len(schoolSet))
		for id := range schoolSet {
			schoolIDs = append(schoolIDs, id)
		}

		ancestorIDs, err := tenancy.ListAncestorIDsForMany(ctx, schoolIDs)
		if err != nil {
			summary.Errors++
			continue
		}

		scopeSet := make(map[int64]struct{}, len(ancestorIDs)+len(schoolIDs))
		for _, id := range ancestorIDs {
			scopeSet[id] = struct{}{}
		}
		for _, id := range schoolIDs {
			scopeSet[id] = struct{}{}
		}

		for scopeID := range scopeSet {
			if summary.ScopesRefreshed >= RefreshAllOpenCampaignsLimit {
				summary.DurationMs = time.Since(start).Milliseconds()
				return summary, nil
			}
			if _, err := RefreshSnapshot(ctx, scopeID, camp.PeriodID, camp.ID); err != nil {
				summary.Errors++
				continue
			}
			summary.ScopesRefreshed++
		}
	}

	summary.DurationMs = time.Since(start).Milliseconds()
	return summary, nil
}

func DrilldownByScopeViaSnapshots(ctx context.Context, scopeNodeID, periodID, campaignID int64) (*Drilldown, error) {
	scopeProg, err := GetSnapshot(ctx, scopeNodeID, periodID, campaignID)
	if err != nil {
		return nil, err
	}

	parentID := scopeNodeID
	children, err := tenancy.ListChildren(ctx, &parentID)
	if err != nil {
		return nil, err
	}

	out := &Drilldown{Scope: *scopeProg, Children: make([]DrilldownChild, 0, len(children))}
	for _, child := range children {
		childProg, err := GetSnapshot(ctx, child.ID, periodID, campaignID)
		if err != nil {
			return nil, err
		}
		out.Children = append(out.Children, DrilldownChild{
			NodeID: child.ID, Code: child.Code, Label: child.Label, Level: child.Level,
			Bands: childProg.Bands, Total: childProg.TotalScored,
		})
	}
	return out, nil
}
