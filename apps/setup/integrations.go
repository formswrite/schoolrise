package setup

import (
	"context"
	"encoding/json"
	"strings"

	"encore.dev/beta/errs"
)

type IntegrationsRequest struct {
	SessionToken    string
	ResendAPIKey    string
	EmailFrom       string
	EmailFromName   string
	OpenAIAPIKey    string
	OpenAIModel     string
	AnthropicAPIKey string
	S3Endpoint      string
	S3Bucket        string
	S3AccessKey     string
	S3SecretKey     string
	S3Region        string
}

//encore:api public method=POST path=/v1/setup/integrations
func (s *Service) SaveIntegrationsAPI(ctx context.Context, req *IntegrationsRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	settings := map[string]string{
		"resend_api_key":    req.ResendAPIKey,
		"email_from":        req.EmailFrom,
		"email_from_name":   req.EmailFromName,
		"openai_api_key":    req.OpenAIAPIKey,
		"openai_model":      req.OpenAIModel,
		"anthropic_api_key": req.AnthropicAPIKey,
		"s3_endpoint":       req.S3Endpoint,
		"s3_bucket":         req.S3Bucket,
		"s3_access_key":     req.S3AccessKey,
		"s3_secret_key":     req.S3SecretKey,
		"s3_region":         req.S3Region,
	}

	for k, v := range settings {
		if strings.TrimSpace(v) == "" {
			continue
		}

		raw, _ := json.Marshal(v)
		if err := upsertSystemSetting(ctx, k, raw); err != nil {
			return &errs.Error{Code: errs.Internal, Message: "could not save integration setting"}
		}
	}

	if err := markStepComplete(ctx, "integrations", []byte("{}")); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark progress"}
	}

	return nil
}

type SkipIntegrationsRequest struct {
	SessionToken string
}

//encore:api public method=POST path=/v1/setup/integrations/skip
func (s *Service) SkipIntegrationsAPI(ctx context.Context, req *SkipIntegrationsRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	if err := markStepSkipped(ctx, "integrations"); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark step skipped"}
	}

	return nil
}
