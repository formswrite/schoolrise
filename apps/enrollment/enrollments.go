package enrollment

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"encore.app/apps/enrollment/dbenrollment"
)

const (
	StatusActive       = "active"
	StatusTransferred  = "transferred"
	StatusDropped      = "dropped"
	StatusGraduated    = "graduated"
	StatusReinstated   = "reinstated"
)

const (
	EventCreated      = "created"
	EventTransferred  = "transferred"
	EventDropped      = "dropped"
	EventGraduated    = "graduated"
	EventReinstated   = "reinstated"
)

var (
	ErrEnrollmentNotFound        = errors.New("enrollment: not found")
	ErrInvalidEnrollmentInput    = errors.New("enrollment: invalid input")
	ErrAlreadyActiveEnrollment   = errors.New("enrollment: student already actively enrolled this period")
	ErrSameInstitutionTransfer   = errors.New("enrollment: cannot transfer to the same institution")
	ErrEnrollmentNotActive       = errors.New("enrollment: enrollment is not active")
	ErrInvalidStatusTransition   = errors.New("enrollment: invalid status transition")
)

type Enrollment struct {
	ID            int64
	StudentID     int64
	InstitutionID int64
	PeriodID      int64
	Status        string
	EnrolledOn    time.Time
	EndedOn       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Event struct {
	ID                 int64
	EnrollmentID       int64
	Kind               string
	FromInstitutionID  *int64
	ToInstitutionID    *int64
	Note               string
	OccurredAt         time.Time
}

type CreateEnrollmentParams struct {
	StudentID     int64
	InstitutionID int64
	PeriodID      int64
	EnrolledOn    time.Time
	Note          string
}

func CreateEnrollment(ctx context.Context, p CreateEnrollmentParams) (*Enrollment, error) {
	if p.StudentID <= 0 || p.InstitutionID <= 0 || p.PeriodID <= 0 || p.EnrolledOn.IsZero() {
		return nil, ErrInvalidEnrollmentInput
	}

	row, err := queries.CreateEnrollment(ctx, dbenrollment.CreateEnrollmentParams{
		StudentID:     p.StudentID,
		InstitutionID: p.InstitutionID,
		PeriodID:      p.PeriodID,
		EnrolledOn:    p.EnrolledOn,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyActiveEnrollment
		}
		return nil, err
	}

	if _, err := queries.CreateEnrollmentEvent(ctx, dbenrollment.CreateEnrollmentEventParams{
		EnrollmentID:      row.ID,
		Kind:              EventCreated,
		ToInstitutionID:   nullableInt64(p.InstitutionID),
		Note:              nullableString(p.Note),
	}); err != nil {
		return nil, err
	}

	return enrollmentFromRow(row), nil
}

func GetEnrollmentByID(ctx context.Context, id int64) (*Enrollment, error) {
	row, err := queries.GetEnrollmentByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEnrollmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return enrollmentFromRow(row), nil
}

func GetActiveEnrollment(ctx context.Context, studentID, periodID int64) (*Enrollment, error) {
	row, err := queries.GetActiveEnrollment(ctx, dbenrollment.GetActiveEnrollmentParams{
		StudentID: studentID, PeriodID: periodID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEnrollmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return enrollmentFromRow(row), nil
}

func ListActiveStudentsInScope(ctx context.Context, institutionIDs []int64, periodID int64) (map[int64]int64, error) {
	if len(institutionIDs) == 0 {
		return map[int64]int64{}, nil
	}
	rows, err := queries.ListActiveStudentIDsForInstitutions(ctx, dbenrollment.ListActiveStudentIDsForInstitutionsParams{
		InstitutionIds: institutionIDs,
		PeriodID:       periodID,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for _, r := range rows {
		out[r.StudentID] = r.InstitutionID
	}
	return out, nil
}

func ListEnrollmentsByInstitution(ctx context.Context, institutionID, periodID int64, includeInactive bool) ([]*Enrollment, error) {
	if includeInactive {
		rows, err := queries.ListEnrollmentsByInstitution(ctx, dbenrollment.ListEnrollmentsByInstitutionParams{
			InstitutionID: institutionID, PeriodID: periodID,
		})
		if err != nil {
			return nil, err
		}
		return enrollmentsFromRows(rows), nil
	}
	rows, err := queries.ListActiveEnrollmentsByInstitution(ctx, dbenrollment.ListActiveEnrollmentsByInstitutionParams{
		InstitutionID: institutionID, PeriodID: periodID,
	})
	if err != nil {
		return nil, err
	}
	return enrollmentsFromRows(rows), nil
}

func ListEnrollmentsByStudent(ctx context.Context, studentID int64) ([]*Enrollment, error) {
	rows, err := queries.ListEnrollmentsByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return enrollmentsFromRows(rows), nil
}

type DropParams struct {
	EnrollmentID int64
	EndedOn      time.Time
	Note         string
}

func DropEnrollment(ctx context.Context, p DropParams) (*Enrollment, error) {
	current, err := GetEnrollmentByID(ctx, p.EnrollmentID)
	if err != nil {
		return nil, err
	}
	if current.Status != StatusActive {
		return nil, ErrEnrollmentNotActive
	}

	if p.EndedOn.IsZero() {
		p.EndedOn = time.Now().UTC()
	}

	row, err := queries.SetEnrollmentStatus(ctx, dbenrollment.SetEnrollmentStatusParams{
		ID:      p.EnrollmentID,
		Status:  StatusDropped,
		EndedOn: nullableDate(&p.EndedOn),
	})
	if err != nil {
		return nil, err
	}

	if _, err := queries.CreateEnrollmentEvent(ctx, dbenrollment.CreateEnrollmentEventParams{
		EnrollmentID:      p.EnrollmentID,
		Kind:              EventDropped,
		FromInstitutionID: nullableInt64(current.InstitutionID),
		Note:              nullableString(p.Note),
	}); err != nil {
		return nil, err
	}

	return enrollmentFromRow(row), nil
}

type TransferParams struct {
	StudentID         int64
	PeriodID          int64
	ToInstitutionID   int64
	EffectiveOn       time.Time
	Note              string
}

type TransferResult struct {
	Closed *Enrollment
	Opened *Enrollment
}

func TransferEnrollment(ctx context.Context, p TransferParams) (*TransferResult, error) {
	if p.StudentID <= 0 || p.PeriodID <= 0 || p.ToInstitutionID <= 0 {
		return nil, ErrInvalidEnrollmentInput
	}
	if p.EffectiveOn.IsZero() {
		p.EffectiveOn = time.Now().UTC()
	}

	current, err := GetActiveEnrollment(ctx, p.StudentID, p.PeriodID)
	if err != nil {
		return nil, err
	}
	if current.InstitutionID == p.ToInstitutionID {
		return nil, ErrSameInstitutionTransfer
	}

	closedRow, err := queries.SetEnrollmentStatus(ctx, dbenrollment.SetEnrollmentStatusParams{
		ID:      current.ID,
		Status:  StatusTransferred,
		EndedOn: nullableDate(&p.EffectiveOn),
	})
	if err != nil {
		return nil, err
	}

	if _, err := queries.CreateEnrollmentEvent(ctx, dbenrollment.CreateEnrollmentEventParams{
		EnrollmentID:      current.ID,
		Kind:              EventTransferred,
		FromInstitutionID: nullableInt64(current.InstitutionID),
		ToInstitutionID:   nullableInt64(p.ToInstitutionID),
		Note:              nullableString(p.Note),
	}); err != nil {
		return nil, err
	}

	openedRow, err := queries.CreateEnrollment(ctx, dbenrollment.CreateEnrollmentParams{
		StudentID:     p.StudentID,
		InstitutionID: p.ToInstitutionID,
		PeriodID:      p.PeriodID,
		EnrolledOn:    p.EffectiveOn,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyActiveEnrollment
		}
		return nil, err
	}

	if _, err := queries.CreateEnrollmentEvent(ctx, dbenrollment.CreateEnrollmentEventParams{
		EnrollmentID:      openedRow.ID,
		Kind:              EventCreated,
		FromInstitutionID: nullableInt64(current.InstitutionID),
		ToInstitutionID:   nullableInt64(p.ToInstitutionID),
		Note:              nullableString(p.Note),
	}); err != nil {
		return nil, err
	}

	return &TransferResult{
		Closed: enrollmentFromRow(closedRow),
		Opened: enrollmentFromRow(openedRow),
	}, nil
}

func ListEvents(ctx context.Context, enrollmentID int64) ([]*Event, error) {
	rows, err := queries.ListEnrollmentEvents(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}
	out := make([]*Event, 0, len(rows))
	for _, r := range rows {
		out = append(out, eventFromRow(r))
	}
	return out, nil
}

func enrollmentFromRow(r dbenrollment.Enrollment) *Enrollment {
	e := &Enrollment{
		ID:            r.ID,
		StudentID:     r.StudentID,
		InstitutionID: r.InstitutionID,
		PeriodID:      r.PeriodID,
		Status:        r.Status,
		EnrolledOn:    r.EnrolledOn,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
	if r.EndedOn.Valid {
		t := r.EndedOn.Time
		e.EndedOn = &t
	}
	return e
}

func enrollmentsFromRows(rows []dbenrollment.Enrollment) []*Enrollment {
	out := make([]*Enrollment, 0, len(rows))
	for _, r := range rows {
		out = append(out, enrollmentFromRow(r))
	}
	return out
}

func eventFromRow(r dbenrollment.EnrollmentEvent) *Event {
	e := &Event{
		ID:           r.ID,
		EnrollmentID: r.EnrollmentID,
		Kind:         r.Kind,
		OccurredAt:   r.OccurredAt,
	}
	if r.FromInstitutionID.Valid {
		v := r.FromInstitutionID.Int64
		e.FromInstitutionID = &v
	}
	if r.ToInstitutionID.Valid {
		v := r.ToInstitutionID.Int64
		e.ToInstitutionID = &v
	}
	if r.Note.Valid {
		e.Note = r.Note.String
	}
	return e
}

func nullableInt64(v int64) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}

func nullableDate(t *time.Time) sql.NullTime {
	if t == nil || t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullableString(s string) sql.NullString {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
