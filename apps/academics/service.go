package academics

import (
	"context"
)

//encore:service
type Service struct{}

func initService() (*Service, error) {
	return &Service{}, nil
}

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

//encore:api public method=GET path=/v1/academics/health
func (s *Service) Health(ctx context.Context) (*HealthResponse, error) {
	return &HealthResponse{Service: "academics", Status: "ok"}, nil
}
