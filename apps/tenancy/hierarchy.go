package tenancy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/tenancy/dbtenancy"
)

var (
	ErrNodeNotFound           = errors.New("tenancy: node not found")
	ErrInvalidLevelTransition = errors.New("tenancy: invalid level transition")
	ErrCodeAlreadyExists      = errors.New("tenancy: code already exists under parent")
	ErrInvalidNodeInput       = errors.New("tenancy: invalid node input")
	ErrNodeHasChildren        = errors.New("tenancy: node has children, cannot delete")
)

type Node struct {
	ID        int64
	ParentID  *int64
	Level     string
	Code      string
	Label     string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateNodeParams struct {
	ParentID *int64
	Level    string
	Code     string
	Label    string
	Metadata map[string]any
}

func CreateNode(ctx context.Context, p CreateNodeParams) (*Node, error) {
	if strings.TrimSpace(p.Code) == "" || strings.TrimSpace(p.Label) == "" {
		return nil, ErrInvalidNodeInput
	}

	def, err := GetLevel(ctx, p.Level)
	if err != nil {
		if errors.Is(err, ErrUnknownLevel) {
			return nil, ErrInvalidLevelTransition
		}

		return nil, err
	}

	if def.ParentLevel == "" {
		if p.ParentID != nil {
			return nil, ErrInvalidLevelTransition
		}
	} else {
		if p.ParentID == nil {
			return nil, ErrInvalidLevelTransition
		}

		parent, err := queries.GetHierarchyNodeByID(ctx, *p.ParentID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNodeNotFound
		}

		if err != nil {
			return nil, err
		}

		if parent.Level != def.ParentLevel {
			return nil, ErrInvalidLevelTransition
		}
	}

	metaBytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, err
	}

	if string(metaBytes) == "null" {
		metaBytes = []byte("{}")
	}

	parentSQL := sql.NullInt64{}
	if p.ParentID != nil {
		parentSQL = sql.NullInt64{Int64: *p.ParentID, Valid: true}
	}

	row, err := queries.CreateHierarchyNode(ctx, dbtenancy.CreateHierarchyNodeParams{
		ParentID: parentSQL,
		Level:    p.Level,
		Code:     p.Code,
		Label:    p.Label,
		Metadata: metaBytes,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeAlreadyExists
		}

		return nil, err
	}

	if err := queries.InsertSelfClosure(ctx, row.ID); err != nil {
		return nil, err
	}

	if p.ParentID != nil {
		if err := queries.InsertClosureFromParent(ctx, dbtenancy.InsertClosureFromParentParams{
			Column1: *p.ParentID,
			Column2: row.ID,
		}); err != nil {
			return nil, err
		}
	}

	return nodeFromRow(row), nil
}

func GetNodeByID(ctx context.Context, id int64) (*Node, error) {
	row, err := queries.GetHierarchyNodeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNodeNotFound
	}

	if err != nil {
		return nil, err
	}

	return nodeFromRow(row), nil
}

func nodeFromRow(r dbtenancy.HierarchyNode) *Node {
	n := &Node{
		ID:        r.ID,
		Level:     r.Level,
		Code:      r.Code,
		Label:     r.Label,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}

	if r.ParentID.Valid {
		pid := r.ParentID.Int64
		n.ParentID = &pid
	}

	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &n.Metadata)
	}

	if n.Metadata == nil {
		n.Metadata = map[string]any{}
	}

	return n
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
