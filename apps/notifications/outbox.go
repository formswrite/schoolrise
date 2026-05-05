package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"encore.app/apps/notifications/dbnotifications"
)

const (
	KindAssignmentLink = "assignment_link"
	KindImportSummary  = "import_summary"
	KindCampaignClosed = "campaign_closed"
	KindTest           = "test"

	StatusPending = "pending"
	StatusSending = "sending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusDropped = "dropped"

	maxAttempts = 5
)

var (
	ErrInvalidEmail = errors.New("notifications: invalid recipient email")
	ErrEmptySubject = errors.New("notifications: subject required")
)

type Email struct {
	ID         int64
	Kind       string
	ToEmail    string
	ToName     string
	Subject    string
	BodyHTML   string
	BodyText   string
	Status     string
	Attempts   int32
	LastError  string
	ProviderID string
	Metadata   map[string]any
	CreatedAt  time.Time
	SentAt     *time.Time
}

type EnqueueParams struct {
	Kind     string
	ToEmail  string
	ToName   string
	Subject  string
	BodyHTML string
	BodyText string
	Metadata map[string]any
}

func Enqueue(ctx context.Context, p EnqueueParams) (*Email, error) {
	to := strings.TrimSpace(p.ToEmail)
	if to == "" || !strings.Contains(to, "@") {
		return nil, ErrInvalidEmail
	}
	if strings.TrimSpace(p.Subject) == "" {
		return nil, ErrEmptySubject
	}

	meta, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, err
	}
	if string(meta) == "null" {
		meta = []byte("{}")
	}

	row, err := queries.EnqueueEmail(ctx, dbnotifications.EnqueueEmailParams{
		Kind:     p.Kind,
		ToEmail:  to,
		ToName:   sql.NullString{String: p.ToName, Valid: p.ToName != ""},
		Subject:  p.Subject,
		BodyHtml: p.BodyHTML,
		BodyText: p.BodyText,
		Metadata: meta,
	})
	if err != nil {
		return nil, err
	}
	return emailFromRow(row), nil
}

func SendNow(ctx context.Context, id int64) (*Email, error) {
	row, err := queries.GetEmailByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status != StatusPending && row.Status != StatusFailed {
		return emailFromRow(row), nil
	}
	if row.Attempts >= maxAttempts {
		_ = queries.MarkFailed(ctx, dbnotifications.MarkFailedParams{
			ID: id, Status: StatusDropped,
			LastError: sql.NullString{String: "max attempts exceeded", Valid: true},
		})
		return nil, errors.New("notifications: max attempts exceeded")
	}

	if err := queries.MarkSending(ctx, id); err != nil {
		return nil, err
	}

	provider := getProvider()
	if provider == nil {
		_ = queries.MarkFailed(ctx, dbnotifications.MarkFailedParams{
			ID: id, Status: StatusFailed,
			LastError: sql.NullString{String: "no provider configured", Valid: true},
		})
		return nil, errors.New("notifications: no provider configured")
	}

	toName := ""
	if row.ToName.Valid {
		toName = row.ToName.String
	}
	res, err := provider.Send(ctx, EmailRequest{
		From:    getSender(),
		To:      row.ToEmail,
		ToName:  toName,
		Subject: row.Subject,
		HTML:    row.BodyHtml,
		Text:    row.BodyText,
	})
	if err != nil {
		_ = queries.MarkFailed(ctx, dbnotifications.MarkFailedParams{
			ID: id, Status: StatusFailed,
			LastError: sql.NullString{String: err.Error(), Valid: true},
		})
		return nil, err
	}

	if err := queries.MarkSent(ctx, dbnotifications.MarkSentParams{
		ID: id, ProviderID: sql.NullString{String: res.ProviderID, Valid: true},
	}); err != nil {
		return nil, err
	}

	updated, err := queries.GetEmailByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return emailFromRow(updated), nil
}

func EnqueueAndSend(ctx context.Context, p EnqueueParams) (*Email, error) {
	row, err := Enqueue(ctx, p)
	if err != nil {
		return nil, err
	}
	return SendNow(ctx, row.ID)
}

type ProcessResult struct {
	Attempted int
	Sent      int
	Failed    int
}

func ProcessOutbox(ctx context.Context, batchSize int32) (*ProcessResult, error) {
	if batchSize <= 0 {
		batchSize = 25
	}
	rows, err := queries.ListPending(ctx, batchSize)
	if err != nil {
		return nil, err
	}
	res := &ProcessResult{}
	for _, r := range rows {
		res.Attempted++
		if _, err := SendNow(ctx, r.ID); err != nil {
			res.Failed++
			continue
		}
		res.Sent++
	}
	return res, nil
}

func ListRecent(ctx context.Context, limit, offset int32) ([]*Email, error) {
	rows, err := queries.ListRecent(ctx, dbnotifications.ListRecentParams{
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*Email, 0, len(rows))
	for _, r := range rows {
		out = append(out, emailFromRow(r))
	}
	return out, nil
}

func emailFromRow(r dbnotifications.NotificationsOutbox) *Email {
	e := &Email{
		ID:        r.ID,
		Kind:      r.Kind,
		ToEmail:   r.ToEmail,
		Subject:   r.Subject,
		BodyHTML:  r.BodyHtml,
		BodyText:  r.BodyText,
		Status:    r.Status,
		Attempts:  r.Attempts,
		CreatedAt: r.CreatedAt,
	}
	if r.ToName.Valid {
		e.ToName = r.ToName.String
	}
	if r.LastError.Valid {
		e.LastError = r.LastError.String
	}
	if r.ProviderID.Valid {
		e.ProviderID = r.ProviderID.String
	}
	if r.SentAt.Valid {
		t := r.SentAt.Time
		e.SentAt = &t
	}
	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &e.Metadata)
	}
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	return e
}
