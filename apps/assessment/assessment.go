package assessment

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/assessment/dbassessment"
	"encore.app/apps/notifications"
	"encore.app/apps/people"
)

const (
	StatusDraft  = "draft"
	StatusOpen   = "open"
	StatusClosed = "closed"

	ScaleFrench = "french_5level"
	ScaleMaths  = "maths_5level"
)

var (
	ErrCampaignNotFound      = errors.New("assessment: campaign not found")
	ErrAssignmentNotFound    = errors.New("assessment: assignment not found")
	ErrScoreNotFound         = errors.New("assessment: score not found")
	ErrInvalidCampaignInput  = errors.New("assessment: invalid campaign input")
	ErrInvalidScale          = errors.New("assessment: scale not found")
	ErrCampaignNotOpen       = errors.New("assessment: campaign is not open")
	ErrAlreadySubmitted      = errors.New("assessment: assignment already submitted")
	ErrScoreOutOfRange       = errors.New("assessment: score must be between 0 and 100")
	ErrAssignmentExists      = errors.New("assessment: student already assigned to this campaign")
)

type Scale struct {
	Code  string
	Label string
}

type Band struct {
	ID         int64
	ScaleCode  string
	Ordinal    int32
	Code       string
	Label      string
	MinScore   int32
	MaxScore   int32
}

type Campaign struct {
	ID            int64
	PublicID      string
	Title         string
	ScaleCode     string
	FormID        int64
	FormVersionID int64
	PeriodID      int64
	ScopeNodeID   int64
	Status        string
	OpensAt       *time.Time
	ClosesAt      *time.Time
	CreatedBy     *int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Assignment struct {
	ID          int64
	CampaignID  int64
	StudentID   int64
	AccessToken string
	SubmittedAt *time.Time
	CreatedAt   time.Time
}

type Score struct {
	ID          int64
	ResponseID  int64
	CampaignID  int64
	StudentID   int64
	RawScore    int32
	BandCode    string
	BandOrdinal int32
	FinalizedAt time.Time
}

func ListScales(ctx context.Context) ([]Scale, error) {
	rows, err := queries.ListScales(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Scale, 0, len(rows))
	for _, r := range rows {
		out = append(out, Scale{Code: r.Code, Label: r.Label})
	}
	return out, nil
}

func ListBands(ctx context.Context, scaleCode string) ([]Band, error) {
	rows, err := queries.ListBandsForScale(ctx, scaleCode)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		if _, err := queries.GetScale(ctx, scaleCode); errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidScale
		}
	}
	out := make([]Band, 0, len(rows))
	for _, r := range rows {
		out = append(out, Band{
			ID: r.ID, ScaleCode: r.ScaleCode, Ordinal: r.Ordinal, Code: r.Code, Label: r.Label,
			MinScore: r.MinScore, MaxScore: r.MaxScore,
		})
	}
	return out, nil
}

type CreateCampaignParams struct {
	Title         string
	ScaleCode     string
	FormID        int64
	FormVersionID int64
	PeriodID      int64
	ScopeNodeID   int64
	CreatedBy     int64
}

func CreateCampaign(ctx context.Context, p CreateCampaignParams) (*Campaign, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" || p.FormID <= 0 || p.FormVersionID <= 0 || p.PeriodID <= 0 || p.ScopeNodeID <= 0 {
		return nil, ErrInvalidCampaignInput
	}

	if _, err := queries.GetScale(ctx, p.ScaleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidScale
		}
		return nil, err
	}

	publicID, err := newToken(12)
	if err != nil {
		return nil, err
	}

	row, err := queries.CreateCampaign(ctx, dbassessment.CreateCampaignParams{
		PublicID: publicID, Title: title, ScaleCode: p.ScaleCode,
		FormID: p.FormID, FormVersionID: p.FormVersionID,
		PeriodID: p.PeriodID, ScopeNodeID: p.ScopeNodeID,
		Status:    StatusDraft,
		CreatedBy: nullableInt64(p.CreatedBy),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrInvalidCampaignInput
		}
		return nil, err
	}
	return campaignFromRow(row), nil
}

func GetCampaignByID(ctx context.Context, id int64) (*Campaign, error) {
	row, err := queries.GetCampaignByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCampaignNotFound
	}
	if err != nil {
		return nil, err
	}
	return campaignFromRow(row), nil
}

