package imports

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"encore.app/apps/imports/dbimports"
	"encore.app/apps/people"
)

const (
	KindStudents = "students"
	KindStaff    = "staff"
	KindSchools  = "schools"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

var (
	ErrInvalidImportInput = errors.New("imports: invalid input")
	ErrEmptyCSV           = errors.New("imports: csv is empty")
	ErrMissingHeaders     = errors.New("imports: required header missing")
)

var studentRequiredHeaders = []string{"full_name"}
var studentOptionalHeaders = []string{"given_name", "family_name", "gender", "date_of_birth", "email", "phone", "student_code", "enrollment_date"}

type StudentRow struct {
	FullName       string
	GivenName      string
	FamilyName     string
	Gender         string
	DateOfBirth    *time.Time
	Email          string
	Phone          string
	StudentCode    string
	EnrollmentDate *time.Time
}

type RowError struct {
	RowNumber int    `json:"row_number"`
	Field     string `json:"field,omitempty"`
	Error     string `json:"error"`
}

type Job struct {
	ID            int64
	Kind          string
	InstitutionID int64
	Status        string
	TotalRows     int
	Succeeded     int
	Failed        int
	DryRun        bool
	Errors        []RowError
	CreatedAt     time.Time
	CompletedAt   *time.Time
}

type ImportStudentsParams struct {
	InstitutionID int64
	CSVData       string
	DryRun        bool
	CreatedBy     int64
}

func ImportStudents(ctx context.Context, p ImportStudentsParams) (*Job, error) {
	if p.InstitutionID <= 0 {
		return nil, ErrInvalidImportInput
	}
	if strings.TrimSpace(p.CSVData) == "" {
		return nil, ErrEmptyCSV
	}

	job, err := queries.CreateJob(ctx, dbimports.CreateJobParams{
		Kind:          KindStudents,
		InstitutionID: sql.NullInt64{Int64: p.InstitutionID, Valid: true},
		Status:        StatusRunning,
		DryRun:        p.DryRun,
		CreatedBy:     sql.NullInt64{Int64: p.CreatedBy, Valid: p.CreatedBy > 0},
	})
	if err != nil {
		return nil, err
	}

	rows, headerErr := parseStudentCSV(p.CSVData)
	if headerErr != nil {
		_ = recordRowError(ctx, job.ID, 0, "", headerErr.Error(), nil)
		_, _ = queries.UpdateJobResult(ctx, dbimports.UpdateJobResultParams{
			ID: job.ID, Status: StatusFailed, TotalRows: 0, SucceededRows: 0, FailedRows: 1,
			Summary: jsonString(`{"error":"header_validation"}`),
		})
		return collectJob(ctx, job.ID)
	}

	total := int32(len(rows))
	var succeeded, failed int32

	for i, raw := range rows {
		rowNum := int32(i + 2)
		row, err := raw.parse()
		if err != nil {
			failed++
			_ = recordRowError(ctx, job.ID, rowNum, "", err.Error(), raw.fields)
			continue
		}

		if !p.DryRun {
			if _, _, err := people.CreateStudent(ctx, people.CreateStudentParams{
				InstitutionID: p.InstitutionID,
				StudentCode:   row.StudentCode,
				EnrollmentDate: row.EnrollmentDate,
				Person: people.CreatePersonParams{
					FullName:    row.FullName,
					GivenName:   row.GivenName,
					FamilyName:  row.FamilyName,
					Gender:      row.Gender,
					DateOfBirth: row.DateOfBirth,
					Email:       row.Email,
					Phone:       row.Phone,
				},
			}); err != nil {
				failed++
				_ = recordRowError(ctx, job.ID, rowNum, "", err.Error(), raw.fields)
				continue
			}
		}
		succeeded++
	}

	finalStatus := StatusCompleted
	if failed > 0 && succeeded == 0 {
		finalStatus = StatusFailed
	}

	summary, _ := json.Marshal(map[string]any{
		"dry_run":   p.DryRun,
		"succeeded": succeeded,
		"failed":    failed,
	})

	if _, err := queries.UpdateJobResult(ctx, dbimports.UpdateJobResultParams{
		ID: job.ID, Status: finalStatus,
		TotalRows: total, SucceededRows: succeeded, FailedRows: failed,
		Summary: summary,
	}); err != nil {
		return nil, err
	}

	return collectJob(ctx, job.ID)
}

func GetJob(ctx context.Context, id int64) (*Job, error) {
	return collectJob(ctx, id)
}

type rawRow struct {
	fields map[string]string
}

func (r rawRow) parse() (*StudentRow, error) {
	full := strings.TrimSpace(r.fields["full_name"])
	if full == "" {
		return nil, fmt.Errorf("full_name is required")
	}

	out := &StudentRow{
		FullName:    full,
		GivenName:   strings.TrimSpace(r.fields["given_name"]),
		FamilyName:  strings.TrimSpace(r.fields["family_name"]),
		Gender:      strings.TrimSpace(r.fields["gender"]),
		Email:       strings.TrimSpace(r.fields["email"]),
		Phone:       strings.TrimSpace(r.fields["phone"]),
		StudentCode: strings.TrimSpace(r.fields["student_code"]),
	}

	if dob := strings.TrimSpace(r.fields["date_of_birth"]); dob != "" {
		t, err := time.Parse("2006-01-02", dob)
		if err != nil {
			return nil, fmt.Errorf("date_of_birth must be YYYY-MM-DD")
		}
		out.DateOfBirth = &t
	}

	if ed := strings.TrimSpace(r.fields["enrollment_date"]); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			return nil, fmt.Errorf("enrollment_date must be YYYY-MM-DD")
		}
		out.EnrollmentDate = &t
	}

	return out, nil
}

