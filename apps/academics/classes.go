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
	ErrClassNotFound      = errors.New("academics: class not found")
	ErrClassCodeTaken     = errors.New("academics: class code already used in this institution+period")
	ErrInvalidClassInput  = errors.New("academics: invalid class input")
)

type Class struct {
	ID            int64
	PeriodID      int64
	NiveauID      int64
	InstitutionID int64
	Code          string
	Label         string
	Capacity      int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateClassParams struct {
	PeriodID      int64
	NiveauID      int64
	InstitutionID int64
	Code          string
	Label         string
	Capacity      int32
}

func CreateClass(ctx context.Context, p CreateClassParams) (*Class, error) {
	code := strings.TrimSpace(p.Code)
	label := strings.TrimSpace(p.Label)

	if code == "" || label == "" || p.PeriodID <= 0 || p.NiveauID <= 0 || p.InstitutionID <= 0 {
		return nil, ErrInvalidClassInput
	}

	if _, err := GetPeriodByID(ctx, p.PeriodID); err != nil {
		return nil, err
	}

	if _, err := GetNiveauByID(ctx, p.NiveauID); err != nil {
		return nil, err
	}

	row, err := queries.CreateClass(ctx, dbacademics.CreateClassParams{
		PeriodID:      p.PeriodID,
		NiveauID:      p.NiveauID,
		InstitutionID: p.InstitutionID,
		Code:          code,
		Label:         label,
		Capacity:      nullableInt32(p.Capacity),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrClassCodeTaken
		}
		return nil, err
	}

	return classFromRow(row), nil
}

func GetClassByID(ctx context.Context, id int64) (*Class, error) {
	row, err := queries.GetClassByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrClassNotFound
	}
	if err != nil {
		return nil, err
	}
	return classFromRow(row), nil
}

func ListClassesByInstitution(ctx context.Context, institutionID int64) ([]*Class, error) {
	rows, err := queries.ListClassesByInstitution(ctx, institutionID)
	if err != nil {
		return nil, err
	}
	out := make([]*Class, 0, len(rows))
	for _, r := range rows {
		out = append(out, classFromRow(r))
	}
	return out, nil
}

func ListClassesByPeriod(ctx context.Context, periodID int64) ([]*Class, error) {
	rows, err := queries.ListClassesByPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}
	out := make([]*Class, 0, len(rows))
	for _, r := range rows {
		out = append(out, classFromRow(r))
	}
	return out, nil
}

func DeleteClass(ctx context.Context, id int64) error {
	if _, err := GetClassByID(ctx, id); err != nil {
		return err
	}
	return queries.SoftDeleteClass(ctx, id)
}

func classFromRow(r dbacademics.Class) *Class {
	c := &Class{
		ID:            r.ID,
		PeriodID:      r.PeriodID,
		NiveauID:      r.NiveauID,
		InstitutionID: r.InstitutionID,
		Code:          r.Code,
		Label:         r.Label,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if r.Capacity.Valid {
		c.Capacity = r.Capacity.Int32
	}
	return c
}

func nullableInt32(v int32) sql.NullInt32 {
	if v <= 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: v, Valid: true}
}
