package setup_test

import (
	"context"
	"testing"

	"encore.app/apps/setup"
)

func newService(t *testing.T) *setup.Service {
	t.Helper()
	return &setup.Service{}
}

func TestStatus_FreshInstallReturnsIncomplete(t *testing.T) {
	t.Parallel()

	resp, err := newService(t).StatusAPI(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}

	if resp.SetupComplete {
		t.Error("fresh setup should not be complete")
	}
}

func TestIsSetupComplete_FreshIsFalse(t *testing.T) {
	t.Parallel()

	complete, err := setup.IsSetupComplete(context.Background())
	if err != nil {
		t.Fatalf("IsSetupComplete: %v", err)
	}

	if complete {
		t.Error("fresh install should not be marked complete")
	}
}
