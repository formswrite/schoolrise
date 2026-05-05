package setup

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"encore.dev/beta/errs"

	"encore.app/apps/tenancy"
)

type ImportSchoolsRequest struct {
	SessionToken string
	CSV          string
}

type ImportSchoolsResponse struct {
	Imported int
	Errors   []string
}

//encore:api public method=POST path=/v1/setup/schools/import
func (s *Service) ImportSchoolsAPI(ctx context.Context, req *ImportSchoolsRequest) (*ImportSchoolsResponse, error) {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.CSV) == "" {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "CSV body required"}
	}

	imported, rowErrors := importSchoolsCSV(ctx, req.CSV)

	payload, _ := json.Marshal(map[string]any{"imported": imported, "row_errors": rowErrors})
	if err := markStepComplete(ctx, "schools", payload); err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not mark progress"}
	}

	return &ImportSchoolsResponse{Imported: imported, Errors: rowErrors}, nil
}

type SkipSchoolsRequest struct {
	SessionToken string
}

//encore:api public method=POST path=/v1/setup/schools/skip
func (s *Service) SkipSchoolsAPI(ctx context.Context, req *SkipSchoolsRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	if err := markStepSkipped(ctx, "schools"); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark step skipped"}
	}

	return nil
}

func importSchoolsCSV(ctx context.Context, raw string) (int, []string) {
	reader := csv.NewReader(strings.NewReader(raw))
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		return 0, []string{fmt.Sprintf("csv parse error: %v", err)}
	}

	if len(rows) <= 1 {
		return 0, nil
	}

	nodesByCode := map[string]int64{}

	var (
		imported int
		errors   []string
	)

	for i, row := range rows[1:] {
		rowNum := i + 2

		if len(row) < 4 {
			errors = append(errors, fmt.Sprintf("row %d: expected 4 columns, got %d", rowNum, len(row)))

			continue
		}

		parentCode := strings.TrimSpace(row[0])
		level := strings.TrimSpace(row[1])
		code := strings.TrimSpace(row[2])
		label := strings.TrimSpace(row[3])

		params := tenancy.CreateNodeParams{Level: level, Code: code, Label: label}

		if parentCode != "" {
			parentID, ok := nodesByCode[parentCode]
			if !ok {
				errors = append(errors, fmt.Sprintf("row %d: unknown parent code %q (parent must appear earlier in CSV)", rowNum, parentCode))

				continue
			}

			params.ParentID = &parentID
		}

		node, err := tenancy.CreateNode(ctx, params)
		if err != nil {
			errors = append(errors, fmt.Sprintf("row %d: %v", rowNum, err))

			continue
		}

		nodesByCode[code] = node.ID
		imported++
	}

	return imported, errors
}
