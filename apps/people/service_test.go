package people_test

import (
	"context"
	"testing"

	"encore.app/apps/people"
)

func TestServiceHealth(t *testing.T) {

	svc := &people.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "people" {
		t.Errorf("Service = %q, want %q", resp.Service, "people")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
