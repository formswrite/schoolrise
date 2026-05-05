package auth_test

import (
	"context"
	"testing"

	"encore.app/apps/auth"
)

func TestServiceHealth(t *testing.T) {

	svc := &auth.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "auth" {
		t.Errorf("Service = %q, want %q", resp.Service, "auth")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
