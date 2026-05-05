package assessment

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type ScaleDTO struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type BandDTO struct {
	Ordinal  int32  `json:"ordinal"`
	Code     string `json:"code"`
	Label    string `json:"label"`
	MinScore int32  `json:"min_score"`
	MaxScore int32  `json:"max_score"`
}

type CampaignDTO struct {
	ID            int64      `json:"id"`
	PublicID      string     `json:"public_id"`
	Title         string     `json:"title"`
	ScaleCode     string     `json:"scale_code"`
	FormID        int64      `json:"form_id"`
	FormVersionID int64      `json:"form_version_id"`
	PeriodID      int64      `json:"period_id"`
	ScopeNodeID   int64      `json:"scope_node_id"`
	Status        string     `json:"status"`
	OpensAt       *time.Time `json:"opens_at,omitempty"`
	ClosesAt      *time.Time `json:"closes_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AssignmentDTO struct {
	ID          int64      `json:"id"`
	CampaignID  int64      `json:"campaign_id"`
	StudentID   int64      `json:"student_id"`
	AccessToken string     `json:"access_token"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ScoreDTO struct {
	ID          int64     `json:"id"`
	ResponseID  int64     `json:"response_id"`
	CampaignID  int64     `json:"campaign_id"`
	StudentID   int64     `json:"student_id"`
	RawScore    int32     `json:"raw_score"`
	BandCode    string    `json:"band_code"`
	BandOrdinal int32     `json:"band_ordinal"`
	FinalizedAt time.Time `json:"finalized_at"`
}

func campaignToDTO(c *Campaign) CampaignDTO {
	return CampaignDTO{
		ID: c.ID, PublicID: c.PublicID, Title: c.Title,
		ScaleCode: c.ScaleCode, FormID: c.FormID, FormVersionID: c.FormVersionID,
		PeriodID: c.PeriodID, ScopeNodeID: c.ScopeNodeID,
		Status: c.Status, OpensAt: c.OpensAt, ClosesAt: c.ClosesAt, CreatedAt: c.CreatedAt,
	}
}
func assignmentToDTO(a *Assignment) AssignmentDTO {
	return AssignmentDTO{
		ID: a.ID, CampaignID: a.CampaignID, StudentID: a.StudentID,
		AccessToken: a.AccessToken, SubmittedAt: a.SubmittedAt, CreatedAt: a.CreatedAt,
	}
}
func scoreToDTO(s *Score) ScoreDTO {
	return ScoreDTO{
		ID: s.ID, ResponseID: s.ResponseID, CampaignID: s.CampaignID, StudentID: s.StudentID,
		RawScore: s.RawScore, BandCode: s.BandCode, BandOrdinal: s.BandOrdinal, FinalizedAt: s.FinalizedAt,
	}
}

type ListScalesResponse struct {
	Scales []ScaleDTO `json:"scales"`
}

//encore:api auth method=GET path=/v1/scales
func (s *Service) ListScalesAPI(ctx context.Context) (*ListScalesResponse, error) {
	rows, err := ListScales(ctx)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]ScaleDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ScaleDTO{Code: r.Code, Label: r.Label})
	}
	return &ListScalesResponse{Scales: out}, nil
}

type ListBandsResponse struct {
	Bands []BandDTO `json:"bands"`
}

//encore:api auth method=GET path=/v1/scales/:code/bands
func (s *Service) ListBandsAPI(ctx context.Context, code string) (*ListBandsResponse, error) {
	rows, err := ListBands(ctx, code)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]BandDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, BandDTO{Ordinal: r.Ordinal, Code: r.Code, Label: r.Label, MinScore: r.MinScore, MaxScore: r.MaxScore})
	}
	return &ListBandsResponse{Bands: out}, nil
}

type CreateCampaignAPIRequest struct {
	Title         string `json:"title"`
	ScaleCode     string `json:"scale_code"`
	FormID        int64  `json:"form_id"`
	FormVersionID int64  `json:"form_version_id"`
	PeriodID      int64  `json:"period_id"`
	ScopeNodeID   int64  `json:"scope_node_id"`
}

//encore:api auth method=POST path=/v1/campaigns
func (s *Service) CreateCampaignAPI(ctx context.Context, req *CreateCampaignAPIRequest) (*CampaignDTO, error) {
	uid, _ := encauth.UserID()
	createdBy, _ := strconv.ParseInt(string(uid), 10, 64)
	c, err := CreateCampaign(ctx, CreateCampaignParams{
		Title: req.Title, ScaleCode: req.ScaleCode, FormID: req.FormID, FormVersionID: req.FormVersionID,
		PeriodID: req.PeriodID, ScopeNodeID: req.ScopeNodeID, CreatedBy: createdBy,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := campaignToDTO(c)
	return &out, nil
}

type ListCampaignsRequest struct {
	ScopeNodeID int64 `query:"scope_node_id"`
}

type ListCampaignsResponse struct {
	Campaigns []CampaignDTO `json:"campaigns"`
}

//encore:api auth method=GET path=/v1/campaigns
func (s *Service) ListCampaignsAPI(ctx context.Context, req *ListCampaignsRequest) (*ListCampaignsResponse, error) {
	if req.ScopeNodeID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "scope_node_id required"}
	}
	rows, err := ListCampaignsByScope(ctx, req.ScopeNodeID)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]CampaignDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, campaignToDTO(r))
	}
	return &ListCampaignsResponse{Campaigns: out}, nil
}

