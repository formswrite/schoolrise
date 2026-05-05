package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

func currentUserID(ctx context.Context) (int64, error) {
	uid, ok := encauth.UserID()
	if !ok || uid == "" {
		return 0, &errs.Error{Code: errs.Unauthenticated, Message: "missing user id"}
	}

	id, err := strconv.ParseInt(string(uid), 10, 64)
	if err != nil {
		return 0, &errs.Error{Code: errs.Unauthenticated, Message: "invalid user id"}
	}

	return id, nil
}

func requireGlobalAdmin(ctx context.Context) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}

	assignments, err := ListRoleAssignmentsForUser(ctx, userID)
	if err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not check role"}
	}

	for _, a := range assignments {
		if a.Role == "admin" && a.ScopeNodeID == nil {
			return nil
		}
	}

	return &errs.Error{Code: errs.PermissionDenied, Message: "global admin role required"}
}

type UserDTO struct {
	ID                 int64      `json:"id"`
	Email              string     `json:"email"`
	FullName           string     `json:"fullName"`
	Role               string     `json:"role"`
	MustChangePassword bool       `json:"mustChangePassword"`
	LockedAt           *time.Time `json:"lockedAt"`
	LastLoginAt        *time.Time `json:"lastLoginAt"`
	CreatedAt          time.Time  `json:"createdAt"`
}

func userToDTO(u *User) UserDTO {
	return UserDTO{
		ID:                 u.ID,
		Email:              u.Email,
		FullName:           u.FullName,
		Role:               u.Role,
		MustChangePassword: u.MustChangePassword,
		LockedAt:           u.LockedAt,
		LastLoginAt:        u.LastLoginAt,
		CreatedAt:          u.CreatedAt,
	}
}

type ListUsersResponse struct {
	Users []UserDTO `json:"users"`
}

//encore:api auth method=GET path=/v1/auth/users
func (s *Service) ListUsersAPI(ctx context.Context) (*ListUsersResponse, error) {
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}

	users, err := ListUsers(ctx, 200, 0)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not list users"}
	}

	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, userToDTO(u))
	}

	return &ListUsersResponse{Users: out}, nil
}

type CreateUserAPIRequest struct {
	Email              string `json:"email"`
	FullName           string `json:"fullName"`
	Password           string `json:"password"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"mustChangePassword"`
}

type CreateUserAPIResponse struct {
	User UserDTO `json:"user"`
}

//encore:api auth method=POST path=/v1/auth/users
func (s *Service) CreateUserAPI(ctx context.Context, req *CreateUserAPIRequest) (*CreateUserAPIResponse, error) {
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "teacher"
	}

	user, err := CreateUser(ctx, CreateUserParams{
		Email:              req.Email,
		Password:           req.Password,
		FullName:           req.FullName,
		Role:               role,
		MustChangePassword: req.MustChangePassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			return nil, &errs.Error{Code: errs.AlreadyExists, Message: "email already exists"}
		case errors.Is(err, ErrInvalidUserInput):
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid user input"}
		case errors.Is(err, ErrEmptyPassword):
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "password required"}
		default:
			return nil, &errs.Error{Code: errs.Internal, Message: "could not create user"}
		}
	}

	return &CreateUserAPIResponse{User: userToDTO(user)}, nil
}

type GetUserResponse struct {
	User UserDTO `json:"user"`
}

//encore:api auth method=GET path=/v1/auth/users/:id
func (s *Service) GetUserAPI(ctx context.Context, id int64) (*GetUserResponse, error) {
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}

	user, err := GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "user not found"}
		}

		return nil, &errs.Error{Code: errs.Internal, Message: "could not load user"}
	}

	return &GetUserResponse{User: userToDTO(user)}, nil
}

type AssignmentDTO struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userId"`
	Role        string `json:"role"`
	ScopeNodeID *int64 `json:"scopeNodeId"`
}

type ListAssignmentsResponse struct {
	Assignments []AssignmentDTO `json:"assignments"`
}

//encore:api auth method=GET path=/v1/auth/users/:userId/assignments
func (s *Service) ListUserAssignmentsAPI(ctx context.Context, userId int64) (*ListAssignmentsResponse, error) {
	if err := requireGlobalAdmin(ctx); err != nil {
		return nil, err
	}

	assignments, err := ListRoleAssignmentsForUser(ctx, userId)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not list assignments"}
	}

	out := make([]AssignmentDTO, 0, len(assignments))
	for _, a := range assignments {
		out = append(out, AssignmentDTO{
			ID:          a.ID,
			UserID:      a.UserID,
			Role:        a.Role,
			ScopeNodeID: a.ScopeNodeID,
		})
	}

	return &ListAssignmentsResponse{Assignments: out}, nil
}

type CreateAssignmentRequest struct {
	UserID      int64  `json:"userId"`
	Role        string `json:"role"`
	ScopeNodeID *int64 `json:"scopeNodeId"`
}

//encore:api auth method=POST path=/v1/auth/assignments
func (s *Service) CreateAssignmentAPI(ctx context.Context, req *CreateAssignmentRequest) error {
	if err := requireGlobalAdmin(ctx); err != nil {
		return err
	}

	if err := AssignRole(ctx, req.UserID, req.Role, req.ScopeNodeID); err != nil {
		if errors.Is(err, ErrInvalidRoleAssignment) {
			return &errs.Error{Code: errs.InvalidArgument, Message: "invalid role assignment"}
		}

		return &errs.Error{Code: errs.Internal, Message: "could not create assignment"}
	}

	return nil
}

//encore:api auth method=DELETE path=/v1/auth/assignments/:id
func (s *Service) DeleteAssignmentAPI(ctx context.Context, id int64) error {
	if err := requireGlobalAdmin(ctx); err != nil {
		return err
	}

	if err := RevokeRole(ctx, id); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not revoke assignment"}
	}

	return nil
}
