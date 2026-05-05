package ai

import (
	"context"
	"errors"
	"strconv"
	"time"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

func currentUserID() int64 {
	uid, _ := encauth.UserID()
	id, _ := strconv.ParseInt(string(uid), 10, 64)
	return id
}

type SuggestItemsRequest struct {
	Topic       string `json:"topic"`
	ScaleCode   string `json:"scale_code"`
	NiveauLabel string `json:"niveau_label"`
	Count       int    `json:"count"`
}

type SuggestItemsResponse struct {
	Items []SuggestedItem `json:"items"`
}

//encore:api auth method=POST path=/v1/ai/suggest-items
func (s *Service) SuggestItemsAPI(ctx context.Context, req *SuggestItemsRequest) (*SuggestItemsResponse, error) {
	items, err := SuggestItems(ctx, SuggestItemsParams{
		Topic: req.Topic, ScaleCode: req.ScaleCode, NiveauLabel: req.NiveauLabel,
		Count: req.Count, UserID: currentUserID(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &SuggestItemsResponse{Items: items}, nil
}

type DraftRubricRequest struct {
	QuestionTitle string   `json:"question_title"`
	ScaleCode     string   `json:"scale_code"`
	BandCodes     []string `json:"band_codes"`
}

type DraftRubricResponse struct {
	Bands []RubricBand `json:"bands"`
}

//encore:api auth method=POST path=/v1/ai/draft-rubric
func (s *Service) DraftRubricAPI(ctx context.Context, req *DraftRubricRequest) (*DraftRubricResponse, error) {
	bands, err := DraftRubric(ctx, DraftRubricParams{
		QuestionTitle: req.QuestionTitle, ScaleCode: req.ScaleCode,
		BandCodes: req.BandCodes, UserID: currentUserID(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &DraftRubricResponse{Bands: bands}, nil
}

type GradeEssayRequest struct {
	QuestionTitle string       `json:"question_title"`
	StudentAnswer string       `json:"student_answer"`
	Rubric        []RubricBand `json:"rubric"`
}

//encore:api auth method=POST path=/v1/ai/grade-essay
func (s *Service) GradeEssayAPI(ctx context.Context, req *GradeEssayRequest) (*GradeEssayResult, error) {
	res, err := GradeEssay(ctx, GradeEssayParams{
		QuestionTitle: req.QuestionTitle, StudentAnswer: req.StudentAnswer,
		Rubric: req.Rubric, UserID: currentUserID(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return res, nil
}

type GenerateDistractorsRequest struct {
	QuestionTitle string `json:"question_title"`
	CorrectAnswer string `json:"correct_answer"`
	Count         int    `json:"count"`
}

type GenerateDistractorsResponse struct {
	Distractors []string `json:"distractors"`
}

//encore:api auth method=POST path=/v1/ai/generate-distractors
func (s *Service) GenerateDistractorsAPI(ctx context.Context, req *GenerateDistractorsRequest) (*GenerateDistractorsResponse, error) {
	d, err := GenerateDistractors(ctx, GenerateDistractorsParams{
		QuestionTitle: req.QuestionTitle, CorrectAnswer: req.CorrectAnswer,
		Count: req.Count, UserID: currentUserID(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &GenerateDistractorsResponse{Distractors: d}, nil
}

type ProviderStatusResponse struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

//encore:api auth method=GET path=/v1/ai/provider
func (s *Service) ProviderStatusAPI(ctx context.Context) (*ProviderStatusResponse, error) {
	return &ProviderStatusResponse{Provider: providerName(), Model: providerModel()}, nil
}

type JobDTO struct {
	ID             int64      `json:"id"`
	Kind           string     `json:"kind"`
	Model          string     `json:"model"`
	Status         string     `json:"status"`
	PromptSummary  string     `json:"prompt_summary"`
	RequestTokens  int32      `json:"request_tokens"`
	ResponseTokens int32      `json:"response_tokens"`
	LatencyMs      int32      `json:"latency_ms"`
	Error          string     `json:"error,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type ListJobsRequest struct {
	Limit  int32 `query:"limit"`
	Offset int32 `query:"offset"`
}

type ListJobsResponse struct {
	Jobs []JobDTO `json:"jobs"`
}

//encore:api auth method=GET path=/v1/ai/jobs
func (s *Service) ListJobsAPI(ctx context.Context, req *ListJobsRequest) (*ListJobsResponse, error) {
	rows, err := ListRecentJobs(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, apierr.WrapInternal("ai", err)
	}
	out := make([]JobDTO, 0, len(rows))
	for _, j := range rows {
		out = append(out, JobDTO{
			ID: j.ID, Kind: j.Kind, Model: j.Model, Status: j.Status,
			PromptSummary: j.PromptSummary, RequestTokens: j.RequestTokens,
			ResponseTokens: j.ResponseTokens, LatencyMs: j.LatencyMs, Error: j.Error,
			CreatedAt: j.CreatedAt, CompletedAt: j.CompletedAt,
		})
	}
	return &ListJobsResponse{Jobs: out}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	case errors.Is(err, ErrNoProvider):
		return &errs.Error{Code: errs.FailedPrecondition, Message: err.Error()}
	}
	return apierr.WrapInternal("ai", err)
}
