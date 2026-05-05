package assessment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/assessment"
)

func uniqueIDs(t *testing.T) (formID, formVersionID, periodID, scopeID int64) {
	t.Helper()
	now := time.Now().UnixNano()
	return now, now + 1, now + 2, now + 3
}

func TestListScales_SeedSeesBothScales(t *testing.T) {
	ctx := context.Background()
	scales, err := assessment.ListScales(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	have := map[string]bool{}
	for _, s := range scales {
		have[s.Code] = true
	}
	if !have[assessment.ScaleFrench] || !have[assessment.ScaleMaths] {
		t.Fatalf("missing scales: %v", have)
	}
}

func TestListBands_FrenchHas5Bands(t *testing.T) {
	ctx := context.Background()
	bands, err := assessment.ListBands(ctx, assessment.ScaleFrench)
	if err != nil {
		t.Fatalf("bands: %v", err)
	}
	if len(bands) != 5 {
		t.Fatalf("len=%d, want 5", len(bands))
	}
	if bands[0].Code != "debutant" || bands[4].Code != "histoire" {
		t.Fatalf("order wrong: %v", bands)
	}
}

func TestCreateCampaign_RejectsUnknownScale(t *testing.T) {
	ctx := context.Background()
	formID, formVersionID, periodID, scopeID := uniqueIDs(t)
	_, err := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "x", ScaleCode: "fake_scale",
		FormID: formID, FormVersionID: formVersionID,
		PeriodID: periodID, ScopeNodeID: scopeID,
	})
	if !errors.Is(err, assessment.ErrInvalidScale) {
		t.Fatalf("err=%v, want ErrInvalidScale", err)
	}
}

func TestCreateCampaign_Success(t *testing.T) {
	ctx := context.Background()
	formID, formVersionID, periodID, scopeID := uniqueIDs(t)
	c, err := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "French Q1", ScaleCode: assessment.ScaleFrench,
		FormID: formID, FormVersionID: formVersionID,
		PeriodID: periodID, ScopeNodeID: scopeID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.Status != assessment.StatusDraft {
		t.Fatalf("status=%q", c.Status)
	}
}

func TestAssignStudents_GeneratesUniqueTokensAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	formID, fvID, periodID, scopeID := uniqueIDs(t)
	c, err := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "Assign", ScaleCode: assessment.ScaleMaths,
		FormID: formID, FormVersionID: fvID, PeriodID: periodID, ScopeNodeID: scopeID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	students := []int64{101, 102, 103}
	res, err := assessment.AssignStudents(ctx, assessment.AssignParams{CampaignID: c.ID, StudentIDs: students})
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	if len(res.Created) != 3 {
		t.Fatalf("created=%d, want 3", len(res.Created))
	}

	tokens := map[string]bool{}
	for _, a := range res.Created {
		if tokens[a.AccessToken] {
			t.Fatalf("duplicate token: %s", a.AccessToken)
		}
		tokens[a.AccessToken] = true
		if len(a.AccessToken) < 16 {
			t.Fatalf("token too short: %q", a.AccessToken)
		}
	}

	res2, err := assessment.AssignStudents(ctx, assessment.AssignParams{CampaignID: c.ID, StudentIDs: students})
	if err != nil {
		t.Fatalf("re-assign: %v", err)
	}
	if len(res2.Created) != 0 || res2.Existing != 3 {
		t.Fatalf("idempotency failed: created=%d existing=%d", len(res2.Created), res2.Existing)
	}
}

