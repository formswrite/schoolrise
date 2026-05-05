package imports

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"encore.app/apps/imports/dbimports"
	"encore.app/apps/tenancy"
)

var schoolRequiredHeaders = []string{"code", "label", "parent_id"}

type SchoolRow struct {
	Code     string
	Label    string
	ParentID int64
}

type ImportSchoolsParams struct {
	CSVData   string
	DryRun    bool
	CreatedBy int64
}

func ImportSchools(ctx context.Context, p ImportSchoolsParams) (*Job, error) {
	if strings.TrimSpace(p.CSVData) == "" {
		return nil, ErrEmptyCSV
	}

	job, err := queries.CreateJob(ctx, dbimports.CreateJobParams{
		Kind:          KindSchools,
		InstitutionID: sql.NullInt64{},
		Status:        StatusRunning,
		DryRun:        p.DryRun,
		CreatedBy:     sql.NullInt64{Int64: p.CreatedBy, Valid: p.CreatedBy > 0},
	})
	if err != nil {
		return nil, err
	}

	rows, headerErr := parseSchoolCSV(p.CSVData)
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
		row, err := raw.parseSchool()
		if err != nil {
			failed++
			_ = recordRowError(ctx, job.ID, rowNum, "", err.Error(), raw.fields)
			continue
		}

		if !p.DryRun {
			if _, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
				ParentID: &row.ParentID,
				Level:    tenancy.LevelInstitution,
				Code:     row.Code,
				Label:    row.Label,
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

func (r rawRow) parseSchool() (*SchoolRow, error) {
	code := strings.TrimSpace(r.fields["code"])
	label := strings.TrimSpace(r.fields["label"])
	if code == "" || label == "" {
		return nil, fmt.Errorf("code and label are required")
	}
	parentStr := strings.TrimSpace(r.fields["parent_id"])
	if parentStr == "" {
		return nil, fmt.Errorf("parent_id is required")
	}
	var parentID int64
	if _, err := fmt.Sscanf(parentStr, "%d", &parentID); err != nil || parentID <= 0 {
		return nil, fmt.Errorf("parent_id must be a positive integer")
	}
	return &SchoolRow{Code: code, Label: label, ParentID: parentID}, nil
}

func parseSchoolCSV(data string) ([]rawRow, error) {
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
	for _, req := range schoolRequiredHeaders {
		if _, ok := headerIdx[req]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrMissingHeaders, req)
		}
	}

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
		for _, h := range schoolRequiredHeaders {
			if idx, ok := headerIdx[h]; ok && idx < len(rec) {
				fields[h] = rec[idx]
			}
		}
		out = append(out, rawRow{fields: fields})
	}
	return out, nil
}
