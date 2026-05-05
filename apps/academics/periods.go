package academics

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/academics/dbacademics"
)

var (
	ErrPeriodNotFound      = errors.New("academics: period not found")
	ErrPeriodCodeTaken     = errors.New("academics: period code already used")
	ErrInvalidPeriodInput  = errors.New("academics: invalid period input")
	ErrPeriodDateRange     = errors.New("academics: ends_on must be on/after starts_on")
)

type Period struct {
	ID        int64
	Code      string
	Label     string
	StartsOn  time.Time
	EndsOn    time.Time
	IsCurrent bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreatePeriodParams struct {
	Code      string
	Label     string
	StartsOn  time.Time
	EndsOn    time.Time
	IsCurrent bool
}

func CreatePeriod(ctx context.Context, p CreatePeriodParams) (*Period, error) {
	code := strings.TrimSpace(p.Code)
	label := strings.TrimSpace(p.Label)

	if code == "" || label == "" {
		return nil, ErrInvalidPeriodInput
	}

	if p.StartsOn.IsZero() || p.EndsOn.IsZero() {
		return nil, ErrInvalidPeriodInput
	}

	if p.EndsOn.Before(p.StartsOn) {
		return nil, ErrPeriodDateRange
	}

	if p.IsCurrent {
		if err := queries.ClearCurrentPeriod(ctx); err != nil {
			return nil, err
		}
	}

	row, err := queries.CreatePeriod(ctx, dbacademics.CreatePeriodParams{
		Code:      code,
		Label:     label,
		StartsOn:  p.StartsOn,
		EndsOn:    p.EndsOn,
		IsCurrent: p.IsCurrent,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrPeriodCodeTaken
		}
		return nil, err
	}

	return periodFromRow(row), nil
}

func GetPeriodByID(ctx context.Context, id int64) (*Period, error) {
	row, err := queries.GetPeriodByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}
	return periodFromRow(row), nil
}

func GetCurrentPeriod(ctx context.Context) (*Period, error) {
	row, err := queries.GetCurrentPeriod(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}
	return periodFromRow(row), nil
}

func ListPeriods(ctx context.Context) ([]*Period, error) {
	rows, err := queries.ListPeriods(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Period, 0, len(rows))
	for _, r := range rows {
		out = append(out, periodFromRow(r))
	}
	return out, nil
}

func SetPeriodCurrent(ctx context.Context, id int64) (*Period, error) {
	if _, err := GetPeriodByID(ctx, id); err != nil {
		return nil, err
	}

	if err := queries.ClearCurrentPeriod(ctx); err != nil {
		return nil, err
	}

	row, err := queries.SetPeriodCurrent(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}
	return periodFromRow(row), nil
}

func DeletePeriod(ctx context.Context, id int64) error {
	if _, err := GetPeriodByID(ctx, id); err != nil {
		return err
	}
	return queries.SoftDeletePeriod(ctx, id)
}

func periodFromRow(r dbacademics.AcademicPeriod) *Period {
	return &Period{
		ID:        r.ID,
		Code:      r.Code,
		Label:     r.Label,
		StartsOn:  r.StartsOn,
		EndsOn:    r.EndsOn,
		IsCurrent: r.IsCurrent,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
