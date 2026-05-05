package notifications_test

import (
	"context"
	"testing"

	"encore.app/apps/notifications"
)

func TestServiceHealth(t *testing.T) {

	svc := &notifications.Service{}

	resp, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}

	if resp == nil {
		t.Fatal("Health returned nil response")
	}

	if resp.Service != "notifications" {
		t.Errorf("Service = %q, want %q", resp.Service, "notifications")
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}
