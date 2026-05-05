package imports

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"encore.app/apps/imports/dbimports"
	"encore.app/apps/people"
)

var staffRequiredHeaders = []string{"full_name"}
var staffOptionalHeaders = []string{"given_name", "family_name", "gender", "email", "phone", "position", "staff_code", "hire_date"}

type StaffRow struct {
	FullName   string
	GivenName  string
	FamilyName string
	Gender     string
	Email      string
	Phone      string
	Position   string
	StaffCode  string
	HireDate   *time.Time
}

type ImportStaffParams struct {
	ScopeNodeID int64
	CSVData     string
	DryRun      bool
	CreatedBy   int64
}

func ImportStaff(ctx context.Context, p ImportStaffParams) (*Job, error) {
	if p.ScopeNodeID <= 0 {
		return nil, ErrInvalidImportInput
	}
	if strings.TrimSpace(p.CSVData) == "" {
		return nil, ErrEmptyCSV
	}

	job, err := queries.CreateJob(ctx, dbimports.CreateJobParams{
		Kind:          KindStaff,
		InstitutionID: sql.NullInt64{Int64: p.ScopeNodeID, Valid: true},
		Status:        StatusRunning,
		DryRun:        p.DryRun,
		CreatedBy:     sql.NullInt64{Int64: p.CreatedBy, Valid: p.CreatedBy > 0},
	})
	if err != nil {
		return nil, err
	}

	rows, headerErr := parseStaffCSV(p.CSVData)
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
		row, err := raw.parseStaff()
		if err != nil {
			failed++
			_ = recordRowError(ctx, job.ID, rowNum, "", err.Error(), raw.fields)
			continue
		}

		if !p.DryRun {
			if _, _, err := people.CreateStaff(ctx, people.CreateStaffParams{
				ScopeNodeID: p.ScopeNodeID,
				Position:    row.Position,
				StaffCode:   row.StaffCode,
				HireDate:    row.HireDate,
				Person: people.CreatePersonParams{
					FullName:   row.FullName,
					GivenName:  row.GivenName,
					FamilyName: row.FamilyName,
					Gender:     row.Gender,
					Email:      row.Email,
					Phone:      row.Phone,
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
		"dry_run": p.DryRun, "succeeded": succeeded, "failed": failed,
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

func (r rawRow) parseStaff() (*StaffRow, error) {
	full := strings.TrimSpace(r.fields["full_name"])
	if full == "" {
		return nil, fmt.Errorf("full_name is required")
	}
	out := &StaffRow{
		FullName:   full,
		GivenName:  strings.TrimSpace(r.fields["given_name"]),
		FamilyName: strings.TrimSpace(r.fields["family_name"]),
		Gender:     strings.TrimSpace(r.fields["gender"]),
		Email:      strings.TrimSpace(r.fields["email"]),
		Phone:      strings.TrimSpace(r.fields["phone"]),
		Position:   strings.TrimSpace(r.fields["position"]),
		StaffCode:  strings.TrimSpace(r.fields["staff_code"]),
	}
	if hd := strings.TrimSpace(r.fields["hire_date"]); hd != "" {
		t, err := time.Parse("2006-01-02", hd)
		if err != nil {
			return nil, fmt.Errorf("hire_date must be YYYY-MM-DD")
		}
		out.HireDate = &t
	}
	return out, nil
}

func parseStaffCSV(data string) ([]rawRow, error) {
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
	for _, req := range staffRequiredHeaders {
		if _, ok := headerIdx[req]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingHeaders, req)
		}
	}
	allHeaders := append([]string{}, staffRequiredHeaders...)
	allHeaders = append(allHeaders, staffOptionalHeaders...)

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
