package academics

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"encore.app/apps/academics/dbacademics"
)

var (
	ErrNiveauNotFound     = errors.New("academics: niveau not found")
	ErrNiveauCodeTaken    = errors.New("academics: niveau code already used")
	ErrInvalidNiveauInput = errors.New("academics: invalid niveau input")
)

type Niveau struct {
	ID        int64
	Code      string
	Label     string
	SortOrder int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateNiveauParams struct {
	Code      string
	Label     string
	SortOrder int32
}

func CreateNiveau(ctx context.Context, p CreateNiveauParams) (*Niveau, error) {
	code := strings.TrimSpace(p.Code)
	label := strings.TrimSpace(p.Label)

	if code == "" || label == "" {
		return nil, ErrInvalidNiveauInput
	}

	row, err := queries.CreateNiveau(ctx, dbacademics.CreateNiveauParams{
		Code:      code,
		Label:     label,
		SortOrder: p.SortOrder,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrNiveauCodeTaken
		}
		return nil, err
	}

	return niveauFromRow(row), nil
}

func GetNiveauByID(ctx context.Context, id int64) (*Niveau, error) {
	row, err := queries.GetNiveauByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNiveauNotFound
	}
	if err != nil {
		return nil, err
	}
	return niveauFromRow(row), nil
}

func ListNiveaux(ctx context.Context) ([]*Niveau, error) {
	rows, err := queries.ListNiveaux(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Niveau, 0, len(rows))
	for _, r := range rows {
		out = append(out, niveauFromRow(r))
	}
	return out, nil
}

func DeleteNiveau(ctx context.Context, id int64) error {
	if _, err := GetNiveauByID(ctx, id); err != nil {
		return err
	}
	return queries.SoftDeleteNiveau(ctx, id)
}

func niveauFromRow(r dbacademics.Niveaux) *Niveau {
	return &Niveau{
		ID:        r.ID,
		Code:      r.Code,
		Label:     r.Label,
		SortOrder: r.SortOrder,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
