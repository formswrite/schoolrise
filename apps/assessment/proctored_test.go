package assessment_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"encore.app/apps/assessment"
)

func newOpenCampaign(t *testing.T) *assessment.Campaign {
	t.Helper()
	ctx := context.Background()
	formID, fvID, periodID, scopeID := uniqueIDs(t)
	c, err := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "Proctored Test", ScaleCode: assessment.ScaleFrench,
		FormID: formID, FormVersionID: fvID, PeriodID: periodID, ScopeNodeID: scopeID,
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	if _, err := assessment.OpenCampaign(ctx, c.ID); err != nil {
		t.Fatalf("open: %v", err)
	}
	return c
}

func score(v int32) *int32 { return &v }

func TestSubmitProctored_HappyPath_SingleEntry(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	studentID := time.Now().UnixNano()

	res, err := assessment.SubmitProctoredScores(ctx, c.ID, 1234, []assessment.ProctoredEntry{
		{StudentID: studentID, RawScore: score(65), Mode: assessment.EntryModeProctoredScore},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Created != 1 || res.Updated != 0 || len(res.Errors) != 0 {
		t.Fatalf("counts wrong: %+v", res)
	}

	roster, err := assessment.ListGradingRoster(ctx, c.ID, []int64{studentID})
	if err != nil {
		t.Fatalf("roster: %v", err)
	}
	if len(roster) != 1 {
		t.Fatalf("roster len = %d", len(roster))
	}
	row := roster[0]
	if !row.HasScore || row.RawScore == nil || *row.RawScore != 65 || row.BandCode != "paragraphe" {
		t.Fatalf("row wrong: %+v", row)
	}
	if row.EntryMode != assessment.EntryModeProctoredScore {
		t.Fatalf("entry_mode = %q, want %q", row.EntryMode, assessment.EntryModeProctoredScore)
	}
	if row.ProctoredByUserID == nil || *row.ProctoredByUserID != 1234 {
		t.Fatalf("proctor user not recorded: %+v", row.ProctoredByUserID)
	}
}

func TestSubmitProctored_BatchOfFive_AllCreated(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	base := time.Now().UnixNano()

	entries := make([]assessment.ProctoredEntry, 5)
	for i := 0; i < 5; i++ {
		entries[i] = assessment.ProctoredEntry{
			StudentID: base + int64(i),
			RawScore:  score(int32(20 + i*15)),
			Mode:      assessment.EntryModeProctoredScore,
		}
	}

	res, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, entries)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Created != 5 || res.Updated != 0 || len(res.Errors) != 0 {
		t.Fatalf("counts wrong: %+v", res)
	}
}

func TestSubmitProctored_ReentryUpdates(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	studentID := time.Now().UnixNano()

	if _, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: studentID, RawScore: score(30), Mode: assessment.EntryModeProctoredScore},
	}); err != nil {
		t.Fatalf("first: %v", err)
	}

	res, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: studentID, RawScore: score(85), Mode: assessment.EntryModeProctoredScore},
	})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if res.Created != 0 || res.Updated != 1 {
		t.Fatalf("counts wrong: %+v", res)
	}

	roster, _ := assessment.ListGradingRoster(ctx, c.ID, []int64{studentID})
	if roster[0].RawScore == nil || *roster[0].RawScore != 85 {
		t.Fatalf("score not updated: %+v", roster[0])
	}
	if roster[0].BandCode != "histoire" {
		t.Fatalf("band not recomputed: %q", roster[0].BandCode)
	}
}

func TestSubmitProctored_RejectsScoreOutOfRange(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)

	res, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: time.Now().UnixNano(), RawScore: score(150), Mode: assessment.EntryModeProctoredScore},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Created != 0 || len(res.Errors) != 1 {
		t.Fatalf("expected one error, got %+v", res)
	}
}

