package notifications

import (
	"context"
	"fmt"
	"html"
	"strings"
)

type AssignmentLinkParams struct {
	ToEmail       string
	ToName        string
	CampaignTitle string
	AccessURL     string
	StudentID     int64
	CampaignID    int64
	AccessToken   string
}

func SendAssignmentLink(ctx context.Context, p AssignmentLinkParams) (*Email, error) {
	subject := fmt.Sprintf("Your assessment: %s", p.CampaignTitle)
	greeting := "Hello"
	if p.ToName != "" {
		greeting = "Hello " + p.ToName
	}

	htmlBody := buildAssignmentHTML(greeting, p.CampaignTitle, p.AccessURL)
	textBody := fmt.Sprintf(
		"%s,\n\nYou've been assigned the assessment \"%s\" on SchoolRise.\n\nOpen this link to start: %s\n\nThis is a one-time link. Do not share it.\n",
		greeting, p.CampaignTitle, p.AccessURL,
	)

	return EnqueueAndSend(ctx, EnqueueParams{
		Kind: KindAssignmentLink, ToEmail: p.ToEmail, ToName: p.ToName,
		Subject: subject, BodyHTML: htmlBody, BodyText: textBody,
		Metadata: map[string]any{
			"student_id":   p.StudentID,
			"campaign_id":  p.CampaignID,
			"access_token": p.AccessToken,
		},
	})
}

type ImportSummaryParams struct {
	ToEmail   string
	ToName    string
	JobID     int64
	Kind      string
	Total     int
	Succeeded int
	Failed    int
}

func SendImportSummary(ctx context.Context, p ImportSummaryParams) (*Email, error) {
	subject := fmt.Sprintf("Import #%d (%s) — %d/%d rows succeeded", p.JobID, p.Kind, p.Succeeded, p.Total)
	htmlBody := fmt.Sprintf(`
<div style="font-family: -apple-system, sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #060419;">Import job #%d completed</h2>
  <p>Kind: <strong>%s</strong></p>
  <ul style="line-height: 1.6;">
    <li>Total rows: <strong>%d</strong></li>
    <li>Succeeded: <strong style="color: #16a34a;">%d</strong></li>
    <li>Failed: <strong style="color: #dc2626;">%d</strong></li>
  </ul>
</div>`, p.JobID, html.EscapeString(p.Kind), p.Total, p.Succeeded, p.Failed)
	textBody := fmt.Sprintf("Import job #%d (%s) completed.\nTotal: %d  Succeeded: %d  Failed: %d\n", p.JobID, p.Kind, p.Total, p.Succeeded, p.Failed)

	return EnqueueAndSend(ctx, EnqueueParams{
		Kind: KindImportSummary, ToEmail: p.ToEmail, ToName: p.ToName,
		Subject: subject, BodyHTML: htmlBody, BodyText: textBody,
		Metadata: map[string]any{"job_id": p.JobID},
	})
}

type CampaignClosedParams struct {
	ToEmail       string
	ToName        string
	CampaignTitle string
	CampaignID    int64
	TotalScored   int
}

func SendCampaignClosed(ctx context.Context, p CampaignClosedParams) (*Email, error) {
	subject := fmt.Sprintf("Campaign closed: %s", p.CampaignTitle)
	htmlBody := fmt.Sprintf(`
<div style="font-family: -apple-system, sans-serif; max-width: 600px; margin: 0 auto; padding: 24px;">
  <h2 style="color: #060419;">Campaign closed</h2>
  <p><strong>%s</strong> has been closed.</p>
  <p>Final scored count: <strong>%d</strong></p>
</div>`, html.EscapeString(p.CampaignTitle), p.TotalScored)
	textBody := fmt.Sprintf("Campaign \"%s\" has been closed.\nFinal scored count: %d\n", p.CampaignTitle, p.TotalScored)

	return EnqueueAndSend(ctx, EnqueueParams{
		Kind: KindCampaignClosed, ToEmail: p.ToEmail, ToName: p.ToName,
		Subject: subject, BodyHTML: htmlBody, BodyText: textBody,
		Metadata: map[string]any{"campaign_id": p.CampaignID},
	})
}

func buildAssignmentHTML(greeting, campaignTitle, accessURL string) string {
	var sb strings.Builder
	sb.WriteString(`<div style="font-family: -apple-system, BlinkMacSystemFont, sans-serif; max-width: 600px; margin: 0 auto; padding: 32px; background: #f0ecff;">`)
	sb.WriteString(`<div style="background: white; border: 2px solid #0b0d2a; border-radius: 14px; padding: 32px;">`)
	sb.WriteString(fmt.Sprintf(`<p style="font-size: 16px; color: #060419;">%s,</p>`, html.EscapeString(greeting)))
	sb.WriteString(fmt.Sprintf(`<p style="font-size: 16px; color: #060419;">You've been assigned the assessment <strong>%s</strong> on SchoolRise.</p>`, html.EscapeString(campaignTitle)))
	sb.WriteString(`<div style="text-align: center; margin: 32px 0;">`)
	sb.WriteString(fmt.Sprintf(`<a href="%s" style="display: inline-block; background: #6439B5; color: white; text-decoration: none; padding: 14px 32px; border: 2px solid #0b0d2a; border-radius: 10px; font-weight: 600; font-size: 16px;">Start assessment</a>`, html.EscapeString(accessURL)))
	sb.WriteString(`</div>`)
	sb.WriteString(`<p style="font-size: 13px; color: #51545d;">Or copy this link into your browser:<br>`)
	sb.WriteString(fmt.Sprintf(`<code style="word-break: break-all;">%s</code></p>`, html.EscapeString(accessURL)))
	sb.WriteString(`<p style="font-size: 12px; color: #51545d; margin-top: 24px;">This link is unique to you and can only be used once. Do not share it.</p>`)
	sb.WriteString(`</div></div>`)
	return sb.String()
}
