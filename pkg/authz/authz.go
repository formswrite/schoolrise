package authz

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const modelDef = `
[request_definition]
r = sub, obj

[policy_definition]
p = sub, scope

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && (p.scope == "0" || hasAncestor(p.scope, r.obj))
`

type Assignment struct {
	UserID      int64
	Role        string
	ScopeNodeID *int64
}

type AssignmentLoader func(ctx context.Context, userID int64) ([]Assignment, error)
type AncestryCheck func(ctx context.Context, ancestorID, descendantID int64) (bool, error)

type Enforcer struct {
	loadAssignments AssignmentLoader
	isAncestor      AncestryCheck
}

func New(loadAssignments AssignmentLoader, isAncestor AncestryCheck) (*Enforcer, error) {
	if loadAssignments == nil || isAncestor == nil {
		return nil, errors.New("authz: loader and ancestry check required")
	}

	return &Enforcer{
		loadAssignments: loadAssignments,
		isAncestor:      isAncestor,
	}, nil
}

func (e *Enforcer) CanAccessNode(ctx context.Context, userID, nodeID int64) (bool, error) {
	if userID == 0 {
		return false, nil
	}

	assignments, err := e.loadAssignments(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("load assignments: %w", err)
	}

	if len(assignments) == 0 {
		return false, nil
	}

	enf, err := e.buildPerRequestEnforcer(ctx, assignments)
	if err != nil {
		return false, err
	}

	return enf.Enforce(fmt.Sprint(userID), fmt.Sprint(nodeID))
}

func (e *Enforcer) buildPerRequestEnforcer(ctx context.Context, assignments []Assignment) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(modelDef)
	if err != nil {
		return nil, fmt.Errorf("authz: model: %w", err)
	}

	enf, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("authz: enforcer: %w", err)
	}

	enf.AddFunction("hasAncestor", e.makeAncestorFunc(ctx))

	for _, a := range assignments {
		var scope int64
		if a.ScopeNodeID != nil {
			scope = *a.ScopeNodeID
		}

		if _, err := enf.AddPolicy(fmt.Sprint(a.UserID), fmt.Sprint(scope)); err != nil {
			return nil, fmt.Errorf("authz: add policy: %w", err)
		}
	}

	return enf, nil
}

func (e *Enforcer) makeAncestorFunc(ctx context.Context) func(args ...any) (any, error) {
	return func(args ...any) (any, error) {
		if len(args) != 2 {
			return false, errors.New("hasAncestor requires 2 args")
		}

		ancestor, err := toInt64(args[0])
		if err != nil {
			return false, err
		}

		descendant, err := toInt64(args[1])
		if err != nil {
			return false, err
		}

		ok, err := e.isAncestor(ctx, ancestor, descendant)
		if err != nil {
			return false, err
		}

		return ok, nil
	}
}

func toInt64(v any) (int64, error) {
	switch x := v.(type) {
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	case float64:
		return int64(x), nil
	case string:
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("authz: cannot parse %q as int64", x)
		}

		return n, nil
	default:
		return 0, fmt.Errorf("authz: unsupported type %T", v)
	}
}
