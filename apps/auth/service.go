package auth

import (
	"context"
)

//encore:service
type Service struct{}

func initService() (*Service, error) {
	ctx := context.Background()

	var n int
	if err := db.QueryRow(ctx, "SELECT 1").Scan(&n); err != nil {
		return nil, err
	}

	return &Service{}, nil
}

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/v1/auth/health
func (s *Service) Health(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{Service: "auth", Status: "ok"}, nil
}
