package ai_test

import (
	"context"
	"testing"

	"encore.app/apps/ai"
)

func TestServiceHealth(t *testing.T) {

	svc := &ai.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "ai" {
		t.Errorf("Service = %q, want %q", resp.Service, "ai")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
