package enrollment

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"encore.app/apps/enrollment/dbenrollment"
	"encore.app/apps/people"
	"encore.app/apps/tenancy"
)

var ErrInvalidCoverageInput = errors.New("enrollment: invalid coverage input")

type Coverage struct {
	ScopeNodeID    int64
	PeriodID       int64
	TotalEnrolled  int
	Male           int
	Female         int
	Other          int
	Unknown        int
	InstitutionIDs []int64
}

func ComputeCoverage(ctx context.Context, scopeNodeID, periodID int64) (*Coverage, error) {
	if scopeNodeID <= 0 || periodID <= 0 {
		return nil, ErrInvalidCoverageInput
	}

	institutionIDs, err := tenancy.ListDescendantIDs(ctx, scopeNodeID)
	if err != nil {
		return nil, err
	}
	if len(institutionIDs) == 0 {
		institutionIDs = []int64{scopeNodeID}
	}

	rows, err := queries.ListActiveStudentIDsForInstitutions(ctx, dbenrollment.ListActiveStudentIDsForInstitutionsParams{
		InstitutionIds: institutionIDs,
		PeriodID:       periodID,
	})
	if err != nil {
		return nil, err
	}

	studentIDs := make([]int64, 0, len(rows))
	for _, r := range rows {
		studentIDs = append(studentIDs, r.StudentID)
	}

	genders, err := people.GetStudentGendersByIDs(ctx, studentIDs)
	if err != nil {
		return nil, err
	}

	cov := &Coverage{
		ScopeNodeID:    scopeNodeID,
		PeriodID:       periodID,
		TotalEnrolled:  len(studentIDs),
		InstitutionIDs: institutionIDs,
	}
	for _, sid := range studentIDs {
		switch strings.ToLower(genders[sid]) {
		case "male", "m", "garcon", "garçon":
			cov.Male++
		case "female", "f", "fille":
			cov.Female++
		case "":
			cov.Unknown++
		default:
			cov.Other++
		}
	}
	return cov, nil
}

func SnapshotCoverage(ctx context.Context, scopeNodeID, periodID int64, campaignID *int64) (*Coverage, error) {
	cov, err := ComputeCoverage(ctx, scopeNodeID, periodID)
	if err != nil {
		return nil, err
	}

	var camp sql.NullInt64
	if campaignID != nil && *campaignID > 0 {
		camp = sql.NullInt64{Int64: *campaignID, Valid: true}
	}

	if _, err := queries.UpsertCoverageSnapshot(ctx, dbenrollment.UpsertCoverageSnapshotParams{
		ScopeNodeID:   scopeNodeID,
		PeriodID:      periodID,
		CampaignID:    camp,
		TotalEnrolled: int32(cov.TotalEnrolled),
		TotalMale:     int32(cov.Male),
		TotalFemale:   int32(cov.Female),
		TotalOther:    int32(cov.Other),
		TotalTested:   0,
	}); err != nil {
		return nil, err
	}
	return cov, nil
}
