package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"encore.app/apps/auth/dbauth"
)

var ErrInvalidRoleAssignment = errors.New("auth: invalid role assignment")

type RoleAssignment struct {
	ID          int64
	UserID      int64
	Role        string
	ScopeNodeID *int64
}

func AssignRole(ctx context.Context, userID int64, role string, scopeNodeID *int64) error {
	if userID == 0 || strings.TrimSpace(role) == "" {
		return ErrInvalidRoleAssignment
	}

	scope := sql.NullInt64{}
	if scopeNodeID != nil {
		scope = sql.NullInt64{Int64: *scopeNodeID, Valid: true}
	}

	return queries.CreateRoleAssignment(ctx, dbauth.CreateRoleAssignmentParams{
		UserID:      userID,
		Role:        role,
		ScopeNodeID: scope,
	})
}

func RevokeRole(ctx context.Context, assignmentID int64) error {
	return queries.DeleteRoleAssignment(ctx, assignmentID)
}

func ListRoleAssignmentsForUser(ctx context.Context, userID int64) ([]RoleAssignment, error) {
	rows, err := queries.ListRoleAssignmentsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]RoleAssignment, 0, len(rows))
	for _, r := range rows {
		ra := RoleAssignment{
			ID:     r.ID,
			UserID: r.UserID,
			Role:   r.Role,
		}

		if r.ScopeNodeID.Valid {
			id := r.ScopeNodeID.Int64
			ra.ScopeNodeID = &id
		}

		out = append(out, ra)
	}

	return out, nil
}