func ListCampaignsByScope(ctx context.Context, scopeNodeID int64) ([]*Campaign, error) {
	rows, err := queries.ListCampaignsByScope(ctx, scopeNodeID)
	if err != nil {
		return nil, err
	}
	out := make([]*Campaign, 0, len(rows))
	for _, r := range rows {
		out = append(out, campaignFromRow(r))
	}
	return out, nil
}

func OpenCampaign(ctx context.Context, id int64) (*Campaign, error) {
	if _, err := GetCampaignByID(ctx, id); err != nil {
		return nil, err
	}
	row, err := queries.UpdateCampaignStatus(ctx, dbassessment.UpdateCampaignStatusParams{ID: id, Status: StatusOpen})
	if err != nil {
		return nil, err
	}
	return campaignFromRow(row), nil
}

func CloseCampaign(ctx context.Context, id int64) (*Campaign, error) {
	if _, err := GetCampaignByID(ctx, id); err != nil {
		return nil, err
	}
	row, err := queries.UpdateCampaignStatus(ctx, dbassessment.UpdateCampaignStatusParams{ID: id, Status: StatusClosed})
	if err != nil {
		return nil, err
	}
	return campaignFromRow(row), nil
}

type AssignParams struct {
	CampaignID int64
	StudentIDs []int64
	NotifyByEmail bool
}

type AssignResult struct {
	Created       []*Assignment
	Existing      int
	EmailsSent    int
	EmailsSkipped int
}

func AssignStudents(ctx context.Context, p AssignParams) (*AssignResult, error) {
	camp, err := GetCampaignByID(ctx, p.CampaignID)
	if err != nil {
		return nil, err
	}

	res := &AssignResult{Created: []*Assignment{}}
	for _, sid := range p.StudentIDs {
		token, err := newToken(24)
		if err != nil {
			return nil, err
		}
		row, err := queries.CreateAssignment(ctx, dbassessment.CreateAssignmentParams{
			CampaignID: p.CampaignID, StudentID: sid, AccessToken: token,
		})
		if err != nil {
			if isUniqueViolation(err) {
				res.Existing++
				continue
			}
			return nil, err
		}
		assignment := assignmentFromRow(row)
		res.Created = append(res.Created, assignment)

		if p.NotifyByEmail {
			if sent := dispatchAssignmentEmail(ctx, camp, assignment); sent {
				res.EmailsSent++
			} else {
				res.EmailsSkipped++
			}
		}
	}
	return res, nil
}

func dispatchAssignmentEmail(ctx context.Context, camp *Campaign, assignment *Assignment) bool {
	student, err := people.GetStudentByID(ctx, assignment.StudentID)
	if err != nil {
		log.Printf("assessment: cannot lookup student %d for email: %v", assignment.StudentID, err)
		return false
	}
	person, err := people.GetPersonByID(ctx, student.PersonID)
	if err != nil {
		log.Printf("assessment: cannot lookup person for student %d: %v", assignment.StudentID, err)
		return false
	}
	if strings.TrimSpace(person.Email) == "" {
		return false
	}

	baseURL := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}

	if _, err := notifications.SendAssignmentLink(ctx, notifications.AssignmentLinkParams{
		ToEmail:       person.Email,
		ToName:        person.FullName,
		CampaignTitle: camp.Title,
		AccessURL:     fmt.Sprintf("%s/r/%s", baseURL, assignment.AccessToken),
		StudentID:     assignment.StudentID,
		CampaignID:    camp.ID,
		AccessToken:   assignment.AccessToken,
	}); err != nil {
		log.Printf("assessment: assignment email failed for student %d: %v", assignment.StudentID, err)
		return false
	}
	return true
}

func GetAssignmentByToken(ctx context.Context, token string) (*Assignment, error) {
	row, err := queries.GetAssignmentByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAssignmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return assignmentFromRow(row), nil
}

func ListAssignments(ctx context.Context, campaignID int64) ([]*Assignment, error) {
	rows, err := queries.ListAssignmentsByCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	out := make([]*Assignment, 0, len(rows))
	for _, r := range rows {
		out = append(out, assignmentFromRow(r))
	}
	return out, nil
}

type SubmitParams struct {
	AccessToken string
	Payload     map[string]any
	RawScore    int32
}

type SubmitResult struct {
	Assignment *Assignment
	Score      *Score
}

