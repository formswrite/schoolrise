package people

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/people/dbpeople"
)

var (
	ErrStudentNotFound       = errors.New("people: student not found")
	ErrStudentCodeAlreadyUsed = errors.New("people: student code already used in this institution")
	ErrInvalidStudentInput   = errors.New("people: invalid student input")
)

type Student struct {
	ID             int64
	PersonID       int64
	InstitutionID  int64
	StudentCode    string
	EnrollmentDate *time.Time
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateStudentParams struct {
	Person         CreatePersonParams
	InstitutionID  int64
	StudentCode    string
	EnrollmentDate *time.Time
	Metadata       map[string]any
}

func CreateStudent(ctx context.Context, p CreateStudentParams) (*Student, *Person, error) {
	if p.InstitutionID <= 0 {
		return nil, nil, ErrInvalidStudentInput
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

	row, err := queries.CreateStudent(ctx, dbpeople.CreateStudentParams{
		PersonID:       person.ID,
		InstitutionID:  p.InstitutionID,
		StudentCode:    nullableString(p.StudentCode),
		EnrollmentDate: nullableTime(p.EnrollmentDate),
		Metadata:       metaBytes,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, nil, ErrStudentCodeAlreadyUsed
		}

		return nil, nil, err
	}

	return studentFromRow(row), person, nil
}

func GetStudentByID(ctx context.Context, id int64) (*Student, error) {
	row, err := queries.GetStudentByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStudentNotFound
	}

	if err != nil {
		return nil, err
	}

	return studentFromRow(row), nil
}

func ListStudentsByInstitution(ctx context.Context, institutionID int64, limit, offset int32) ([]*Student, error) {
	if limit <= 0 {
		limit = 200
	}

	rows, err := queries.ListStudentsByInstitution(ctx, dbpeople.ListStudentsByInstitutionParams{
		InstitutionID: institutionID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*Student, 0, len(rows))
	for _, r := range rows {
		out = append(out, studentFromRow(r))
	}

	return out, nil
}

type StudentWithPerson struct {
	Student *Student
	Person  *Person
}

func ListStudentsByInstitutionWithPerson(ctx context.Context, institutionID int64, limit, offset int32) ([]StudentWithPerson, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := queries.ListStudentsByInstitutionWithPerson(ctx, dbpeople.ListStudentsByInstitutionWithPersonParams{
		InstitutionID: institutionID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]StudentWithPerson, 0, len(rows))
	for _, r := range rows {
		s := &Student{
			ID:            r.StudentID,
			PersonID:      r.PersonID,
			InstitutionID: r.InstitutionID,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		}
		if r.StudentCode.Valid {
			s.StudentCode = r.StudentCode.String
		}
		if r.EnrollmentDate.Valid {
			t := r.EnrollmentDate.Time
			s.EnrollmentDate = &t
		}
		if len(r.Metadata) > 0 {
			_ = json.Unmarshal(r.Metadata, &s.Metadata)
		}
		p := &Person{
			ID:        r.PersonID,
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
		out = append(out, StudentWithPerson{Student: s, Person: p})
	}
	return out, nil
}

func ListStudentsByInstitutions(ctx context.Context, institutionIDs []int64, limit, offset int32) ([]*Student, error) {
	if limit <= 0 {
		limit = 200
	}

	if len(institutionIDs) == 0 {
		return nil, nil
	}

	rows, err := queries.ListStudentsByInstitutions(ctx, dbpeople.ListStudentsByInstitutionsParams{
		Column1: institutionIDs,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*Student, 0, len(rows))
	for _, r := range rows {
		out = append(out, studentFromRow(r))
	}

	return out, nil
}

func GetStudentGendersByIDs(ctx context.Context, studentIDs []int64) (map[int64]string, error) {
	if len(studentIDs) == 0 {
		return map[int64]string{}, nil
	}
	rows, err := queries.GetStudentGendersByIDs(ctx, studentIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]string, len(rows))
	for _, r := range rows {
		g := ""
		if r.Gender.Valid {
			g = r.Gender.String
		}
		out[r.StudentID] = g
	}
	return out, nil
}

func SoftDeleteStudent(ctx context.Context, id int64) error {
	return queries.SoftDeleteStudent(ctx, id)
}

func GetSchoolsForStudents(ctx context.Context, studentIDs []int64) (map[int64]int64, error) {
	if len(studentIDs) == 0 {
		return map[int64]int64{}, nil
	}
	rows, err := queries.GetSchoolsForStudents(ctx, studentIDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for _, r := range rows {
		out[r.StudentID] = r.InstitutionID
	}
	return out, nil
}

func studentFromRow(r dbpeople.Student) *Student {
	s := &Student{
		ID:            r.ID,
		PersonID:      r.PersonID,
		InstitutionID: r.InstitutionID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}

	if r.StudentCode.Valid {
		s.StudentCode = r.StudentCode.String
	}

	if r.EnrollmentDate.Valid {
		t := r.EnrollmentDate.Time
		s.EnrollmentDate = &t
	}

	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &s.Metadata)
	}

	if s.Metadata == nil {
		s.Metadata = map[string]any{}
	}

	return s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
