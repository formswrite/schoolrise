package progression_test

import (
	"context"
	"testing"

	"encore.app/apps/progression"
)

func TestServiceHealth(t *testing.T) {

	svc := &progression.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "progression" {
		t.Errorf("Service = %q, want %q", resp.Service, "progression")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
