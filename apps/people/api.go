package people

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/errs"
)

type PersonDTO struct {
	ID         int64  `json:"id"`
	FullName   string `json:"fullName"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Gender     string `json:"gender"`
}

type StudentDTO struct {
	ID             int64      `json:"id"`
	PersonID       int64      `json:"personId"`
	InstitutionID  int64      `json:"institutionId"`
	StudentCode    string     `json:"studentCode"`
	EnrollmentDate *time.Time `json:"enrollmentDate"`
	Person         PersonDTO  `json:"person"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type StaffDTO struct {
	ID          int64      `json:"id"`
	PersonID    int64      `json:"personId"`
	ScopeNodeID int64      `json:"scopeNodeId"`
	Position    string     `json:"position"`
	StaffCode   string     `json:"staffCode"`
	HireDate    *time.Time `json:"hireDate"`
	Person      PersonDTO  `json:"person"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func personToDTO(p *Person) PersonDTO {
	if p == nil {
		return PersonDTO{}
	}

	return PersonDTO{
		ID:         p.ID,
		FullName:   p.FullName,
		GivenName:  p.GivenName,
		FamilyName: p.FamilyName,
		Email:      p.Email,
		Phone:      p.Phone,
		Gender:     p.Gender,
	}
}

func studentToDTO(s *Student, person *Person) StudentDTO {
	return StudentDTO{
		ID:             s.ID,
		PersonID:       s.PersonID,
		InstitutionID:  s.InstitutionID,
		StudentCode:    s.StudentCode,
		EnrollmentDate: s.EnrollmentDate,
		Person:         personToDTO(person),
		CreatedAt:      s.CreatedAt,
	}
}

func staffToDTO(s *Staff, person *Person) StaffDTO {
	return StaffDTO{
		ID:          s.ID,
		PersonID:    s.PersonID,
		ScopeNodeID: s.ScopeNodeID,
		Position:    s.Position,
		StaffCode:   s.StaffCode,
		HireDate:    s.HireDate,
		Person:      personToDTO(person),
		CreatedAt:   s.CreatedAt,
	}
}

type ListStudentsRequest struct {
	InstitutionID int64 `query:"institutionId"`
	Limit         int32 `query:"limit"`
	Offset        int32 `query:"offset"`
}

type ListStudentsResponse struct {
	Students []StudentDTO `json:"students"`
}

//encore:api auth method=GET path=/v1/people/students
func (s *Service) ListStudentsAPI(ctx context.Context, req *ListStudentsRequest) (*ListStudentsResponse, error) {
	if req.InstitutionID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "institutionId required"}
	}

	if err := requireNodeAccess(ctx, req.InstitutionID); err != nil {
		return nil, err
	}

	students, err := ListStudentsByInstitutionWithPerson(ctx, req.InstitutionID, req.Limit, req.Offset)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not list students"}
	}

	out := make([]StudentDTO, 0, len(students))
	for _, sp := range students {
		out = append(out, studentToDTO(sp.Student, sp.Person))
	}

	return &ListStudentsResponse{Students: out}, nil
}

type CreateStudentAPIRequest struct {
	InstitutionID  int64      `json:"institutionId"`
	StudentCode    string     `json:"studentCode"`
	EnrollmentDate *time.Time `json:"enrollmentDate"`
	FullName       string     `json:"fullName"`
	GivenName      string     `json:"givenName"`
	FamilyName     string     `json:"familyName"`
	DateOfBirth    *time.Time `json:"dateOfBirth"`
	Gender         string     `json:"gender"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
}

type CreateStudentResponse struct {
	Student StudentDTO `json:"student"`
}

//encore:api auth method=POST path=/v1/people/students
func (s *Service) CreateStudentAPI(ctx context.Context, req *CreateStudentAPIRequest) (*CreateStudentResponse, error) {
	if err := requireNodeAccess(ctx, req.InstitutionID); err != nil {
		return nil, err
	}

	student, person, err := CreateStudent(ctx, CreateStudentParams{
		Person: CreatePersonParams{
			FullName:    req.FullName,
			GivenName:   req.GivenName,
			FamilyName:  req.FamilyName,
			DateOfBirth: req.DateOfBirth,
			Gender:      req.Gender,
			Email:       req.Email,
			Phone:       req.Phone,
		},
		InstitutionID:  req.InstitutionID,
		StudentCode:    req.StudentCode,
		EnrollmentDate: req.EnrollmentDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStudentInput), errors.Is(err, ErrInvalidPersonInput):
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid student input"}
		case errors.Is(err, ErrStudentCodeAlreadyUsed):
			return nil, &errs.Error{Code: errs.AlreadyExists, Message: "student code already in use under this institution"}
		default:
			return nil, &errs.Error{Code: errs.Internal, Message: "could not create student"}
		}
	}

	return &CreateStudentResponse{Student: studentToDTO(student, person)}, nil
}

