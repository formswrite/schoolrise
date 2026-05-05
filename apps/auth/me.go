package auth

import (
	"context"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type AssignmentSummary struct {
	Role        string `json:"role"`
	ScopeNodeID *int64 `json:"scopeNodeId"`
}

type MeResponse struct {
	UserID             int64
	Email              string
	FullName           string
	Role               string
	MustChangePassword bool
	Assignments        []AssignmentSummary
}

//encore:api auth method=GET path=/v1/auth/me
func (s *Service) MeAPI(ctx context.Context) (*MeResponse, error) {
	data, ok := encauth.Data().(*AuthData)
	if !ok || data == nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "missing auth data"}
	}

	user, err := GetUserByID(ctx, data.UserID)
	if err != nil {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "user not found"}
	}

	rawAssignments, err := ListRoleAssignmentsForUser(ctx, user.ID)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not load role assignments"}
	}

	assignments := make([]AssignmentSummary, 0, len(rawAssignments))
	for _, a := range rawAssignments {
		assignments = append(assignments, AssignmentSummary{Role: a.Role, ScopeNodeID: a.ScopeNodeID})
	}

	return &MeResponse{
		UserID:             user.ID,
		Email:              user.Email,
		FullName:           user.FullName,
		Role:               user.Role,
		MustChangePassword: user.MustChangePassword,
		Assignments:        assignments,
	}, nil
}