func TestSubmitResponse_OnlyWhenOpen(t *testing.T) {
	ctx := context.Background()
	formID, fvID, periodID, scopeID := uniqueIDs(t)
	c, _ := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "Open?", ScaleCode: assessment.ScaleFrench,
		FormID: formID, FormVersionID: fvID, PeriodID: periodID, ScopeNodeID: scopeID,
	})
	res, err := assessment.AssignStudents(ctx, assessment.AssignParams{CampaignID: c.ID, StudentIDs: []int64{200}})
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	token := res.Created[0].AccessToken

	if _, err := assessment.SubmitResponse(ctx, assessment.SubmitParams{
		AccessToken: token, RawScore: 50, Payload: map[string]any{"answers": []int{1, 2}},
	}); !errors.Is(err, assessment.ErrCampaignNotOpen) {
		t.Fatalf("err=%v, want ErrCampaignNotOpen", err)
	}

	if _, err := assessment.OpenCampaign(ctx, c.ID); err != nil {
		t.Fatalf("open: %v", err)
	}

	out, err := assessment.SubmitResponse(ctx, assessment.SubmitParams{
		AccessToken: token, RawScore: 50, Payload: map[string]any{"answers": []int{1, 2}},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if out.Score.RawScore != 50 || out.Score.BandCode != "mots" {
		t.Fatalf("score wrong: %+v", out.Score)
	}
	if out.Assignment.SubmittedAt == nil {
		t.Fatalf("submitted_at not set")
	}

	if _, err := assessment.SubmitResponse(ctx, assessment.SubmitParams{
		AccessToken: token, RawScore: 60,
	}); !errors.Is(err, assessment.ErrAlreadySubmitted) {
		t.Fatalf("err=%v, want ErrAlreadySubmitted", err)
	}
}

func TestSubmitResponse_BandBoundaries(t *testing.T) {
	ctx := context.Background()
	formID, fvID, periodID, scopeID := uniqueIDs(t)
	c, _ := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "Bands", ScaleCode: assessment.ScaleFrench,
		FormID: formID, FormVersionID: fvID, PeriodID: periodID, ScopeNodeID: scopeID,
	})
	if _, err := assessment.OpenCampaign(ctx, c.ID); err != nil {
		t.Fatalf("open: %v", err)
	}

	cases := []struct {
		score int32
		band  string
	}{
		{0, "debutant"},
		{19, "debutant"},
		{20, "lettres"},
		{40, "mots"},
		{60, "paragraphe"},
		{80, "histoire"},
		{100, "histoire"},
	}
	studentBase := time.Now().UnixNano()
	for i, tc := range cases {
		res, err := assessment.AssignStudents(ctx, assessment.AssignParams{
			CampaignID: c.ID, StudentIDs: []int64{studentBase + int64(i)},
		})
		if err != nil {
			t.Fatalf("assign %d: %v", i, err)
		}
		out, err := assessment.SubmitResponse(ctx, assessment.SubmitParams{
			AccessToken: res.Created[0].AccessToken, RawScore: tc.score,
		})
		if err != nil {
			t.Fatalf("submit score=%d: %v", tc.score, err)
		}
		if out.Score.BandCode != tc.band {
			t.Fatalf("score=%d → band=%q, want %q", tc.score, out.Score.BandCode, tc.band)
		}
	}
}

func TestSubmitResponse_ScoreOutOfRangeRejected(t *testing.T) {
	ctx := context.Background()
	formID, fvID, periodID, scopeID := uniqueIDs(t)
	c, _ := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "OOR", ScaleCode: assessment.ScaleFrench,
		FormID: formID, FormVersionID: fvID, PeriodID: periodID, ScopeNodeID: scopeID,
	})
	_, _ = assessment.OpenCampaign(ctx, c.ID)
	res, _ := assessment.AssignStudents(ctx, assessment.AssignParams{CampaignID: c.ID, StudentIDs: []int64{500}})

	if _, err := assessment.SubmitResponse(ctx, assessment.SubmitParams{
		AccessToken: res.Created[0].AccessToken, RawScore: 150,
	}); !errors.Is(err, assessment.ErrScoreOutOfRange) {
		t.Fatalf("err=%v, want ErrScoreOutOfRange", err)
	}
}

func TestGetAssignmentByToken_Unknown(t *testing.T) {
	ctx := context.Background()
	if _, err := assessment.GetAssignmentByToken(ctx, "totally-fake-token-xxxxx"); !errors.Is(err, assessment.ErrAssignmentNotFound) {
		t.Fatalf("err=%v, want ErrAssignmentNotFound", err)
	}
}
