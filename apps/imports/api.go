package imports

import (
	"context"
	"errors"
	"strconv"
	"time"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type JobDTO struct {
	ID            int64      `json:"id"`
	Kind          string     `json:"kind"`
	InstitutionID int64      `json:"institution_id,omitempty"`
	Status        string     `json:"status"`
	TotalRows     int        `json:"total_rows"`
	Succeeded     int        `json:"succeeded"`
	Failed        int        `json:"failed"`
	DryRun        bool       `json:"dry_run"`
	Errors        []RowError `json:"errors"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type ImportStudentsAPIRequest struct {
	InstitutionID int64  `json:"institution_id"`
	CSVData       string `json:"csv_data"`
	DryRun        bool   `json:"dry_run"`
}

//encore:api auth method=POST path=/v1/imports/students
func (s *Service) ImportStudentsAPI(ctx context.Context, req *ImportStudentsAPIRequest) (*JobDTO, error) {
	if req.InstitutionID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "institution_id required"}
	}
	uid, _ := encauth.UserID()
	createdBy, _ := strconv.ParseInt(string(uid), 10, 64)

	job, err := ImportStudents(ctx, ImportStudentsParams{
		InstitutionID: req.InstitutionID,
		CSVData:       req.CSVData,
		DryRun:        req.DryRun,
		CreatedBy:     createdBy,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return jobToDTO(job), nil
}

type ImportStaffAPIRequest struct {
	ScopeNodeID int64  `json:"scope_node_id"`
	CSVData     string `json:"csv_data"`
	DryRun      bool   `json:"dry_run"`
}

//encore:api auth method=POST path=/v1/imports/staff
func (s *Service) ImportStaffAPI(ctx context.Context, req *ImportStaffAPIRequest) (*JobDTO, error) {
	if req.ScopeNodeID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "scope_node_id required"}
	}
	uid, _ := encauth.UserID()
	createdBy, _ := strconv.ParseInt(string(uid), 10, 64)
	job, err := ImportStaff(ctx, ImportStaffParams{
		ScopeNodeID: req.ScopeNodeID, CSVData: req.CSVData, DryRun: req.DryRun, CreatedBy: createdBy,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return jobToDTO(job), nil
}

type ImportSchoolsAPIRequest struct {
	CSVData string `json:"csv_data"`
	DryRun  bool   `json:"dry_run"`
}

//encore:api auth method=POST path=/v1/imports/schools
func (s *Service) ImportSchoolsAPI(ctx context.Context, req *ImportSchoolsAPIRequest) (*JobDTO, error) {
	uid, _ := encauth.UserID()
	createdBy, _ := strconv.ParseInt(string(uid), 10, 64)
	job, err := ImportSchools(ctx, ImportSchoolsParams{
		CSVData: req.CSVData, DryRun: req.DryRun, CreatedBy: createdBy,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return jobToDTO(job), nil
}

//encore:api auth method=GET path=/v1/imports/jobs/:id
func (s *Service) GetJobAPI(ctx context.Context, id int64) (*JobDTO, error) {
	job, err := GetJob(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return jobToDTO(job), nil
}

func jobToDTO(j *Job) *JobDTO {
	return &JobDTO{
		ID:            j.ID,
		Kind:          j.Kind,
		InstitutionID: j.InstitutionID,
		Status:        j.Status,
		TotalRows:     j.TotalRows,
		Succeeded:     j.Succeeded,
		Failed:        j.Failed,
		DryRun:        j.DryRun,
		Errors:        j.Errors,
		CreatedAt:     j.CreatedAt,
		CompletedAt:   j.CompletedAt,
	}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrInvalidImportInput), errors.Is(err, ErrEmptyCSV):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	}
	return apierr.WrapInternal("imports", err)
}
