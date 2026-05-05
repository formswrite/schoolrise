package academics

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"

	"encore.app/pkg/apierr"
)

type PeriodDTO struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	StartsOn  time.Time `json:"starts_on"`
	EndsOn    time.Time `json:"ends_on"`
	IsCurrent bool      `json:"is_current"`
}

type NiveauDTO struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Label     string `json:"label"`
	SortOrder int32  `json:"sort_order"`
}

type ClassDTO struct {
	ID            int64  `json:"id"`
	PeriodID      int64  `json:"period_id"`
	NiveauID      int64  `json:"niveau_id"`
	InstitutionID int64  `json:"institution_id"`
	Code          string `json:"code"`
	Label         string `json:"label"`
	Capacity      int32  `json:"capacity,omitempty"`
}

func periodToDTO(p *Period) PeriodDTO {
	return PeriodDTO{ID: p.ID, Code: p.Code, Label: p.Label, StartsOn: p.StartsOn, EndsOn: p.EndsOn, IsCurrent: p.IsCurrent}
}
func niveauToDTO(n *Niveau) NiveauDTO {
	return NiveauDTO{ID: n.ID, Code: n.Code, Label: n.Label, SortOrder: n.SortOrder}
}
func classToDTO(c *Class) ClassDTO {
	return ClassDTO{ID: c.ID, PeriodID: c.PeriodID, NiveauID: c.NiveauID, InstitutionID: c.InstitutionID, Code: c.Code, Label: c.Label, Capacity: c.Capacity}
}

type ListPeriodsResponse struct {
	Periods []PeriodDTO `json:"periods"`
}

//encore:api auth method=GET path=/v1/academics/periods
func (s *Service) ListPeriodsAPI(ctx context.Context) (*ListPeriodsResponse, error) {
	rows, err := ListPeriods(ctx)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]PeriodDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, periodToDTO(r))
	}
	return &ListPeriodsResponse{Periods: out}, nil
}

type CreatePeriodAPIRequest struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	StartsOn  string `json:"starts_on"`
	EndsOn    string `json:"ends_on"`
	IsCurrent bool   `json:"is_current"`
}

//encore:api auth method=POST path=/v1/academics/periods
func (s *Service) CreatePeriodAPI(ctx context.Context, req *CreatePeriodAPIRequest) (*PeriodDTO, error) {
	starts, err := time.Parse("2006-01-02", req.StartsOn)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "starts_on must be YYYY-MM-DD"}
	}
	ends, err := time.Parse("2006-01-02", req.EndsOn)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "ends_on must be YYYY-MM-DD"}
	}
	p, err := CreatePeriod(ctx, CreatePeriodParams{
		Code: req.Code, Label: req.Label, StartsOn: starts, EndsOn: ends, IsCurrent: req.IsCurrent,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := periodToDTO(p)
	return &out, nil
}

//encore:api auth method=POST path=/v1/academics/periods/:id/current
func (s *Service) SetCurrentPeriodAPI(ctx context.Context, id int64) (*PeriodDTO, error) {
	p, err := SetPeriodCurrent(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	out := periodToDTO(p)
	return &out, nil
}

//encore:api auth method=DELETE path=/v1/academics/periods/:id
func (s *Service) DeletePeriodAPI(ctx context.Context, id int64) error {
	if err := DeletePeriod(ctx, id); err != nil {
		return mapErr(err)
	}
	return nil
}

type ListNiveauxResponse struct {
	Niveaux []NiveauDTO `json:"niveaux"`
}

//encore:api auth method=GET path=/v1/academics/niveaux
func (s *Service) ListNiveauxAPI(ctx context.Context) (*ListNiveauxResponse, error) {
	rows, err := ListNiveaux(ctx)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]NiveauDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, niveauToDTO(r))
	}
	return &ListNiveauxResponse{Niveaux: out}, nil
}

type CreateNiveauAPIRequest struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	SortOrder int32  `json:"sort_order"`
}

//encore:api auth method=POST path=/v1/academics/niveaux
func (s *Service) CreateNiveauAPI(ctx context.Context, req *CreateNiveauAPIRequest) (*NiveauDTO, error) {
	n, err := CreateNiveau(ctx, CreateNiveauParams{Code: req.Code, Label: req.Label, SortOrder: req.SortOrder})
	if err != nil {
		return nil, mapErr(err)
	}
	out := niveauToDTO(n)
	return &out, nil
}

//encore:api auth method=DELETE path=/v1/academics/niveaux/:id
func (s *Service) DeleteNiveauAPI(ctx context.Context, id int64) error {
	if err := DeleteNiveau(ctx, id); err != nil {
		return mapErr(err)
	}
	return nil
}

type ListClassesRequest struct {
	InstitutionID int64 `query:"institution_id"`
	PeriodID      int64 `query:"period_id"`
}

