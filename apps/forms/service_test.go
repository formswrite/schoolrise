package forms_test

import (
	"context"
	"testing"

	"encore.app/apps/forms"
)

func TestServiceHealth(t *testing.T) {
	t.Parallel()

	svc := &forms.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "forms" {
		t.Errorf("Service = %q, want %q", resp.Service, "forms")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
