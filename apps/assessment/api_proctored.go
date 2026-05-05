package assessment

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	encauth "encore.dev/beta/auth"
	"encore.dev/beta/errs"

	"encore.app/apps/academics"
	"encore.app/apps/auth"
	"encore.app/apps/people"
	"encore.app/pkg/apierr"
)

type TeacherClassDTO struct {
	ID            int64  `json:"id"`
	PeriodID      int64  `json:"period_id"`
	NiveauID      int64  `json:"niveau_id"`
	InstitutionID int64  `json:"institution_id"`
	Code          string `json:"code"`
	Label         string `json:"label"`
}

type ListTeacherClassesResponse struct {
	Classes []TeacherClassDTO `json:"classes"`
	Role    string            `json:"role"`
}

//encore:api auth method=GET path=/v1/teacher/classes
func (s *Service) ListTeacherClassesAPI(ctx context.Context) (*ListTeacherClassesResponse, error) {
	userID, role, err := teacherContext(ctx)
	if err != nil {
		return nil, err
	}

	if role == "admin-global" {
		return &ListTeacherClassesResponse{Classes: []TeacherClassDTO{}, Role: role}, nil
	}

	staffIDs, err := staffIDsForUser(ctx, userID)
	if err != nil {
		return nil, internal(err)
	}

	classes, err := academics.ListClassesForAnyStaff(ctx, staffIDs)
	if err != nil {
		return nil, internal(err)
	}

	out := make([]TeacherClassDTO, 0, len(classes))
	for _, c := range classes {
		out = append(out, TeacherClassDTO{
			ID: c.ID, PeriodID: c.PeriodID, NiveauID: c.NiveauID,
			InstitutionID: c.InstitutionID, Code: c.Code, Label: c.Label,
		})
	}
	return &ListTeacherClassesResponse{Classes: out, Role: role}, nil
}

type TeacherCampaignDTO struct {
	CampaignDTO
	Assigned int `json:"assigned"`
	Scored   int `json:"scored"`
}

type ListTeacherCampaignsResponse struct {
	Campaigns []TeacherCampaignDTO `json:"campaigns"`
}

//encore:api auth method=GET path=/v1/teacher/classes/:classID/campaigns
func (s *Service) ListTeacherClassCampaignsAPI(ctx context.Context, classID int64) (*ListTeacherCampaignsResponse, error) {
	userID, role, err := teacherContext(ctx)
	if err != nil {
		return nil, err
	}

	cls, err := academics.GetClassByID(ctx, classID)
	if err != nil {
		return nil, mapErr(err)
	}

	if role != "admin-global" {
		if err := assertClassAccess(ctx, userID, classID); err != nil {
			return nil, err
		}
	}

	camps, err := ListCampaignsByScope(ctx, cls.InstitutionID)
	if err != nil {
		return nil, internal(err)
	}

	out := make([]TeacherCampaignDTO, 0, len(camps))
	studentIDs, _ := academics.ListClassStudents(ctx, classID)
	for _, c := range camps {
		row := TeacherCampaignDTO{CampaignDTO: campaignToDTO(c)}
		if len(studentIDs) > 0 {
			roster, _ := ListGradingRoster(ctx, c.ID, studentIDs)
			row.Assigned = len(roster)
			for _, r := range roster {
				if r.HasScore {
					row.Scored++
				}
			}
		}
		out = append(out, row)
	}
	return &ListTeacherCampaignsResponse{Campaigns: out}, nil
}

type GradingRosterRowDTO struct {
	StudentID    int64  `json:"student_id"`
	FullName     string `json:"full_name"`
	StudentCode  string `json:"student_code"`
	HasScore     bool   `json:"has_score"`
	RawScore     *int32 `json:"raw_score,omitempty"`
	BandCode     string `json:"band_code,omitempty"`
	BandOrdinal  *int32 `json:"band_ordinal,omitempty"`
	EntryMode    string `json:"entry_mode,omitempty"`
}

type GradingRosterResponse struct {
	Campaign CampaignDTO            `json:"campaign"`
	ClassID  int64                  `json:"class_id"`
	Rows     []GradingRosterRowDTO  `json:"rows"`
}

