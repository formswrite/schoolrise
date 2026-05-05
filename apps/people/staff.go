package people

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"encore.app/apps/people/dbpeople"
)

var (
	ErrStaffNotFound       = errors.New("people: staff not found")
	ErrStaffCodeAlreadyUsed = errors.New("people: staff code already used in this scope")
	ErrInvalidStaffInput   = errors.New("people: invalid staff input")
)

type Staff struct {
	ID          int64
	PersonID    int64
	ScopeNodeID int64
	Position    string
	StaffCode   string
	HireDate    *time.Time
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateStaffParams struct {
	Person      CreatePersonParams
	ScopeNodeID int64
	Position    string
	StaffCode   string
	HireDate    *time.Time
	Metadata    map[string]any
}

func CreateStaff(ctx context.Context, p CreateStaffParams) (*Staff, *Person, error) {
	if p.ScopeNodeID <= 0 {
		return nil, nil, ErrInvalidStaffInput
	}

	person, err := CreatePerson(ctx, p.Person)
	if err != nil {
		return nil, nil, err
	}

	metaBytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, nil, err
	}

	if string(metaBytes) == "null" {
		metaBytes = []byte("{}")
	}

	row, err := queries.CreateStaff(ctx, dbpeople.CreateStaffParams{
		PersonID:    person.ID,
		ScopeNodeID: p.ScopeNodeID,
		Position:    nullableString(p.Position),
		StaffCode:   nullableString(p.StaffCode),
		HireDate:    nullableTime(p.HireDate),
		Metadata:    metaBytes,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, nil, ErrStaffCodeAlreadyUsed
		}

		return nil, nil, err
	}

	return staffFromRow(row), person, nil
}

func GetStaffByID(ctx context.Context, id int64) (*Staff, error) {
	row, err := queries.GetStaffByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStaffNotFound
	}

	if err != nil {
		return nil, err
	}

	return staffFromRow(row), nil
}

func ListStaffByPersonID(ctx context.Context, personID int64) ([]*Staff, error) {
	rows, err := queries.ListStaffByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}
	out := make([]*Staff, 0, len(rows))
	for _, r := range rows {
		out = append(out, staffFromRow(r))
	}
	return out, nil
}

func ListStaffByScope(ctx context.Context, scopeNodeID int64, limit, offset int32) ([]*Staff, error) {
	if limit <= 0 {
		limit = 200
	}

	rows, err := queries.ListStaffByScope(ctx, dbpeople.ListStaffByScopeParams{
		ScopeNodeID: scopeNodeID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*Staff, 0, len(rows))
	for _, r := range rows {
		out = append(out, staffFromRow(r))
	}

	return out, nil
}

func SoftDeleteStaff(ctx context.Context, id int64) error {
	return queries.SoftDeleteStaff(ctx, id)
}

func staffFromRow(r dbpeople.Staff) *Staff {
	s := &Staff{
		ID:          r.ID,
		PersonID:    r.PersonID,
		ScopeNodeID: r.ScopeNodeID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	if r.Position.Valid {
		s.Position = r.Position.String
	}

	if r.StaffCode.Valid {
		s.StaffCode = r.StaffCode.String
	}

	if r.HireDate.Valid {
		t := r.HireDate.Time
		s.HireDate = &t
	}

	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &s.Metadata)
	}

	if s.Metadata == nil {
		s.Metadata = map[string]any{}
	}

	return s
}
