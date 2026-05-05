package seed

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProbeOpenAI_PlaceholderKeySkipped(t *testing.T) {
	for _, key := range []string{"", "sk-...", "sk-xxx", "sk-placeholder-anything", "sk-stub-yyy", "stub-value", "CHANGEME"} {
		r := probeOpenAI(context.Background(), nil, key)
		if r.Status != ProbeStatusSkipped {
			t.Errorf("key %q: status = %s, want skipped", key, r.Status)
		}
	}
}

func TestProbeResend_PlaceholderKeySkipped(t *testing.T) {
	for _, key := range []string{"", "re_...", "re_xxx", "re_placeholder-anything", "re_stub-yyy", "stub-value"} {
		r := probeResend(context.Background(), nil, key)
		if r.Status != ProbeStatusSkipped {
			t.Errorf("key %q: status = %s, want skipped", key, r.Status)
		}
	}
}

func TestRunProbe_2xxIsOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	r := runProbe(srv.Client(), req, "openai")
	if r.Status != ProbeStatusOK {
		t.Fatalf("status = %s, detail = %s", r.Status, r.Detail)
	}
}

func TestRunProbe_4xxIsFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	r := runProbe(srv.Client(), req, "openai")
	if r.Status != ProbeStatusFail {
		t.Fatalf("status = %s, want fail", r.Status)
	}
	if !strings.Contains(r.Detail, "401") {
		t.Errorf("detail should mention status code, got %q", r.Detail)
	}
}

func TestRunProbe_TransportErrorIsFail(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:1/will-not-connect", nil)
	r := runProbe(&http.Client{}, req, "resend")
	if r.Status != ProbeStatusFail {
		t.Fatalf("status = %s, want fail", r.Status)
	}
}

func TestLogProbeResults_FormatsByStatus(t *testing.T) {
	results := []ProbeResult{
		{Service: "openai", Status: ProbeStatusOK, Detail: "HTTP 200"},
		{Service: "resend", Status: ProbeStatusSkipped, Detail: "placeholder or empty key"},
		{Service: "stripe", Status: ProbeStatusFail, Detail: "HTTP 401: bad key"},
	}
	var lines []string
	LogProbeResults(results, func(format string, args ...any) {
		lines = append(lines, fmt.Sprintf(format, args...))
	})
	if len(lines) != 3 {
		t.Fatalf("got %d log lines, want 3", len(lines))
	}
	if !strings.Contains(lines[0], "openai=ok") {
		t.Errorf("ok log = %q", lines[0])
	}
	if !strings.Contains(lines[1], "resend=skipped") {
		t.Errorf("skipped log = %q", lines[1])
	}
	if !strings.Contains(lines[2], "stripe=FAIL") || !strings.Contains(lines[2], "401") {
		t.Errorf("fail log = %q", lines[2])
	}
}

