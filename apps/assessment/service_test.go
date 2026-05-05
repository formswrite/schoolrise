package assessment_test

import (
	"context"
	"testing"

	"encore.app/apps/assessment"
)

func TestServiceHealth(t *testing.T) {

	svc := &assessment.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "assessment" {
		t.Errorf("Service = %q, want %q", resp.Service, "assessment")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