func TestSubmitProctored_PartialFailureContinues(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	base := time.Now().UnixNano()

	res, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: base + 1, RawScore: score(50), Mode: assessment.EntryModeProctoredScore},
		{StudentID: base + 2, RawScore: score(200), Mode: assessment.EntryModeProctoredScore},
		{StudentID: base + 3, RawScore: score(75), Mode: assessment.EntryModeProctoredScore},
		{StudentID: base + 4, RawScore: nil, Mode: assessment.EntryModeProctoredScore},
		{StudentID: base + 5, RawScore: score(10), Mode: assessment.EntryModeProctoredScore},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Created != 3 || len(res.Errors) != 2 {
		t.Fatalf("counts: %+v", res)
	}
}

func TestSubmitProctored_RejectsClosedCampaign(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	if _, err := assessment.CloseCampaign(ctx, c.ID); err != nil {
		t.Fatalf("close: %v", err)
	}

	_, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: time.Now().UnixNano(), RawScore: score(50), Mode: assessment.EntryModeProctoredScore},
	})
	if !errors.Is(err, assessment.ErrCampaignNotOpen) {
		t.Fatalf("err = %v, want ErrCampaignNotOpen", err)
	}
}

func TestSubmitProctored_RejectsMissingCampaign(t *testing.T) {
	ctx := context.Background()
	_, err := assessment.SubmitProctoredScores(ctx, 999999999, 1, []assessment.ProctoredEntry{
		{StudentID: 1, RawScore: score(50), Mode: assessment.EntryModeProctoredScore},
	})
	if !errors.Is(err, assessment.ErrCampaignNotFound) {
		t.Fatalf("err = %v, want ErrCampaignNotFound", err)
	}
}

func TestSubmitProctored_RejectsEmptyBatch(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	_, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, nil)
	if !errors.Is(err, assessment.ErrProctoredEmptyBatch) {
		t.Fatalf("err = %v, want ErrProctoredEmptyBatch", err)
	}
}

func TestSubmitProctored_AnswersModeStoresPayload(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	studentID := time.Now().UnixNano()
	answers := json.RawMessage(`{"q1":"answer one","q2":"answer two"}`)

	res, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: studentID, RawScore: score(70), Mode: assessment.EntryModeProctoredAnswers, Answers: answers},
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if res.Created != 1 {
		t.Fatalf("created = %d", res.Created)
	}

	roster, _ := assessment.ListGradingRoster(ctx, c.ID, []int64{studentID})
	if roster[0].EntryMode != assessment.EntryModeProctoredAnswers {
		t.Fatalf("entry_mode = %q", roster[0].EntryMode)
	}
}

func TestListGradingRoster_LeftJoinShowsUnscored(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	base := time.Now().UnixNano()
	students := []int64{base + 1, base + 2, base + 3, base + 4}

	if _, err := assessment.SubmitProctoredScores(ctx, c.ID, 1, []assessment.ProctoredEntry{
		{StudentID: students[0], RawScore: score(40), Mode: assessment.EntryModeProctoredScore},
		{StudentID: students[2], RawScore: score(85), Mode: assessment.EntryModeProctoredScore},
	}); err != nil {
		t.Fatalf("submit: %v", err)
	}

	roster, err := assessment.ListGradingRoster(ctx, c.ID, students)
	if err != nil {
		t.Fatalf("roster: %v", err)
	}
	if len(roster) != 4 {
		t.Fatalf("len = %d, want 4", len(roster))
	}

	scored := 0
	for _, r := range roster {
		if r.HasScore {
			scored++
		}
	}
	if scored != 2 {
		t.Fatalf("scored = %d, want 2", scored)
	}
}

func TestListGradingRoster_EmptyStudentList(t *testing.T) {
	ctx := context.Background()
	c := newOpenCampaign(t)
	roster, err := assessment.ListGradingRoster(ctx, c.ID, nil)
	if err != nil {
		t.Fatalf("roster: %v", err)
	}
	if len(roster) != 0 {
		t.Fatalf("expected empty, got %d", len(roster))
	}
}
