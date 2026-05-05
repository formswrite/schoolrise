package seed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ProbeStatusOK      = "ok"
	ProbeStatusFail    = "fail"
	ProbeStatusSkipped = "skipped"
)

type ProbeResult struct {
	Service string
	Status  string
	Detail  string
}

type proberClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var defaultProbeClient proberClient = &http.Client{Timeout: 5 * time.Second}

func ProbeIntegrations(ctx context.Context) []ProbeResult {
	return []ProbeResult{
		probeOpenAI(ctx, defaultProbeClient, strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))),
		probeResend(ctx, defaultProbeClient, strings.TrimSpace(os.Getenv("RESEND_API_KEY"))),
	}
}

func LogProbeResults(results []ProbeResult, logf func(format string, args ...any)) {
	for _, r := range results {
		switch r.Status {
		case ProbeStatusOK:
			logf("schoolrise: probe %s=ok", r.Service)
		case ProbeStatusSkipped:
			logf("schoolrise: probe %s=skipped (%s)", r.Service, r.Detail)
		default:
			logf("schoolrise: probe %s=FAIL (%s)", r.Service, r.Detail)
		}
	}
}

func probeOpenAI(ctx context.Context, client proberClient, key string) ProbeResult {
	if isPlaceholderAIKey(key) {
		return ProbeResult{Service: "openai", Status: ProbeStatusSkipped, Detail: "placeholder or empty key"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return ProbeResult{Service: "openai", Status: ProbeStatusFail, Detail: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+key)
	return runProbe(client, req, "openai")
}

func probeResend(ctx context.Context, client proberClient, key string) ProbeResult {
	if isPlaceholderResendKey(key) {
		return ProbeResult{Service: "resend", Status: ProbeStatusSkipped, Detail: "placeholder or empty key"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.resend.com/domains", nil)
	if err != nil {
		return ProbeResult{Service: "resend", Status: ProbeStatusFail, Detail: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+key)
	return runProbe(client, req, "resend")
}

func runProbe(client proberClient, req *http.Request, service string) ProbeResult {
	resp, err := client.Do(req)
	if err != nil {
		return ProbeResult{Service: service, Status: ProbeStatusFail, Detail: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ProbeResult{Service: service, Status: ProbeStatusOK, Detail: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		snippet = "no body"
	}
	return ProbeResult{Service: service, Status: ProbeStatusFail, Detail: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, snippet)}
}

func isPlaceholderAIKey(k string) bool {
	low := strings.ToLower(strings.TrimSpace(k))
	switch low {
	case "", "sk-...", "sk-xxx", "stub", "placeholder", "changeme", "stub-value":
		return true
	}
	return strings.HasPrefix(low, "sk-placeholder") || strings.HasPrefix(low, "sk-stub")
}

func isPlaceholderResendKey(k string) bool {
	low := strings.ToLower(strings.TrimSpace(k))
	switch low {
	case "", "re_...", "re_xxx", "stub", "placeholder", "changeme", "stub-value":
		return true
	}
	return strings.HasPrefix(low, "re_placeholder") || strings.HasPrefix(low, "re_stub")
}
