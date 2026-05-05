package setup

import (
	"context"
	"encoding/json"
	"strings"

	"encore.dev/beta/errs"
)

type SMTPRequest struct {
	SessionToken string
	Host         string
	Port         int
	Username     string
	Password     string
	UseTLS       bool
	FromAddress  string
}

//encore:api public method=POST path=/v1/setup/smtp
func (s *Service) SaveSMTPAPI(ctx context.Context, req *SMTPRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	if strings.TrimSpace(req.Host) == "" || req.Port <= 0 || strings.TrimSpace(req.FromAddress) == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "host, port and from address required"}
	}

	cfg := map[string]any{
		"host":         req.Host,
		"port":         req.Port,
		"username":     req.Username,
		"password":     req.Password,
		"use_tls":      req.UseTLS,
		"from_address": req.FromAddress,
	}

	raw, _ := json.Marshal(cfg)
	if err := upsertSystemSetting(ctx, "smtp_config", raw); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not save SMTP config"}
	}

	if err := markStepComplete(ctx, "smtp", []byte("{}")); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark progress"}
	}

	return nil
}

type SkipSMTPRequest struct {
	SessionToken string
}

//encore:api public method=POST path=/v1/setup/smtp/skip
func (s *Service) SkipSMTPAPI(ctx context.Context, req *SkipSMTPRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	if err := markStepSkipped(ctx, "smtp"); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark step skipped"}
	}

	return nil
}
