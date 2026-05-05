package enrollment

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type EnrollmentDTO struct {
	ID            int64      `json:"id"`
	StudentID     int64      `json:"student_id"`
	InstitutionID int64      `json:"institution_id"`
	PeriodID      int64      `json:"period_id"`
	Status        string     `json:"status"`
	EnrolledOn    time.Time  `json:"enrolled_on"`
	EndedOn       *time.Time `json:"ended_on,omitempty"`
}

type EventDTO struct {
	ID                int64     `json:"id"`
	Kind              string    `json:"kind"`
	FromInstitutionID *int64    `json:"from_institution_id,omitempty"`
	ToInstitutionID   *int64    `json:"to_institution_id,omitempty"`
	Note              string    `json:"note,omitempty"`
	OccurredAt        time.Time `json:"occurred_at"`
}

type CoverageDTO struct {
	ScopeNodeID   int64 `json:"scope_node_id"`
	PeriodID      int64 `json:"period_id"`
	TotalEnrolled int   `json:"total_enrolled"`
	Male          int   `json:"male"`
	Female        int   `json:"female"`
	Other         int   `json:"other"`
	Unknown       int   `json:"unknown"`
}

func toDTO(e *Enrollment) EnrollmentDTO {
	return EnrollmentDTO{
		ID:            e.ID,
		StudentID:     e.StudentID,
		InstitutionID: e.InstitutionID,
		PeriodID:      e.PeriodID,
		Status:        e.Status,
		EnrolledOn:    e.EnrolledOn,
		EndedOn:       e.EndedOn,
	}
}

type ListEnrollmentsRequest struct {
	InstitutionID   int64 `query:"institution_id"`
	PeriodID        int64 `query:"period_id"`
	IncludeInactive bool  `query:"include_inactive"`
}

type ListEnrollmentsResponse struct {
	Enrollments []EnrollmentDTO `json:"enrollments"`
}

//encore:api auth method=GET path=/v1/enrollment/enrollments
func (s *Service) ListEnrollmentsAPI(ctx context.Context, req *ListEnrollmentsRequest) (*ListEnrollmentsResponse, error) {
	if req.InstitutionID <= 0 || req.PeriodID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "institution_id and period_id are required"}
	}
	if err := requireNodeAccess(ctx, req.InstitutionID); err != nil {
		return nil, err
	}
	rows, err := ListEnrollmentsByInstitution(ctx, req.InstitutionID, req.PeriodID, req.IncludeInactive)
	if err != nil {
		return nil, apierr.WrapInternal("enrollment", err)
	}
	out := make([]EnrollmentDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDTO(r))
	}
	return &ListEnrollmentsResponse{Enrollments: out}, nil
}

type CreateEnrollmentAPIRequest struct {
	StudentID     int64  `json:"student_id"`
	InstitutionID int64  `json:"institution_id"`
	PeriodID      int64  `json:"period_id"`
	EnrolledOn    string `json:"enrolled_on"`
	Note          string `json:"note,omitempty"`
}

type CreateEnrollmentResponse struct {
	Enrollment EnrollmentDTO `json:"enrollment"`
}

//encore:api auth method=POST path=/v1/enrollment/enrollments
func (s *Service) CreateEnrollmentAPI(ctx context.Context, req *CreateEnrollmentAPIRequest) (*CreateEnrollmentResponse, error) {
	if err := requireNodeAccess(ctx, req.InstitutionID); err != nil {
		return nil, err
	}
	enrolledOn, err := time.Parse("2006-01-02", req.EnrolledOn)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "enrolled_on must be YYYY-MM-DD"}
	}
	row, err := CreateEnrollment(ctx, CreateEnrollmentParams{
		StudentID: req.StudentID, InstitutionID: req.InstitutionID, PeriodID: req.PeriodID,
		EnrolledOn: enrolledOn, Note: req.Note,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &CreateEnrollmentResponse{Enrollment: toDTO(row)}, nil
}

type DropEnrollmentAPIRequest struct {
	EnrollmentID int64  `json:"enrollment_id"`
	EndedOn      string `json:"ended_on,omitempty"`
	Note         string `json:"note,omitempty"`
}

//encore:api auth method=POST path=/v1/enrollment/drop
func (s *Service) DropEnrollmentAPI(ctx context.Context, req *DropEnrollmentAPIRequest) (*CreateEnrollmentResponse, error) {
	current, err := GetEnrollmentByID(ctx, req.EnrollmentID)
	if err != nil {
		return nil, mapErr(err)
	}
	if err := requireNodeAccess(ctx, current.InstitutionID); err != nil {
		return nil, err
	}

	var endedOn time.Time
	if req.EndedOn != "" {
		endedOn, err = time.Parse("2006-01-02", req.EndedOn)
		if err != nil {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "ended_on must be YYYY-MM-DD"}
		}
	}
	row, err := DropEnrollment(ctx, DropParams{EnrollmentID: req.EnrollmentID, EndedOn: endedOn, Note: req.Note})
	if err != nil {
		return nil, mapErr(err)
	}
	return &CreateEnrollmentResponse{Enrollment: toDTO(row)}, nil
}

