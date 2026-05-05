package setup

import (
	"context"
	"encoding/json"
	"strings"

	"encore.dev/beta/errs"
)

type SystemSettingsRequest struct {
	SessionToken   string
	InstanceName   string
	DefaultLocale  string
	BaseURL        string
	TimeZone       string
}

//encore:api public method=POST path=/v1/setup/system
func (s *Service) SaveSystemSettingsAPI(ctx context.Context, req *SystemSettingsRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	name := strings.TrimSpace(req.InstanceName)
	locale := strings.TrimSpace(req.DefaultLocale)
	baseURL := strings.TrimSpace(req.BaseURL)
	tz := strings.TrimSpace(req.TimeZone)

	if name == "" || locale == "" || baseURL == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "instance name, default locale and base URL required"}
	}

	settings := map[string]string{
		"instance_name":  name,
		"default_locale": locale,
		"base_url":       baseURL,
		"time_zone":      tz,
	}

	for k, v := range settings {
		raw, _ := json.Marshal(v)
		if err := upsertSystemSetting(ctx, k, raw); err != nil {
			return &errs.Error{Code: errs.Internal, Message: "could not save system setting"}
		}
	}

	payload, _ := json.Marshal(settings)
	if err := markStepComplete(ctx, "system", payload); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark progress"}
	}

	return nil
}