func SubmitResponse(ctx context.Context, p SubmitParams) (*SubmitResult, error) {
	assignment, err := GetAssignmentByToken(ctx, p.AccessToken)
	if err != nil {
		return nil, err
	}
	if assignment.SubmittedAt != nil {
		return nil, ErrAlreadySubmitted
	}

	camp, err := GetCampaignByID(ctx, assignment.CampaignID)
	if err != nil {
		return nil, err
	}
	if camp.Status != StatusOpen {
		return nil, ErrCampaignNotOpen
	}

	if p.RawScore < 0 || p.RawScore > 100 {
		return nil, ErrScoreOutOfRange
	}

	payload, err := json.Marshal(p.Payload)
	if err != nil {
		return nil, err
	}

	respRow, err := queries.CreateResponse(ctx, dbassessment.CreateResponseParams{
		AssignmentID: assignment.ID,
		CampaignID:   assignment.CampaignID,
		StudentID:    assignment.StudentID,
		Payload:      payload,
	})
	if err != nil {
		return nil, err
	}

	bandRow, err := queries.BandForScore(ctx, dbassessment.BandForScoreParams{
		ScaleCode: camp.ScaleCode, Column2: p.RawScore,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScoreOutOfRange
		}
		return nil, err
	}

	scoreRow, err := queries.CreateScore(ctx, dbassessment.CreateScoreParams{
		ResponseID: respRow.ID, CampaignID: camp.ID, StudentID: assignment.StudentID,
		RawScore: p.RawScore, BandCode: bandRow.Code, BandOrdinal: bandRow.Ordinal,
	})
	if err != nil {
		return nil, err
	}

	updated, err := queries.MarkAssignmentSubmitted(ctx, assignment.ID)
	if err != nil {
		return nil, err
	}

	return &SubmitResult{
		Assignment: assignmentFromRow(updated),
		Score:      scoreFromRow(scoreRow),
	}, nil
}

func ListScores(ctx context.Context, campaignID int64) ([]*Score, error) {
	rows, err := queries.ListScoresByCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	out := make([]*Score, 0, len(rows))
	for _, r := range rows {
		out = append(out, scoreFromRow(r))
	}
	return out, nil
}

type OpenCampaignWithScores struct {
	ID          int64
	PeriodID    int64
	ScopeNodeID int64
	ScaleCode   string
}

func ListOpenCampaignsWithScores(ctx context.Context) ([]OpenCampaignWithScores, error) {
	rows, err := queries.ListOpenCampaignsWithScores(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]OpenCampaignWithScores, 0, len(rows))
	for _, r := range rows {
		out = append(out, OpenCampaignWithScores{
			ID: r.ID, PeriodID: r.PeriodID, ScopeNodeID: r.ScopeNodeID, ScaleCode: r.ScaleCode,
		})
	}
	return out, nil
}

func ListScoredStudentIDs(ctx context.Context, campaignID int64) ([]int64, error) {
	return queries.ListScoredStudentIDs(ctx, campaignID)
}

func campaignFromRow(r dbassessment.Campaign) *Campaign {
	c := &Campaign{
		ID: r.ID, PublicID: r.PublicID, Title: r.Title, ScaleCode: r.ScaleCode,
		FormID: r.FormID, FormVersionID: r.FormVersionID,
		PeriodID: r.PeriodID, ScopeNodeID: r.ScopeNodeID,
		Status:    r.Status,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.OpensAt.Valid {
		t := r.OpensAt.Time
		c.OpensAt = &t
	}
	if r.ClosesAt.Valid {
		t := r.ClosesAt.Time
		c.ClosesAt = &t
	}
	if r.CreatedBy.Valid {
		v := r.CreatedBy.Int64
		c.CreatedBy = &v
	}
	return c
}

func assignmentFromRow(r dbassessment.Assignment) *Assignment {
	a := &Assignment{
		ID: r.ID, CampaignID: r.CampaignID, StudentID: r.StudentID,
		AccessToken: r.AccessToken, CreatedAt: r.CreatedAt,
	}
	if r.SubmittedAt.Valid {
		t := r.SubmittedAt.Time
		a.SubmittedAt = &t
	}
	return a
}

func scoreFromRow(r dbassessment.Score) *Score {
	return &Score{
		ID: r.ID, ResponseID: r.ResponseID, CampaignID: r.CampaignID, StudentID: r.StudentID,
		RawScore: r.RawScore, BandCode: r.BandCode, BandOrdinal: r.BandOrdinal,
		FinalizedAt: r.FinalizedAt,
	}
}

func newToken(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func nullableInt64(v int64) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
