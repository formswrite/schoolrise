package ai

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"encore.dev/rlog"

	bamlclient "encore.app/apps/ai/baml_client"
	bamltypes "encore.app/apps/ai/baml_client/types"
	"encore.app/apps/ai/dbai"
	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
)

const (
	KindSuggestItems        = "suggest_items"
	KindDraftRubric         = "draft_rubric"
	KindGradeEssay          = "grade_essay"
	KindGenerateDistractors = "generate_distractors"
)

var (
	ErrInvalidInput = errors.New("ai: invalid input")
	ErrNoProvider   = errors.New("ai: no LLM provider configured")
	ErrLLM          = errors.New("ai: LLM call failed")
)

type SuggestedItem struct {
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
}

type SuggestItemsParams struct {
	Topic       string
	ScaleCode   string
	NiveauLabel string
	Count       int
	UserID      int64
}

func SuggestItems(ctx context.Context, p SuggestItemsParams) ([]SuggestedItem, error) {
	if strings.TrimSpace(p.Topic) == "" {
		return nil, ErrInvalidInput
	}
	if p.Count <= 0 || p.Count > 20 {
		p.Count = 5
	}

	out, err := callBAML(ctx, KindSuggestItems, p.UserID, p.Topic, func(opts ...bamlclient.CallOptionFunc) (any, error) {
		return bamlclient.SuggestItems(ctx, p.Topic, p.ScaleCode, p.NiveauLabel, int64(p.Count), opts...)
	})
	if err != nil {
		return nil, err
	}
	parsed := out.(bamltypes.SuggestedItemList)

	items := make([]SuggestedItem, 0, len(parsed.Items))
	for _, it := range parsed.Items {
		var opts []string
		if it.Options != nil {
			opts = *it.Options
		}
		items = append(items, SuggestedItem{
			Type:     string(it.Type),
			Title:    it.Title,
			Required: it.Required,
			Options:  opts,
		})
	}
	return items, nil
}

type RubricBand struct {
	BandCode    string `json:"band_code"`
	Description string `json:"description"`
	MinScore    int32  `json:"min_score"`
	MaxScore    int32  `json:"max_score"`
}

type DraftRubricParams struct {
	QuestionTitle string
	ScaleCode     string
	BandCodes     []string
	UserID        int64
}

func DraftRubric(ctx context.Context, p DraftRubricParams) ([]RubricBand, error) {
	if strings.TrimSpace(p.QuestionTitle) == "" {
		return nil, ErrInvalidInput
	}
	if len(p.BandCodes) == 0 {
		return nil, ErrInvalidInput
	}

	out, err := callBAML(ctx, KindDraftRubric, p.UserID, p.QuestionTitle, func(opts ...bamlclient.CallOptionFunc) (any, error) {
		return bamlclient.DraftRubric(ctx, p.QuestionTitle, p.ScaleCode, p.BandCodes, opts...)
	})
	if err != nil {
		return nil, err
	}
	parsed := out.(bamltypes.RubricBandList)

	bands := make([]RubricBand, 0, len(parsed.Bands))
	for _, b := range parsed.Bands {
		bands = append(bands, RubricBand{
			BandCode:    b.Band_code,
			Description: b.Description,
			MinScore:    int32(b.Min_score),
			MaxScore:    int32(b.Max_score),
		})
	}
	return bands, nil
}

type GradeEssayResult struct {
	RawScore  int32  `json:"raw_score"`
	BandCode  string `json:"band_code"`
	Reasoning string `json:"reasoning"`
}

type GradeEssayParams struct {
	QuestionTitle string
	StudentAnswer string
	Rubric        []RubricBand
	UserID        int64
}

func GradeEssay(ctx context.Context, p GradeEssayParams) (*GradeEssayResult, error) {
	if strings.TrimSpace(p.StudentAnswer) == "" || len(p.Rubric) == 0 {
		return nil, ErrInvalidInput
	}

	bamlRubric := make([]bamltypes.RubricBand, 0, len(p.Rubric))
	for _, b := range p.Rubric {
		bamlRubric = append(bamlRubric, bamltypes.RubricBand{
			Band_code:   b.BandCode,
			Description: b.Description,
			Min_score:   int64(b.MinScore),
			Max_score:   int64(b.MaxScore),
		})
	}

	out, err := callBAML(ctx, KindGradeEssay, p.UserID, p.QuestionTitle, func(opts ...bamlclient.CallOptionFunc) (any, error) {
		return bamlclient.GradeEssay(ctx, p.QuestionTitle, p.StudentAnswer, bamlRubric, opts...)
	})
	if err != nil {
		return nil, err
	}
	parsed := out.(bamltypes.GradeResult)

	score := int32(parsed.Raw_score.Value)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return &GradeEssayResult{
		RawScore:  score,
		BandCode:  parsed.Band_code,
		Reasoning: parsed.Reasoning,
	}, nil
}

