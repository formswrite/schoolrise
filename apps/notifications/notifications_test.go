package notifications_test

import (
	"context"
	"errors"
	"testing"

	"encore.app/apps/notifications"
)

func setupLogProvider(t *testing.T) *notifications.LogProvider {
	t.Helper()
	p := notifications.NewLogProvider()
	notifications.SetProvider(p)
	notifications.SetSender("test@local.test", "Test")
	return p
}

func TestEnqueue_RejectsInvalidEmail(t *testing.T) {
	ctx := context.Background()
	if _, err := notifications.Enqueue(ctx, notifications.EnqueueParams{
		Kind: "test", ToEmail: "not-an-email", Subject: "Hi", BodyHTML: "x",
	}); !errors.Is(err, notifications.ErrInvalidEmail) {
		t.Fatalf("err=%v, want ErrInvalidEmail", err)
	}
}

func TestEnqueue_RejectsEmptySubject(t *testing.T) {
	ctx := context.Background()
	if _, err := notifications.Enqueue(ctx, notifications.EnqueueParams{
		Kind: "test", ToEmail: "x@y.com", BodyHTML: "x",
	}); !errors.Is(err, notifications.ErrEmptySubject) {
		t.Fatalf("err=%v, want ErrEmptySubject", err)
	}
}

func TestEnqueue_PersistsAsPending(t *testing.T) {
	ctx := context.Background()
	e, err := notifications.Enqueue(ctx, notifications.EnqueueParams{
		Kind: "test", ToEmail: "alice@example.com", ToName: "Alice",
		Subject: "Hello", BodyHTML: "<p>hi</p>",
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if e.Status != notifications.StatusPending {
		t.Fatalf("status=%q, want pending", e.Status)
	}
	if e.Attempts != 0 {
		t.Fatalf("attempts=%d, want 0", e.Attempts)
	}
}

func TestSendNow_MarksSentViaLogProvider(t *testing.T) {
	ctx := context.Background()
	p := setupLogProvider(t)
	p.Reset()

	e, err := notifications.EnqueueAndSend(ctx, notifications.EnqueueParams{
		Kind: "test", ToEmail: "bob@example.com", ToName: "Bob",
		Subject: "Welcome", BodyHTML: "<p>Welcome</p>", BodyText: "Welcome",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if e.Status != notifications.StatusSent {
		t.Fatalf("status=%q, want sent", e.Status)
	}
	if e.SentAt == nil {
		t.Fatalf("sent_at not set")
	}
	if e.ProviderID == "" {
		t.Fatalf("provider_id not set")
	}

	sent := p.Sent()
	if len(sent) != 1 || sent[0].To != "bob@example.com" || sent[0].Subject != "Welcome" {
		t.Fatalf("provider got: %+v", sent)
	}
}

func TestProcessOutbox_DrainsPending(t *testing.T) {
	ctx := context.Background()
	p := setupLogProvider(t)
	p.Reset()

	for i := 0; i < 3; i++ {
		_, err := notifications.Enqueue(ctx, notifications.EnqueueParams{
			Kind: "test", ToEmail: "batch@example.com", Subject: "Batch", BodyHTML: "x",
		})
		if err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
	}
	res, err := notifications.ProcessOutbox(ctx, 100)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if res.Sent < 3 {
		t.Fatalf("sent=%d, want >=3 (got attempted=%d failed=%d)", res.Sent, res.Attempted, res.Failed)
	}
}

func TestSendAssignmentLink_BuildsBrandedHTML(t *testing.T) {
	ctx := context.Background()
	p := setupLogProvider(t)
	p.Reset()

	e, err := notifications.SendAssignmentLink(ctx, notifications.AssignmentLinkParams{
		ToEmail: "student@example.com", ToName: "Student",
		CampaignTitle: "French Q1", AccessURL: "http://localhost:3001/r/abc123",
		StudentID: 42, CampaignID: 7, AccessToken: "abc123",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if e.Status != notifications.StatusSent {
		t.Fatalf("status=%q, want sent", e.Status)
	}
	if e.Kind != notifications.KindAssignmentLink {
		t.Fatalf("kind=%q", e.Kind)
	}
	if e.Subject != "Your assessment: French Q1" {
		t.Fatalf("subject=%q", e.Subject)
	}

	sent := p.Sent()
	if len(sent) != 1 {
		t.Fatalf("provider sent %d emails", len(sent))
	}
	body := sent[0].HTML
	for _, want := range []string{"French Q1", "http://localhost:3001/r/abc123", "#6439B5", "Start assessment"} {
		if !contains(body, want) {
			t.Errorf("html missing %q", want)
		}
	}
	if e.Metadata["campaign_id"] != float64(7) {
		t.Fatalf("metadata.campaign_id wrong: %v", e.Metadata)
	}
}

func TestSendImportSummary(t *testing.T) {
	ctx := context.Background()
	p := setupLogProvider(t)
	p.Reset()

	e, err := notifications.SendImportSummary(ctx, notifications.ImportSummaryParams{
		ToEmail: "ops@example.com", JobID: 99, Kind: "students",
		Total: 100, Succeeded: 95, Failed: 5,
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if e.Kind != notifications.KindImportSummary {
		t.Fatalf("kind=%q", e.Kind)
	}
	if !contains(e.Subject, "95/100") {
		t.Fatalf("subject missing counts: %q", e.Subject)
	}
}

func TestSendNow_ProviderFailureMarksFailed(t *testing.T) {
	ctx := context.Background()
	notifications.SetProvider(failingProvider{})
	notifications.SetSender("test@local.test", "")
	defer notifications.SetProvider(notifications.NewLogProvider())

	e, err := notifications.Enqueue(ctx, notifications.EnqueueParams{
		Kind: "test", ToEmail: "fail@example.com", Subject: "X", BodyHTML: "x",
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if _, err := notifications.SendNow(ctx, e.ID); err == nil {
		t.Fatalf("expected provider error")
	}
}

type failingProvider struct{}

func (failingProvider) Send(_ context.Context, _ notifications.EmailRequest) (*notifications.ProviderResult, error) {
	return nil, errors.New("provider down")
}
func (failingProvider) Name() string { return "failing" }

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