func parseStudentCSV(data string) ([]rawRow, error) {
	r := csv.NewReader(strings.NewReader(data))
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err == io.EOF {
		return nil, ErrEmptyCSV
	}
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	headerIdx := map[string]int{}
	for i, h := range header {
		headerIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	for _, req := range studentRequiredHeaders {
		if _, ok := headerIdx[req]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingHeaders, req)
		}
	}

	allHeaders := append([]string{}, studentRequiredHeaders...)
	allHeaders = append(allHeaders, studentOptionalHeaders...)

	out := []rawRow{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", len(out)+2, err)
		}
		fields := map[string]string{}
		for _, h := range allHeaders {
			if idx, ok := headerIdx[h]; ok && idx < len(rec) {
				fields[h] = rec[idx]
			}
		}
		out = append(out, rawRow{fields: fields})
	}

	return out, nil
}

func recordRowError(ctx context.Context, jobID int64, rowNumber int32, field, msg string, raw map[string]string) error {
	rawJSON := jsonString("{}")
	if raw != nil {
		if b, err := json.Marshal(raw); err == nil {
			rawJSON = b
		}
	}
	return queries.AddRowError(ctx, dbimports.AddRowErrorParams{
		JobID:     jobID,
		RowNumber: rowNumber,
		Field:     sql.NullString{String: field, Valid: field != ""},
		Error:     msg,
		RawData:   rawJSON,
	})
}

func collectJob(ctx context.Context, jobID int64) (*Job, error) {
	row, err := queries.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	errRows, err := queries.ListRowErrors(ctx, dbimports.ListRowErrorsParams{
		JobID: jobID, Limit: 1000, Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	out := &Job{
		ID:        row.ID,
		Kind:      row.Kind,
		Status:    row.Status,
		TotalRows: int(row.TotalRows),
		Succeeded: int(row.SucceededRows),
		Failed:    int(row.FailedRows),
		DryRun:    row.DryRun,
		CreatedAt: row.CreatedAt,
	}
	if row.InstitutionID.Valid {
		out.InstitutionID = row.InstitutionID.Int64
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		out.CompletedAt = &t
	}
	for _, e := range errRows {
		field := ""
		if e.Field.Valid {
			field = e.Field.String
		}
		out.Errors = append(out.Errors, RowError{
			RowNumber: int(e.RowNumber),
			Field:     field,
			Error:     e.Error,
		})
	}
	return out, nil
}

func jsonString(s string) []byte { return []byte(s) }
