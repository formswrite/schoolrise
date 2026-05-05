package notifications

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type EmailDTO struct {
	ID         int64      `json:"id"`
	Kind       string     `json:"kind"`
	ToEmail    string     `json:"to_email"`
	ToName     string     `json:"to_name,omitempty"`
	Subject    string     `json:"subject"`
	Status     string     `json:"status"`
	Attempts   int32      `json:"attempts"`
	LastError  string     `json:"last_error,omitempty"`
	ProviderID string     `json:"provider_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
}

func toDTO(e *Email) EmailDTO {
	return EmailDTO{
		ID: e.ID, Kind: e.Kind, ToEmail: e.ToEmail, ToName: e.ToName,
		Subject: e.Subject, Status: e.Status, Attempts: e.Attempts,
		LastError: e.LastError, ProviderID: e.ProviderID,
		CreatedAt: e.CreatedAt, SentAt: e.SentAt,
	}
}

type ListEmailsRequest struct {
	Limit  int32 `query:"limit"`
	Offset int32 `query:"offset"`
}

type ListEmailsResponse struct {
	Emails []EmailDTO `json:"emails"`
}

//encore:api auth method=GET path=/v1/notifications/outbox
func (s *Service) ListEmailsAPI(ctx context.Context, req *ListEmailsRequest) (*ListEmailsResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := ListRecent(ctx, limit, req.Offset)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]EmailDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDTO(r))
	}
	return &ListEmailsResponse{Emails: out}, nil
}

type ProviderStatusResponse struct {
	Provider string `json:"provider"`
	From     string `json:"from"`
}

//encore:api auth method=GET path=/v1/notifications/provider
func (s *Service) ProviderStatusAPI(ctx context.Context) (*ProviderStatusResponse, error) {
	p := getProvider()
	name := "none"
	if p != nil {
		name = p.Name()
	}
	return &ProviderStatusResponse{Provider: name, From: getSender()}, nil
}

type SendTestEmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

//encore:api auth method=POST path=/v1/notifications/test
func (s *Service) SendTestEmailAPI(ctx context.Context, req *SendTestEmailRequest) (*EmailDTO, error) {
	subject := req.Subject
	if subject == "" {
		subject = "SchoolRise — test email"
	}
	body := req.Body
	if body == "" {
		body = "<p>This is a test email from SchoolRise.</p>"
	}
	e, err := EnqueueAndSend(ctx, EnqueueParams{
		Kind: KindTest, ToEmail: req.To, Subject: subject, BodyHTML: body,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := toDTO(e)
	return &out, nil
}

//encore:api auth method=POST path=/v1/notifications/process
func (s *Service) ProcessOutboxAPI(ctx context.Context) (*ProcessResult, error) {
	res, err := ProcessOutbox(ctx, 50)
	if err != nil {
		return nil, internal(err)
	}
	return res, nil
}

func internal(err error) error { return apierr.WrapInternal("notifications", err) }

func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrEmptySubject):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	}
	return internal(err)
}
