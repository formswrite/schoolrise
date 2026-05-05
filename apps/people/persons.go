package people

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"encore.app/apps/people/dbpeople"
)

var (
	ErrPersonNotFound  = errors.New("people: person not found")
	ErrInvalidPersonInput = errors.New("people: invalid person input")
)

type Person struct {
	ID          int64
	FullName    string
	GivenName   string
	FamilyName  string
	DateOfBirth *time.Time
	Gender      string
	Email       string
	Phone       string
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreatePersonParams struct {
	FullName    string
	GivenName   string
	FamilyName  string
	DateOfBirth *time.Time
	Gender      string
	Email       string
	Phone       string
	Metadata    map[string]any
}

func CreatePerson(ctx context.Context, p CreatePersonParams) (*Person, error) {
	if strings.TrimSpace(p.FullName) == "" {
		return nil, ErrInvalidPersonInput
	}

	metaBytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, err
	}

	if string(metaBytes) == "null" {
		metaBytes = []byte("{}")
	}

	row, err := queries.CreatePerson(ctx, dbpeople.CreatePersonParams{
		FullName:    p.FullName,
		GivenName:   nullableString(p.GivenName),
		FamilyName:  nullableString(p.FamilyName),
		DateOfBirth: nullableTime(p.DateOfBirth),
		Gender:      nullableString(p.Gender),
		Email:       nullableString(p.Email),
		Phone:       nullableString(p.Phone),
		Metadata:    metaBytes,
	})
	if err != nil {
		return nil, err
	}

	return personFromRow(row), nil
}

func GetPersonByEmail(ctx context.Context, email string) (*Person, error) {
	if strings.TrimSpace(email) == "" {
		return nil, ErrPersonNotFound
	}
	row, err := queries.GetPersonByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPersonNotFound
	}
	if err != nil {
		return nil, err
	}
	return personFromRow(row), nil
}

func GetPersonByID(ctx context.Context, id int64) (*Person, error) {
	row, err := queries.GetPersonByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPersonNotFound
	}

	if err != nil {
		return nil, err
	}

	return personFromRow(row), nil
}

func personFromRow(r dbpeople.Person) *Person {
	p := &Person{
		ID:        r.ID,
		FullName:  r.FullName,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}

	if r.GivenName.Valid {
		p.GivenName = r.GivenName.String
	}

	if r.FamilyName.Valid {
		p.FamilyName = r.FamilyName.String
	}

	if r.DateOfBirth.Valid {
		t := r.DateOfBirth.Time
		p.DateOfBirth = &t
	}

	if r.Gender.Valid {
		p.Gender = r.Gender.String
	}

	if r.Email.Valid {
		p.Email = r.Email.String
	}

	if r.Phone.Valid {
		p.Phone = r.Phone.String
	}

	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &p.Metadata)
	}

	if p.Metadata == nil {
		p.Metadata = map[string]any{}
	}

	return p
}

func nullableString(s string) sql.NullString {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: s, Valid: true}
}

func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *t, Valid: true}
}
