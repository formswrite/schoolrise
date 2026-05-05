package academics_test

import (
	"context"
	"testing"

	"encore.app/apps/academics"
)

func TestServiceHealth(t *testing.T) {

	svc := &academics.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "academics" {
		t.Errorf("Service = %q, want %q", resp.Service, "academics")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