type GenerateDistractorsParams struct {
	QuestionTitle string
	CorrectAnswer string
	Count         int
	UserID        int64
}

func GenerateDistractors(ctx context.Context, p GenerateDistractorsParams) ([]string, error) {
	if strings.TrimSpace(p.QuestionTitle) == "" || strings.TrimSpace(p.CorrectAnswer) == "" {
		return nil, ErrInvalidInput
	}
	if p.Count <= 0 || p.Count > 10 {
		p.Count = 3
	}

	out, err := callBAML(ctx, KindGenerateDistractors, p.UserID, p.QuestionTitle, func(opts ...bamlclient.CallOptionFunc) (any, error) {
		return bamlclient.GenerateDistractors(ctx, p.QuestionTitle, p.CorrectAnswer, int64(p.Count), opts...)
	})
	if err != nil {
		return nil, err
	}
	parsed := out.(bamltypes.DistractorList)
	return parsed.Distractors, nil
}

func callBAML(
	ctx context.Context,
	kind string,
	userID int64,
	summary string,
	do func(opts ...bamlclient.CallOptionFunc) (any, error),
) (any, error) {
	if !providerEnabled() {
		return nil, ErrNoProvider
	}

	model := providerModel()

	jobRow, err := queries.CreateJob(ctx, dbai.CreateJobParams{
		Kind:          kind,
		Model:         model,
		PromptSummary: truncate(summary, 200),
		RequestedBy:   sql.NullInt64{Int64: userID, Valid: userID > 0},
		Metadata:      []byte("{}"),
	})
	if err != nil {
		return nil, err
	}

	collector, cerr := bamlclient.NewCollector(fmt.Sprintf("ai-job-%d", jobRow.ID))
	if cerr != nil {
		return nil, fmt.Errorf("%w: collector: %v", ErrLLM, cerr)
	}

	opts := []bamlclient.CallOptionFunc{bamlclient.WithCollector(collector)}
	if reg := currentRegistry(); reg != nil {
		opts = append(opts, bamlclient.WithClientRegistry(reg))
	}

	start := time.Now()
	result, callErr := do(opts...)
	elapsed := int32(time.Since(start) / time.Millisecond)

	if callErr != nil {
		if failErr := queries.FailJob(ctx, dbai.FailJobParams{
			ID:        jobRow.ID,
			Error:     sql.NullString{String: callErr.Error(), Valid: true},
			LatencyMs: elapsed,
		}); failErr != nil {
			rlog.Warn("ai: failed to record job failure", "job_id", jobRow.ID, "err", failErr)
		}
		return nil, fmt.Errorf("%w: %v", ErrLLM, callErr)
	}

	reqTokens, respTokens := tokenUsage(collector)

	if completeErr := queries.CompleteJob(ctx, dbai.CompleteJobParams{
		ID:             jobRow.ID,
		RequestTokens:  reqTokens,
		ResponseTokens: respTokens,
		LatencyMs:      elapsed,
	}); completeErr != nil {
		rlog.Warn("ai: failed to record job completion", "job_id", jobRow.ID, "err", completeErr)
	}
	return result, nil
}

func tokenUsage(c baml.Collector) (req, resp int32) {
	if c == nil {
		return 0, 0
	}
	last, err := c.Last()
	if err != nil || last == nil {
		return 0, 0
	}
	usage, err := last.Usage()
	if err != nil || usage == nil {
		return 0, 0
	}
	if v, err := usage.InputTokens(); err == nil {
		req = int32(v)
	}
	if v, err := usage.OutputTokens(); err == nil {
		resp = int32(v)
	}
	return req, resp
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

type Job struct {
	ID             int64
	Kind           string
	Model          string
	Status         string
	PromptSummary  string
	RequestTokens  int32
	ResponseTokens int32
	LatencyMs      int32
	Error          string
	RequestedBy    *int64
	CreatedAt      time.Time
	CompletedAt    *time.Time
}


func ListRecentJobs(ctx context.Context, limit, offset int32) ([]*Job, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := queries.ListRecentJobs(ctx, dbai.ListRecentJobsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	out := make([]*Job, 0, len(rows))
	for _, r := range rows {
		j := &Job{
			ID:             r.ID,
			Kind:           r.Kind,
			Model:          r.Model,
			Status:         r.Status,
			PromptSummary:  r.PromptSummary,
			RequestTokens:  r.RequestTokens,
			ResponseTokens: r.ResponseTokens,
			LatencyMs:      r.LatencyMs,
			CreatedAt:      r.CreatedAt,
		}
		if r.Error.Valid {
			j.Error = r.Error.String
		}
		if r.RequestedBy.Valid {
			v := r.RequestedBy.Int64
			j.RequestedBy = &v
		}
		if r.CompletedAt.Valid {
			t := r.CompletedAt.Time
			j.CompletedAt = &t
		}
		out = append(out, j)
	}
	return out, nil
}