//encore:api auth method=GET path=/v1/teacher/classes/:classID/campaigns/:campaignID/roster
func (s *Service) GradingRosterAPI(ctx context.Context, classID int64, campaignID int64) (*GradingRosterResponse, error) {
	userID, role, err := teacherContext(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := academics.GetClassByID(ctx, classID); err != nil {
		return nil, mapErr(err)
	}
	if role != "admin-global" {
		if err := assertClassAccess(ctx, userID, classID); err != nil {
			return nil, err
		}
	}

	camp, err := GetCampaignByID(ctx, campaignID)
	if err != nil {
		return nil, mapErr(err)
	}

	studentIDs, err := academics.ListClassStudents(ctx, classID)
	if err != nil {
		return nil, internal(err)
	}

	rosterRows, err := ListGradingRoster(ctx, campaignID, studentIDs)
	if err != nil {
		return nil, internal(err)
	}

	rosterByStudent := make(map[int64]*GradingRosterRow, len(rosterRows))
	for _, r := range rosterRows {
		rosterByStudent[r.StudentID] = r
	}

	out := make([]GradingRosterRowDTO, 0, len(studentIDs))
	for _, sid := range studentIDs {
		row := GradingRosterRowDTO{StudentID: sid}
		if r, ok := rosterByStudent[sid]; ok && r != nil {
			row.HasScore = r.HasScore
			row.RawScore = r.RawScore
			row.BandCode = r.BandCode
			row.BandOrdinal = r.BandOrdinal
			row.EntryMode = r.EntryMode
		}
		if st, err := people.GetStudentByID(ctx, sid); err == nil && st != nil {
			row.StudentCode = st.StudentCode
			if person, err := people.GetPersonByID(ctx, st.PersonID); err == nil && person != nil {
				row.FullName = person.FullName
			}
		}
		out = append(out, row)
	}

	return &GradingRosterResponse{
		Campaign: campaignToDTO(camp),
		ClassID:  classID,
		Rows:     out,
	}, nil
}

type ProctoredEntryDTO struct {
	StudentID int64           `json:"student_id"`
	RawScore  *int32          `json:"raw_score,omitempty"`
	Mode      string          `json:"mode"`
	Answers   json.RawMessage `json:"answers,omitempty"`
}

type SubmitProctoredRequest struct {
	Entries []ProctoredEntryDTO `json:"entries"`
}

type SubmitProctoredResponse struct {
	Created int                   `json:"created"`
	Updated int                   `json:"updated"`
	Errors  []ProctoredEntryError `json:"errors"`
}

//encore:api auth method=POST path=/v1/teacher/classes/:classID/campaigns/:campaignID/scores
func (s *Service) SubmitProctoredAPI(ctx context.Context, classID int64, campaignID int64, req *SubmitProctoredRequest) (*SubmitProctoredResponse, error) {
	userID, role, err := teacherContext(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := academics.GetClassByID(ctx, classID); err != nil {
		return nil, mapErr(err)
	}
	if role != "admin-global" {
		if err := assertClassAccess(ctx, userID, classID); err != nil {
			return nil, err
		}
	}

	classStudents, err := academics.ListClassStudents(ctx, classID)
	if err != nil {
		return nil, internal(err)
	}
	allowed := make(map[int64]bool, len(classStudents))
	for _, s := range classStudents {
		allowed[s] = true
	}

	entries := make([]ProctoredEntry, 0, len(req.Entries))
	rejectedNotInClass := []ProctoredEntryError{}
	for _, e := range req.Entries {
		if !allowed[e.StudentID] {
			rejectedNotInClass = append(rejectedNotInClass, ProctoredEntryError{
				StudentID: e.StudentID,
				Message:   "student is not in this class",
			})
			continue
		}
		entries = append(entries, ProctoredEntry{
			StudentID: e.StudentID,
			RawScore:  e.RawScore,
			Mode:      e.Mode,
			Answers:   e.Answers,
		})
	}

	if len(entries) == 0 {
		return &SubmitProctoredResponse{Errors: rejectedNotInClass}, nil
	}

	res, err := SubmitProctoredScores(ctx, campaignID, userID, entries)
	if err != nil {
		return nil, mapProctoredErr(err)
	}

	res.Errors = append(res.Errors, rejectedNotInClass...)

	return &SubmitProctoredResponse{
		Created: res.Created,
		Updated: res.Updated,
		Errors:  res.Errors,
	}, nil
}

func teacherContext(ctx context.Context) (userID int64, role string, err error) {
	uid, ok := encauth.UserID()
	if !ok || uid == "" {
		return 0, "", &errs.Error{Code: errs.Unauthenticated, Message: "missing user id"}
	}
	id, perr := strconv.ParseInt(string(uid), 10, 64)
	if perr != nil {
		return 0, "", &errs.Error{Code: errs.Unauthenticated, Message: "invalid user id"}
	}

	assignments, err := auth.ListRoleAssignmentsForUser(ctx, id)
	if err != nil {
		return 0, "", &errs.Error{Code: errs.Internal, Message: "could not check role"}
	}

	hasAdminGlobal := false
	hasTeacher := false
	for _, a := range assignments {
		if a.Role == "admin" && a.ScopeNodeID == nil {
			hasAdminGlobal = true
		}
		if a.Role == "teacher" {
			hasTeacher = true
		}
	}

	if hasAdminGlobal {
		return id, "admin-global", nil
	}
	if hasTeacher {
		return id, "teacher", nil
	}
	return 0, "", &errs.Error{Code: errs.PermissionDenied, Message: "teacher or admin role required"}
}

func staffIDsForUser(ctx context.Context, userID int64) ([]int64, error) {
	user, err := auth.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	person, err := people.GetPersonByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, people.ErrPersonNotFound) {
			return nil, nil
		}
		return nil, err
	}
	staffRows, err := people.ListStaffByPersonID(ctx, person.ID)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(staffRows))
	for _, s := range staffRows {
		out = append(out, s.ID)
	}
	return out, nil
}

func assertClassAccess(ctx context.Context, userID int64, classID int64) error {
	staffIDs, err := staffIDsForUser(ctx, userID)
	if err != nil {
		return internal(err)
	}
	if len(staffIDs) == 0 {
		return &errs.Error{Code: errs.PermissionDenied, Message: "you are not on this class's staff"}
	}
	ok, err := academics.IsClassStaff(ctx, classID, staffIDs)
	if err != nil {
		return internal(err)
	}
	if !ok {
		return &errs.Error{Code: errs.PermissionDenied, Message: "you are not on this class's staff"}
	}
	return nil
}

func mapProctoredErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrCampaignNotFound):
		return &errs.Error{Code: errs.NotFound, Message: err.Error()}
	case errors.Is(err, ErrInvalidProctoredInput),
		errors.Is(err, ErrCampaignNotOpen),
		errors.Is(err, ErrProctoredEmptyBatch),
		errors.Is(err, ErrScoreOutOfRange):
		return &errs.Error{Code: errs.InvalidArgument, Message: err.Error()}
	}
	return apierr.WrapInternal("assessment.proctored", err)
}
