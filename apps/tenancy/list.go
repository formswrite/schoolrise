package tenancy

import (
	"context"
	"database/sql"

	"encore.app/apps/tenancy/dbtenancy"
)

type DescendantNode struct {
	Node  *Node
	Depth int
}

func ListChildren(ctx context.Context, parentID *int64) ([]*Node, error) {
	var param int64
	if parentID != nil {
		param = *parentID
	}

	rows, err := queries.ListHierarchyNodesByParent(ctx, param)
	if err != nil {
		return nil, err
	}

	out := make([]*Node, 0, len(rows))
	for _, r := range rows {
		out = append(out, nodeFromRow(r))
	}

	return out, nil
}

func ListByLevel(ctx context.Context, level string) ([]*Node, error) {
	rows, err := queries.ListHierarchyNodesByLevel(ctx, level)
	if err != nil {
		return nil, err
	}

	out := make([]*Node, 0, len(rows))
	for _, r := range rows {
		out = append(out, nodeFromRow(r))
	}

	return out, nil
}

func DescendantsOf(ctx context.Context, ancestorID int64) ([]DescendantNode, error) {
	rows, err := queries.GetDescendants(ctx, ancestorID)
	if err != nil {
		return nil, err
	}

	out := make([]DescendantNode, 0, len(rows))
	for _, r := range rows {
		node := &Node{
			ID:        r.ID,
			Level:     r.Level,
			Code:      r.Code,
			Label:     r.Label,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}

		if r.ParentID.Valid {
			pid := r.ParentID.Int64
			node.ParentID = &pid
		}

		out = append(out, DescendantNode{Node: node, Depth: int(r.Depth)})
	}

	return out, nil
}

func IsAncestor(ctx context.Context, ancestorID, descendantID int64) (bool, error) {
	if ancestorID == descendantID {
		return true, nil
	}

	return queries.IsAncestorClosure(ctx, dbtenancy.IsAncestorClosureParams{
		AncestorID:   ancestorID,
		DescendantID: descendantID,
	})
}

func ListDescendantIDs(ctx context.Context, ancestorID int64) ([]int64, error) {
	return queries.ListDescendantIDs(ctx, ancestorID)
}

func ListAncestorIDs(ctx context.Context, descendantID int64) ([]int64, error) {
	return queries.ListAncestorIDs(ctx, descendantID)
}

func ListAncestorIDsForMany(ctx context.Context, descendantIDs []int64) ([]int64, error) {
	if len(descendantIDs) == 0 {
		return nil, nil
	}
	return queries.ListAncestorIDsForMany(ctx, descendantIDs)
}

func SoftDeleteNode(ctx context.Context, id int64) error {
	hasChildren, err := queries.HasUndeletedChildren(ctx, sql.NullInt64{Int64: id, Valid: true})
	if err != nil {
		return err
	}

	if hasChildren {
		return ErrNodeHasChildren
	}

	return queries.SoftDeleteHierarchyNode(ctx, id)
}
