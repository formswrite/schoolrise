package tenancy

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"
)

type LevelDTO struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	ParentLevel string `json:"parentLevel"`
	Depth       int    `json:"depth"`
	SortOrder   int    `json:"sortOrder"`
}

type ListLevelsResponse struct {
	Levels []LevelDTO `json:"levels"`
}

//encore:api auth method=GET path=/v1/tenancy/levels
func (s *Service) ListLevelsAPI(ctx context.Context) (*ListLevelsResponse, error) {
	defs, err := ListLevels(ctx)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not list levels"}
	}

	out := make([]LevelDTO, 0, len(defs))
	for _, d := range defs {
		out = append(out, LevelDTO{
			Code:        d.Code,
			Label:       d.Label,
			ParentLevel: d.ParentLevel,
			Depth:       d.Depth,
			SortOrder:   d.SortOrder,
		})
	}

	return &ListLevelsResponse{Levels: out}, nil
}

type NodeDTO struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parentId"`
	Level     string    `json:"level"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListNodesRequest struct {
	ParentID int64  `query:"parentId"`
	Level    string `query:"level"`
}

type ListNodesResponse struct {
	Nodes []NodeDTO `json:"nodes"`
}

//encore:api auth method=GET path=/v1/tenancy/nodes
func (s *Service) ListNodesAPI(ctx context.Context, req *ListNodesRequest) (*ListNodesResponse, error) {
	var (
		nodes []*Node
		err   error
	)

	if req.Level != "" {
		nodes, err = ListByLevel(ctx, req.Level)
	} else {
		var parent *int64
		if req.ParentID != 0 {
			id := req.ParentID
			parent = &id
		}

		if parent != nil {
			if err := requireNodeAccess(ctx, *parent); err != nil {
				return nil, err
			}
		}

		nodes, err = ListChildren(ctx, parent)
	}

	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not list nodes"}
	}

	filtered, err := filterAccessibleNodes(ctx, nodes)
	if err != nil {
		return nil, err
	}

	out := make([]NodeDTO, 0, len(filtered))
	for _, n := range filtered {
		out = append(out, nodeToDTO(n))
	}

	return &ListNodesResponse{Nodes: out}, nil
}

type GetNodeResponse struct {
	Node NodeDTO `json:"node"`
}

//encore:api auth method=GET path=/v1/tenancy/nodes/:id
func (s *Service) GetNodeAPI(ctx context.Context, id int64) (*GetNodeResponse, error) {
	if err := requireNodeAccess(ctx, id); err != nil {
		return nil, err
	}

	node, err := GetNodeByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNodeNotFound) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "node not found"}
		}

		return nil, &errs.Error{Code: errs.Internal, Message: "could not load node"}
	}

	return &GetNodeResponse{Node: nodeToDTO(node)}, nil
}

type CreateNodeAPIRequest struct {
	ParentID *int64 `json:"parentId"`
	Level    string `json:"level"`
	Code     string `json:"code"`
	Label    string `json:"label"`
}

type CreateNodeResponse struct {
	Node NodeDTO `json:"node"`
}

//encore:api auth method=POST path=/v1/tenancy/nodes
func (s *Service) CreateNodeAPI(ctx context.Context, req *CreateNodeAPIRequest) (*CreateNodeResponse, error) {
	if req.ParentID != nil {
		if err := requireNodeAccess(ctx, *req.ParentID); err != nil {
			return nil, err
		}
	} else {
		userID, err := currentUserID(ctx)
		if err != nil {
			return nil, err
		}

		assignments, err := loadAssignmentsForUser(ctx, userID)
		if err != nil {
			return nil, &errs.Error{Code: errs.Internal, Message: "authorization check failed"}
		}

		hasGlobal := false
		for _, a := range assignments {
			if a.ScopeNodeID == nil {
				hasGlobal = true
				break
			}
		}

		if !hasGlobal {
			return nil, &errs.Error{Code: errs.PermissionDenied, Message: "creating root nodes requires global admin role"}
		}
	}

	node, err := CreateNode(ctx, CreateNodeParams{
		ParentID: req.ParentID,
		Level:    req.Level,
		Code:     req.Code,
		Label:    req.Label,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidNodeInput):
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "code and label required"}
		case errors.Is(err, ErrInvalidLevelTransition):
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid level or parent level mismatch"}
		case errors.Is(err, ErrCodeAlreadyExists):
			return nil, &errs.Error{Code: errs.AlreadyExists, Message: "code already exists under parent"}
		case errors.Is(err, ErrNodeNotFound):
			return nil, &errs.Error{Code: errs.NotFound, Message: "parent node not found"}
		default:
			return nil, &errs.Error{Code: errs.Internal, Message: "could not create node"}
		}
	}

	return &CreateNodeResponse{Node: nodeToDTO(node)}, nil
}

//encore:api auth method=DELETE path=/v1/tenancy/nodes/:id
func (s *Service) DeleteNodeAPI(ctx context.Context, id int64) error {
	if err := requireNodeAccess(ctx, id); err != nil {
		return err
	}

	if err := SoftDeleteNode(ctx, id); err != nil {
		switch {
		case errors.Is(err, ErrNodeHasChildren):
			return &errs.Error{Code: errs.FailedPrecondition, Message: "node has children — delete them first"}
		default:
			return &errs.Error{Code: errs.Internal, Message: "could not delete node"}
		}
	}

	return nil
}

func nodeToDTO(n *Node) NodeDTO {
	return NodeDTO{
		ID:        n.ID,
		ParentID:  n.ParentID,
		Level:     n.Level,
		Code:      n.Code,
		Label:     n.Label,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}
