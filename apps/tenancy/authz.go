package tenancy

import (
	"context"
	"errors"
	"strconv"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/apps/auth"
	"encore.app/pkg/authz"
)

var enforcer = func() *authz.Enforcer {
	e, err := authz.New(loadAssignmentsForUser, IsAncestor)
	if err != nil {
		panic(err)
	}

	return e
}()

func loadAssignmentsForUser(ctx context.Context, userID int64) ([]authz.Assignment, error) {
	rows, err := auth.ListRoleAssignmentsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]authz.Assignment, 0, len(rows))
	for _, r := range rows {
		out = append(out, authz.Assignment{
			UserID:      r.UserID,
			Role:        r.Role,
			ScopeNodeID: r.ScopeNodeID,
		})
	}

	return out, nil
}

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

func requireNodeAccess(ctx context.Context, nodeID int64) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}

	allowed, err := enforcer.CanAccessNode(ctx, userID, nodeID)
	if err != nil {
		return &errs.Error{Code: errs.Internal, Message: "authorization check failed"}
	}

	if !allowed {
		return &errs.Error{Code: errs.PermissionDenied, Message: "out of scope"}
	}

	return nil
}

func filterAccessibleNodes(ctx context.Context, nodes []*Node) ([]*Node, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*Node, 0, len(nodes))

	for _, n := range nodes {
		allowed, err := enforcer.CanAccessNode(ctx, userID, n.ID)
		if err != nil {
			return nil, &errs.Error{Code: errs.Internal, Message: "authorization check failed"}
		}

		if allowed {
			out = append(out, n)
		}
	}

	return out, nil
}

var (
	_ = errors.New
)
