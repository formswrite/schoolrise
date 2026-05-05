package setup

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"encore.dev/beta/errs"

	"encore.app/apps/auth"
)

type CreateAdminRequest struct {
	SessionToken string
	Email        string
	FullName     string
	Password     string
}

type CreateAdminResponse struct {
	UserID int64
}

//encore:api public method=POST path=/v1/setup/admin
func (s *Service) CreateAdminAPI(ctx context.Context, req *CreateAdminRequest) (*CreateAdminResponse, error) {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.FullName) == "" || req.Password == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "email, full name and password required"}
	}

	user, err := auth.CreateUser(ctx, auth.CreateUserParams{
		Email:              req.Email,
		Password:           req.Password,
		FullName:           req.FullName,
		Role:               "admin",
		MustChangePassword: false,
	})
	if err != nil {
		if errors.Is(err, auth.ErrEmailAlreadyExists) {
			return nil, &errs.Error{Code: errs.AlreadyExists, Message: "an account with that email already exists"}
		}

		if errors.Is(err, auth.ErrInvalidUserInput) {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid admin input"}
		}

		return nil, &errs.Error{Code: errs.Internal, Message: "could not create admin"}
	}

	if err := auth.AssignRole(ctx, user.ID, "admin", nil); err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not assign admin role"}
	}

	payload, _ := json.Marshal(map[string]any{"user_id": user.ID, "email": user.Email})
	if err := markStepComplete(ctx, "admin", payload); err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not mark progress"}
	}

	return &CreateAdminResponse{UserID: user.ID}, nil
}