type GetStudentResponse struct {
	Student StudentDTO `json:"student"`
}

//encore:api auth method=GET path=/v1/people/students/:id
func (s *Service) GetStudentAPI(ctx context.Context, id int64) (*GetStudentResponse, error) {
	student, err := GetStudentByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrStudentNotFound) {
			return nil, &errs.Error{Code: errs.NotFound, Message: "student not found"}
		}

		return nil, &errs.Error{Code: errs.Internal, Message: "could not load student"}
	}

	if err := requireNodeAccess(ctx, student.InstitutionID); err != nil {
		return nil, err
	}

	person, _ := GetPersonByID(ctx, student.PersonID)

	return &GetStudentResponse{Student: studentToDTO(student, person)}, nil
}

//encore:api auth method=DELETE path=/v1/people/students/:id
func (s *Service) DeleteStudentAPI(ctx context.Context, id int64) error {
	student, err := GetStudentByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrStudentNotFound) {
			return &errs.Error{Code: errs.NotFound, Message: "student not found"}
		}

		return &errs.Error{Code: errs.Internal, Message: "could not load student"}
	}

	if err := requireNodeAccess(ctx, student.InstitutionID); err != nil {
		return err
	}

	if err := SoftDeleteStudent(ctx, id); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not delete student"}
	}

	return nil
}

type ListStaffRequest struct {
	ScopeNodeID int64 `query:"scopeNodeId"`
	Limit       int32 `query:"limit"`
	Offset      int32 `query:"offset"`
}

type ListStaffResponse struct {
	Staff []StaffDTO `json:"staff"`
}

//encore:api auth method=GET path=/v1/people/staff
func (s *Service) ListStaffAPI(ctx context.Context, req *ListStaffRequest) (*ListStaffResponse, error) {
	if req.ScopeNodeID <= 0 {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "scopeNodeId required"}
	}

	if err := requireNodeAccess(ctx, req.ScopeNodeID); err != nil {
		return nil, err
	}

	rows, err := ListStaffByScope(ctx, req.ScopeNodeID, req.Limit, req.Offset)
	if err != nil {
		return nil, &errs.Error{Code: errs.Internal, Message: "could not list staff"}
	}

	out := make([]StaffDTO, 0, len(rows))
	for _, st := range rows {
		person, _ := GetPersonByID(ctx, st.PersonID)
		out = append(out, staffToDTO(st, person))
	}

	return &ListStaffResponse{Staff: out}, nil
}

type CreateStaffAPIRequest struct {
	ScopeNodeID int64      `json:"scopeNodeId"`
	Position    string     `json:"position"`
	StaffCode   string     `json:"staffCode"`
	HireDate    *time.Time `json:"hireDate"`
	FullName    string     `json:"fullName"`
	GivenName   string     `json:"givenName"`
	FamilyName  string     `json:"familyName"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
}

type CreateStaffResponse struct {
	Staff StaffDTO `json:"staff"`
}

//encore:api auth method=POST path=/v1/people/staff
func (s *Service) CreateStaffAPI(ctx context.Context, req *CreateStaffAPIRequest) (*CreateStaffResponse, error) {
	if err := requireNodeAccess(ctx, req.ScopeNodeID); err != nil {
		return nil, err
	}

	staff, person, err := CreateStaff(ctx, CreateStaffParams{
		Person: CreatePersonParams{
			FullName:   req.FullName,
			GivenName:  req.GivenName,
			FamilyName: req.FamilyName,
			Email:      req.Email,
			Phone:      req.Phone,
		},
		ScopeNodeID: req.ScopeNodeID,
		Position:    req.Position,
		StaffCode:   req.StaffCode,
		HireDate:    req.HireDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStaffInput), errors.Is(err, ErrInvalidPersonInput):
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "invalid staff input"}
		case errors.Is(err, ErrStaffCodeAlreadyUsed):
			return nil, &errs.Error{Code: errs.AlreadyExists, Message: "staff code already used in this scope"}
		default:
			return nil, &errs.Error{Code: errs.Internal, Message: "could not create staff"}
		}
	}

	return &CreateStaffResponse{Staff: staffToDTO(staff, person)}, nil
}

//encore:api auth method=DELETE path=/v1/people/staff/:id
func (s *Service) DeleteStaffAPI(ctx context.Context, id int64) error {
	staff, err := GetStaffByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrStaffNotFound) {
			return &errs.Error{Code: errs.NotFound, Message: "staff not found"}
		}

		return &errs.Error{Code: errs.Internal, Message: "could not load staff"}
	}

	if err := requireNodeAccess(ctx, staff.ScopeNodeID); err != nil {
		return err
	}

	if err := SoftDeleteStaff(ctx, id); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not delete staff"}
	}

	return nil
}
