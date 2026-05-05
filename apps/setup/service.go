package setup

import (
	"context"
	"fmt"
	"os"

	"encore.app/internal/seed"
)

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/v1/setup/health
func (s *Service) Health(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{Service: "setup", Status: "ok"}, nil
}

//encore:service
type Service struct{}

func initService() (*Service, error) {
	ctx := context.Background()

	plaintext, err := ensureInstallToken(ctx)
	if err != nil {
		return nil, err
	}

	if plaintext != "" {
		printInstallToken(plaintext)
	}

	if err := RunHeadless(ctx); err != nil {
		return nil, err
	}

	seed.LogProbeResults(seed.ProbeIntegrations(ctx), func(format string, args ...any) {
		_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	})

	return &Service{}, nil
}

func printInstallToken(token string) {
	banner := "==========================================================="
	msg := fmt.Sprintf(
		"\n%s\nSchoolRise install token (one-time, single-use):\n    %s\nThe first visitor to /setup with this token claims the admin account.\n%s\n",
		banner, token, banner,
	)

	_, _ = os.Stderr.WriteString(msg)
}
