package enrollment_test

import (
	"context"
	"testing"

	"encore.app/apps/enrollment"
)

func TestServiceHealth(t *testing.T) {

	svc := &enrollment.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "enrollment" {
		t.Errorf("Service = %q, want %q", resp.Service, "enrollment")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
