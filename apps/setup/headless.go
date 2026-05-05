package setup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"encore.app/apps/auth"
	"encore.app/apps/tenancy"
	"encore.app/internal/seed/countries"
)

func RunHeadless(ctx context.Context) error {
	if os.Getenv("SCHOOLRISE_HEADLESS") != "1" {
		return nil
	}

	state, err := loadState(ctx)
	if err != nil {
		return err
	}

	if state.SetupCompleteAt {
		fmt.Fprintln(os.Stderr, "schoolrise: SCHOOLRISE_HEADLESS=1 ignored — setup already complete")

		return nil
	}

	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	password := os.Getenv("ADMIN_PASSWORD")
	fullName := envOr("ADMIN_FULL_NAME", "Initial Administrator")

	if email == "" || password == "" {
		return errors.New("schoolrise: SCHOOLRISE_HEADLESS=1 requires ADMIN_EMAIL and ADMIN_PASSWORD")
	}

	user, err := auth.CreateUser(ctx, auth.CreateUserParams{
		Email:              email,
		Password:           password,
		FullName:           fullName,
		Role:               "admin",
		MustChangePassword: false,
	})
	if err != nil && !errors.Is(err, auth.ErrEmailAlreadyExists) {
		return fmt.Errorf("create admin: %w", err)
	}

	if user == nil {
		existing, lookupErr := auth.GetUserByEmail(ctx, email)
		if lookupErr != nil {
			return fmt.Errorf("lookup existing admin: %w", lookupErr)
		}

		user = existing
	}

	if err := auth.AssignRole(ctx, user.ID, "admin", nil); err != nil {
		return fmt.Errorf("assign admin role: %w", err)
	}

	payload, _ := json.Marshal(map[string]any{"user_id": user.ID, "email": user.Email})
	if err := markStepComplete(ctx, "admin", payload); err != nil {
		return err
	}

	instance := envOr("INSTANCE_NAME", "SchoolRise")
	locale := envOr("DEFAULT_LOCALE", "en")
	baseURL := envOr("BASE_URL", "http://localhost:3000")
	tz := envOr("TIME_ZONE", "UTC")

	system := map[string]string{
		"instance_name":  instance,
		"default_locale": locale,
		"base_url":       baseURL,
		"time_zone":      tz,
	}

	for k, v := range system {
		raw, _ := json.Marshal(v)
		if err := upsertSystemSetting(ctx, k, raw); err != nil {
			return err
		}
	}

	if err := markStepComplete(ctx, "system", mustJSON(system)); err != nil {
		return err
	}

	if err := applyHeadlessLevels(ctx); err != nil {
		return err
	}

	if err := markStepSkipped(ctx, "schools"); err != nil {
		return err
	}

	if err := markStepSkipped(ctx, "integrations"); err != nil {
		return err
	}

	if err := markStepSkipped(ctx, "smtp"); err != nil {
		return err
	}

	if err := queries.MarkSetupComplete(ctx); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "schoolrise: SCHOOLRISE_HEADLESS=1 install complete")

	return nil
}

func applyHeadlessLevels(ctx context.Context) error {
	packCode := strings.TrimSpace(os.Getenv("COUNTRY_PACK"))

	if packCode != "" {
		pack, err := countries.Get(packCode)
		if err != nil {
			return fmt.Errorf("country pack %q: %w", packCode, err)
		}

		defs := make([]tenancy.LevelDef, 0, len(pack.Levels))
		for _, l := range pack.Levels {
			defs = append(defs, tenancy.LevelDef{
				Code:        l.Code,
				Label:       l.Label,
				ParentLevel: l.Parent,
				Depth:       l.Depth,
				SortOrder:   l.Sort,
			})
		}

		if len(defs) > 0 {
			if err := tenancy.ApplyLevels(ctx, defs); err != nil {
				return err
			}
		}

		if err := markStepComplete(ctx, "levels", mustJSON(defs)); err != nil {
			return err
		}

		return nil
	}

	if err := markStepComplete(ctx, "levels", []byte("{}")); err != nil {
		return err
	}

	return nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}

	return fallback
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)

	return b
}