type TransferAPIRequest struct {
	StudentID       int64  `json:"student_id"`
	PeriodID        int64  `json:"period_id"`
	ToInstitutionID int64  `json:"to_institution_id"`
	EffectiveOn     string `json:"effective_on,omitempty"`
	Note            string `json:"note,omitempty"`
}

type TransferAPIResponse struct {
	Closed EnrollmentDTO `json:"closed"`
	Opened EnrollmentDTO `json:"opened"`
}

//encore:api auth method=POST path=/v1/enrollment/transfers
func (s *Service) TransferAPI(ctx context.Context, req *TransferAPIRequest) (*TransferAPIResponse, error) {
	current, err := GetActiveEnrollment(ctx, req.StudentID, req.PeriodID)
	if err != nil {
		return nil, mapErr(err)
	}
	if err := requireNodeAccess(ctx, current.InstitutionID); err != nil {
		return nil, err
	}
	if err := requireNodeAccess(ctx, req.ToInstitutionID); err != nil {
		return nil, err
	}

	var effectiveOn time.Time
	if req.EffectiveOn != "" {
		effectiveOn, err = time.Parse("2006-01-02", req.EffectiveOn)
		if err != nil {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "effective_on must be YYYY-MM-DD"}
		}
	}
	res, err := TransferEnrollment(ctx, TransferParams{
		StudentID: req.StudentID, PeriodID: req.PeriodID, ToInstitutionID: req.ToInstitutionID,
		EffectiveOn: effectiveOn, Note: req.Note,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &TransferAPIResponse{Closed: toDTO(res.Closed), Opened: toDTO(res.Opened)}, nil
}

type CoverageRequest struct {
	ScopeNodeID int64 `query:"scope_node_id"`
	PeriodID    int64 `query:"period_id"`
}

//encore:api auth method=GET path=/v1/enrollment/coverage
func (s *Service) GetCoverageAPI(ctx context.Context, req *CoverageRequest) (*CoverageDTO, error) {
	if req.ScopeNodeID <= 0 || req.PeriodID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "scope_node_id and period_id are required"}
	}
	if err := requireNodeAccess(ctx, req.ScopeNodeID); err != nil {
		return nil, err
	}
	cov, err := ComputeCoverage(ctx, req.ScopeNodeID, req.PeriodID)
	if err != nil {
		return nil, mapErr(err)
	}
	return &CoverageDTO{
		ScopeNodeID:   cov.ScopeNodeID,
		PeriodID:      cov.PeriodID,
		TotalEnrolled: cov.TotalEnrolled,
		Male:          cov.Male,
		Female:        cov.Female,
		Other:         cov.Other,
		Unknown:       cov.Unknown,
	}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrEnrollmentNotFound):
		return &errs.Error{Code: errs.NotFound, Message: err.Error()}
	case errors.Is(err, ErrInvalidEnrollmentInput),
		errors.Is(err, ErrInvalidCoverageInput),
		errors.Is(err, ErrSameInstitutionTransfer),
		errors.Is(err, ErrEnrollmentNotActive):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	case errors.Is(err, ErrAlreadyActiveEnrollment):
		return &errs.Error{Code: errs.AlreadyExists, Message: err.Error()}
	}
	return apierr.WrapInternal("enrollment", err)
}