type ListClassesResponse struct {
	Classes []ClassDTO `json:"classes"`
}

//encore:api auth method=GET path=/v1/academics/classes
func (s *Service) ListClassesAPI(ctx context.Context, req *ListClassesRequest) (*ListClassesResponse, error) {
	var rows []*Class
	var err error
	switch {
	case req.InstitutionID > 0:
		rows, err = ListClassesByInstitution(ctx, req.InstitutionID)
	case req.PeriodID > 0:
		rows, err = ListClassesByPeriod(ctx, req.PeriodID)
	default:
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "institution_id or period_id required"}
	}
	if err != nil {
		return nil, internal(err)
	}
	out := make([]ClassDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, classToDTO(r))
	}
	return &ListClassesResponse{Classes: out}, nil
}

type CreateClassAPIRequest struct {
	PeriodID      int64  `json:"period_id"`
	NiveauID      int64  `json:"niveau_id"`
	InstitutionID int64  `json:"institution_id"`
	Code          string `json:"code"`
	Label         string `json:"label"`
	Capacity      int32  `json:"capacity"`
}

//encore:api auth method=POST path=/v1/academics/classes
func (s *Service) CreateClassAPI(ctx context.Context, req *CreateClassAPIRequest) (*ClassDTO, error) {
	c, err := CreateClass(ctx, CreateClassParams{
		PeriodID: req.PeriodID, NiveauID: req.NiveauID, InstitutionID: req.InstitutionID,
		Code: req.Code, Label: req.Label, Capacity: req.Capacity,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	out := classToDTO(c)
	return &out, nil
}

//encore:api auth method=DELETE path=/v1/academics/classes/:id
func (s *Service) DeleteClassAPI(ctx context.Context, id int64) error {
	if err := DeleteClass(ctx, id); err != nil {
		return mapErr(err)
	}
	return nil
}

//encore:api auth method=GET path=/v1/academics/classes/:id
func (s *Service) GetClassAPI(ctx context.Context, id int64) (*ClassDTO, error) {
	c, err := GetClassByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	out := classToDTO(c)
	return &out, nil
}

type AddStudentRequest struct {
	StudentID int64 `json:"student_id"`
}

//encore:api auth method=POST path=/v1/academics/classes/:id/students
func (s *Service) AddStudentAPI(ctx context.Context, id int64, req *AddStudentRequest) error {
	if err := AddStudentToClass(ctx, id, req.StudentID); err != nil {
		return mapErr(err)
	}
	return nil
}

//encore:api auth method=DELETE path=/v1/academics/classes/:id/students/:studentID
func (s *Service) RemoveStudentAPI(ctx context.Context, id, studentID int64) error {
	if err := RemoveStudentFromClass(ctx, id, studentID); err != nil {
		return internal(err)
	}
	return nil
}

type AddStaffRequest struct {
	StaffID int64  `json:"staff_id"`
	Role    string `json:"role"`
}

//encore:api auth method=POST path=/v1/academics/classes/:id/staff
func (s *Service) AddStaffAPI(ctx context.Context, id int64, req *AddStaffRequest) error {
	if err := AddStaffToClass(ctx, id, req.StaffID, req.Role); err != nil {
		return mapErr(err)
	}
	return nil
}

//encore:api auth method=DELETE path=/v1/academics/classes/:id/staff/:staffID/:role
func (s *Service) RemoveStaffAPI(ctx context.Context, id, staffID int64, role string) error {
	if err := RemoveStaffFromClass(ctx, id, staffID, role); err != nil {
		return internal(err)
	}
	return nil
}

type ClassRosterDTO struct {
	StudentIDs []int64            `json:"student_ids"`
	Staff      []ClassStaffMember `json:"staff"`
}

//encore:api auth method=GET path=/v1/academics/classes/:id/roster
func (s *Service) GetRosterAPI(ctx context.Context, id int64) (*ClassRosterDTO, error) {
	students, err := ListClassStudents(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	staff, err := ListClassStaff(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	return &ClassRosterDTO{StudentIDs: students, Staff: staff}, nil
}

func internal(err error) error { return apierr.WrapInternal("academics", err) }

func mapErr(err error) error {
	switch {
	case errors.Is(err, ErrPeriodNotFound), errors.Is(err, ErrNiveauNotFound), errors.Is(err, ErrClassNotFound):
		return &errs.Error{Code: errs.NotFound, Message: err.Error()}
	case errors.Is(err, ErrPeriodCodeTaken), errors.Is(err, ErrNiveauCodeTaken), errors.Is(err, ErrClassCodeTaken):
		return &errs.Error{Code: errs.AlreadyExists, Message: err.Error()}
	case errors.Is(err, ErrInvalidPeriodInput), errors.Is(err, ErrInvalidNiveauInput), errors.Is(err, ErrInvalidClassInput), errors.Is(err, ErrPeriodDateRange):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	}
	return internal(err)
}