//encore:api auth method=GET path=/v1/campaigns/:id
func (s *Service) GetCampaignAPI(ctx context.Context, id int64) (*CampaignDTO, error) {
	c, err := GetCampaignByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	out := campaignToDTO(c)
	return &out, nil
}

//encore:api auth method=POST path=/v1/campaigns/:id/open
func (s *Service) OpenCampaignAPI(ctx context.Context, id int64) (*CampaignDTO, error) {
	c, err := OpenCampaign(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	out := campaignToDTO(c)
	return &out, nil
}

//encore:api auth method=POST path=/v1/campaigns/:id/close
func (s *Service) CloseCampaignAPI(ctx context.Context, id int64) (*CampaignDTO, error) {
	c, err := CloseCampaign(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	out := campaignToDTO(c)
	return &out, nil
}

type AssignAPIRequest struct {
	StudentIDs    []int64 `json:"student_ids"`
	NotifyByEmail bool    `json:"notify_by_email"`
}

type AssignAPIResponse struct {
	Created       []AssignmentDTO `json:"created"`
	Existing      int             `json:"existing"`
	EmailsSent    int             `json:"emails_sent"`
	EmailsSkipped int             `json:"emails_skipped"`
}

//encore:api auth method=POST path=/v1/campaigns/:id/assign
func (s *Service) AssignAPI(ctx context.Context, id int64, req *AssignAPIRequest) (*AssignAPIResponse, error) {
	res, err := AssignStudents(ctx, AssignParams{
		CampaignID: id, StudentIDs: req.StudentIDs, NotifyByEmail: req.NotifyByEmail,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]AssignmentDTO, 0, len(res.Created))
	for _, a := range res.Created {
		out = append(out, assignmentToDTO(a))
	}
	return &AssignAPIResponse{
		Created: out, Existing: res.Existing,
		EmailsSent: res.EmailsSent, EmailsSkipped: res.EmailsSkipped,
	}, nil
}

type ListAssignmentsResponse struct {
	Assignments []AssignmentDTO `json:"assignments"`
}

//encore:api auth method=GET path=/v1/campaigns/:id/assignments
func (s *Service) ListAssignmentsAPI(ctx context.Context, id int64) (*ListAssignmentsResponse, error) {
	rows, err := ListAssignments(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]AssignmentDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, assignmentToDTO(r))
	}
	return &ListAssignmentsResponse{Assignments: out}, nil
}

type SubmitAPIRequest struct {
	AccessToken string          `json:"access_token"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	RawScore    int32           `json:"raw_score"`
}

type SubmitAPIResponse struct {
	Assignment AssignmentDTO `json:"assignment"`
	Score      ScoreDTO      `json:"score"`
}

//encore:api public method=POST path=/v1/responses
func (s *Service) SubmitResponseAPI(ctx context.Context, req *SubmitAPIRequest) (*SubmitAPIResponse, error) {
	var payload map[string]any
	if len(req.Payload) > 0 && string(req.Payload) != "null" {
		_ = json.Unmarshal(req.Payload, &payload)
	}
	res, err := SubmitResponse(ctx, SubmitParams{
		AccessToken: req.AccessToken, Payload: payload, RawScore: req.RawScore,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &SubmitAPIResponse{Assignment: assignmentToDTO(res.Assignment), Score: scoreToDTO(res.Score)}, nil
}

type LookupAssignmentRequest struct {
	Token string `query:"token"`
}

type LookupAssignmentResponse struct {
	Assignment    AssignmentDTO `json:"assignment"`
	Campaign      CampaignDTO   `json:"campaign"`
	FormVersionID int64         `json:"form_version_id"`
}

//encore:api public method=GET path=/v1/responses/lookup
func (s *Service) LookupAssignmentAPI(ctx context.Context, req *LookupAssignmentRequest) (*LookupAssignmentResponse, error) {
	a, err := GetAssignmentByToken(ctx, req.Token)
	if err != nil {
		return nil, mapErr(err)
	}
	c, err := GetCampaignByID(ctx, a.CampaignID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &LookupAssignmentResponse{
		Assignment: assignmentToDTO(a), Campaign: campaignToDTO(c), FormVersionID: c.FormVersionID,
	}, nil
}

type ListScoresResponse struct {
	Scores []ScoreDTO `json:"scores"`
}

//encore:api auth method=GET path=/v1/campaigns/:id/scores
func (s *Service) ListScoresAPI(ctx context.Context, id int64) (*ListScoresResponse, error) {
	rows, err := ListScores(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]ScoreDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, scoreToDTO(r))
	}
	return &ListScoresResponse{Scores: out}, nil
}

func internal(err error) error { return apierr.WrapInternal("assessment", err) }

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrCampaignNotFound), errors.Is(err, ErrAssignmentNotFound), errors.Is(err, ErrScoreNotFound):
		return &errs.Error{Code: errs.NotFound, Message: err.Error()}
	case errors.Is(err, ErrInvalidCampaignInput), errors.Is(err, ErrInvalidScale),
		errors.Is(err, ErrCampaignNotOpen), errors.Is(err, ErrAlreadySubmitted),
		errors.Is(err, ErrScoreOutOfRange):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	case errors.Is(err, ErrAssignmentExists):
		return &errs.Error{Code: errs.AlreadyExists, Message: err.Error()}
	}
	return internal(err)
}
