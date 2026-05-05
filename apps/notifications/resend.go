package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

type Provider interface {
	Send(ctx context.Context, e EmailRequest) (*ProviderResult, error)
	Name() string
}

type EmailRequest struct {
	From    string
	To      string
	ToName  string
	Subject string
	HTML    string
	Text    string
}

type ProviderResult struct {
	ProviderID string
}

type LogProvider struct {
	mu    sync.Mutex
	sent  []EmailRequest
	logger func(string, ...any)
}

func NewLogProvider() *LogProvider {
	return &LogProvider{
		logger: func(format string, args ...any) {
			fmt.Fprintf(os.Stdout, "[notifications.LogProvider] "+format+"\n", args...)
		},
	}
}

func (p *LogProvider) Send(_ context.Context, e EmailRequest) (*ProviderResult, error) {
	p.mu.Lock()
	p.sent = append(p.sent, e)
	count := len(p.sent)
	p.mu.Unlock()
	p.logger("sent #%d kind=email to=%s subject=%q", count, e.To, e.Subject)
	return &ProviderResult{ProviderID: fmt.Sprintf("log-%d-%d", time.Now().UnixNano(), count)}, nil
}

func (p *LogProvider) Name() string { return "log" }

func (p *LogProvider) Sent() []EmailRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]EmailRequest, len(p.sent))
	copy(out, p.sent)
	return out
}

func (p *LogProvider) Reset() {
	p.mu.Lock()
	p.sent = nil
	p.mu.Unlock()
}

type ResendProvider struct {
	APIKey  string
	Client  *http.Client
	BaseURL string
}

func NewResendProvider(apiKey string) *ResendProvider {
	return &ResendProvider{
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 10 * time.Second},
		BaseURL: resendEndpoint,
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text,omitempty"`
}

type resendResponse struct {
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
}

func (p *ResendProvider) Send(ctx context.Context, e EmailRequest) (*ProviderResult, error) {
	if p.APIKey == "" {
		return nil, errors.New("notifications: missing RESEND_API_KEY")
	}
	from := e.From
	to := e.To
	if e.ToName != "" {
		to = fmt.Sprintf("%s <%s>", e.ToName, e.To)
	}

	body, err := json.Marshal(resendPayload{
		From:    from,
		To:      []string{to},
		Subject: e.Subject,
		HTML:    e.HTML,
		Text:    e.Text,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("resend %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed resendResponse
	_ = json.Unmarshal(respBody, &parsed)
	return &ProviderResult{ProviderID: parsed.ID}, nil
}

func (p *ResendProvider) Name() string { return "resend" }

var (
	providerMu       sync.RWMutex
	currentProvider  Provider
	currentEmailFrom string
	currentFromName  string
)

func SetProvider(p Provider) {
	providerMu.Lock()
	currentProvider = p
	providerMu.Unlock()
}

func SetSender(emailFrom, fromName string) {
	providerMu.Lock()
	currentEmailFrom = emailFrom
	currentFromName = fromName
	providerMu.Unlock()
}

func getProvider() Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return currentProvider
}

func getSender() string {
	providerMu.RLock()
	defer providerMu.RUnlock()
	if currentFromName != "" && currentEmailFrom != "" {
		return fmt.Sprintf("%s <%s>", currentFromName, currentEmailFrom)
	}
	return currentEmailFrom
}

func init() {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	from := strings.TrimSpace(os.Getenv("EMAIL_FROM"))
	fromName := strings.TrimSpace(os.Getenv("EMAIL_FROM_NAME"))
	override := strings.ToLower(strings.TrimSpace(os.Getenv("NOTIFICATIONS_PROVIDER")))
	if from == "" {
		from = "schoolrise@local.test"
	}
	SetSender(from, fromName)

	useResend := apiKey != "" && !isPlaceholderKey(apiKey)
	if override == "log" {
		useResend = false
	} else if override == "resend" {
		useResend = apiKey != ""
	}

	if useResend {
		SetProvider(NewResendProvider(apiKey))
	} else {
		SetProvider(NewLogProvider())
	}
}

func isPlaceholderKey(k string) bool {
	low := strings.ToLower(k)
	switch low {
	case "re_local_stub", "re_xxxxxxxxxxxxxxxxxxxxxx", "re_xxx", "stub", "placeholder", "changeme":
		return true
	}
	return strings.HasPrefix(low, "re_local_") || strings.HasPrefix(low, "re_test_stub") || strings.HasPrefix(low, "re_placeholder")
}
