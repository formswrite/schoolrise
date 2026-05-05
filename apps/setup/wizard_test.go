package setup_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"encore.dev/beta/errs"

	"encore.app/apps/setup"
)

var emailCounter atomic.Uint64

func uniqueAdminEmail(t *testing.T) string {
	t.Helper()

	n := emailCounter.Add(1)

	return fmt.Sprintf("setup-%s-%d@local.test", strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")), n)
}

func TestUnlock_RejectsEmptyToken(t *testing.T) {
	t.Parallel()

	_, err := newService(t).UnlockAPI(context.Background(), &setup.UnlockRequest{InstallToken: ""})
	if err == nil {
		t.Fatal("expected error on empty token")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestUnlock_RejectsWrongToken(t *testing.T) {
	t.Parallel()

	_, err := newService(t).UnlockAPI(context.Background(), &setup.UnlockRequest{InstallToken: strings.Repeat("x", 32)})
	if err == nil {
		t.Fatal("expected error on wrong token")
	}

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) {
		t.Fatalf("expected *errs.Error, got %T", err)
	}

	if errsErr.Code != errs.Unauthenticated && errsErr.Code != errs.PermissionDenied && errsErr.Code != errs.ResourceExhausted {
		t.Errorf("expected Unauthenticated/PermissionDenied/ResourceExhausted, got %v", errsErr.Code)
	}
}

func TestCreateAdmin_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	_, err := newService(t).CreateAdminAPI(context.Background(), &setup.CreateAdminRequest{
		SessionToken: "",
		Email:        uniqueAdminEmail(t),
		FullName:     "X",
		Password:     "Pass1234!",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSaveSystemSettings_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SaveSystemSettingsAPI(context.Background(), &setup.SystemSettingsRequest{
		SessionToken:  "",
		InstanceName:  "X",
		DefaultLocale: "en",
		BaseURL:       "http://x",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSetLevels_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SetLevelsAPI(context.Background(), &setup.SetLevelsRequest{
		SessionToken: "",
		Levels:       []setup.SetLevelsLevel{{Code: "x", Label: "X", Parent: "", Depth: 0, Sort: 0}},
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestImportSchools_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	_, err := newService(t).ImportSchoolsAPI(context.Background(), &setup.ImportSchoolsRequest{
		SessionToken: "",
		CSV:          "parent_code,level,code,label\n",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSkipSchools_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SkipSchoolsAPI(context.Background(), &setup.SkipSchoolsRequest{SessionToken: ""})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSaveIntegrations_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SaveIntegrationsAPI(context.Background(), &setup.IntegrationsRequest{SessionToken: ""})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSkipIntegrations_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SkipIntegrationsAPI(context.Background(), &setup.SkipIntegrationsRequest{SessionToken: ""})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSaveSMTP_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SaveSMTPAPI(context.Background(), &setup.SMTPRequest{
		SessionToken: "",
		Host:         "smtp.example.com",
		Port:         587,
		FromAddress:  "x@example.com",
	})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSkipSMTP_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).SkipSMTPAPI(context.Background(), &setup.SkipSMTPRequest{SessionToken: ""})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestFinalize_RejectsMissingSession(t *testing.T) {
	t.Parallel()

	err := newService(t).FinalizeAPI(context.Background(), &setup.FinalizeRequest{SessionToken: ""})

	var errsErr *errs.Error
	if !errors.As(err, &errsErr) || errsErr.Code != errs.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}
