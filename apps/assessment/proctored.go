package assessment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"encore.app/apps/assessment/dbassessment"
)

const (
	EntryModeStudent          = "student"
	EntryModeProctoredScore   = "proctored_score"
	EntryModeProctoredAnswers = "proctored_answers"
)

var (
	ErrInvalidProctoredInput = errors.New("assessment: invalid proctored input")
	ErrProctoredEmptyBatch   = errors.New("assessment: proctored batch is empty")
)

type ProctoredEntry struct {
	StudentID int64
	RawScore  *int32
	Mode      string
	Answers   json.RawMessage
}

type ProctoredEntryError struct {
	StudentID int64  `json:"student_id"`
	Message   string `json:"message"`
}

type ProctoredBatchResult struct {
	Created int                   `json:"created"`
	Updated int                   `json:"updated"`
	Errors  []ProctoredEntryError `json:"errors"`
}

func SubmitProctoredScores(ctx context.Context, campaignID int64, proctorUserID int64, entries []ProctoredEntry) (*ProctoredBatchResult, error) {
	if campaignID <= 0 {
		return nil, ErrInvalidProctoredInput
	}
	if len(entries) == 0 {
		return nil, ErrProctoredEmptyBatch
	}

	camp, err := GetCampaignByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if camp.Status != StatusOpen {
		return nil, ErrCampaignNotOpen
	}

	res := &ProctoredBatchResult{Errors: []ProctoredEntryError{}}

	for _, entry := range entries {
		updated, err := writeProctoredEntry(ctx, camp, proctorUserID, entry)
		if err != nil {
			res.Errors = append(res.Errors, ProctoredEntryError{
				StudentID: entry.StudentID,
				Message:   err.Error(),
			})
			continue
		}
		if updated {
			res.Updated++
		} else {
			res.Created++
		}
	}

	return res, nil
}

func writeProctoredEntry(ctx context.Context, camp *Campaign, proctorUserID int64, entry ProctoredEntry) (updated bool, err error) {
	if entry.StudentID <= 0 {
		return false, ErrInvalidProctoredInput
	}
	if entry.RawScore == nil {
		return false, ErrInvalidProctoredInput
	}
	if *entry.RawScore < 0 || *entry.RawScore > 100 {
		return false, ErrScoreOutOfRange
	}
	mode := entry.Mode
	if mode == "" {
		mode = EntryModeProctoredScore
	}
	if mode != EntryModeProctoredScore && mode != EntryModeProctoredAnswers {
		return false, ErrInvalidProctoredInput
	}

	existing, err := queries.GetAssignmentByCampaignStudent(ctx, dbassessment.GetAssignmentByCampaignStudentParams{
		CampaignID: camp.ID,
		StudentID:  entry.StudentID,
	})
	hadAssignment := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	var assignmentID int64
	if hadAssignment {
		assignmentID = existing.ID
		if err := queries.DeleteResponseByAssignment(ctx, assignmentID); err != nil {
			return false, err
		}
	} else {
		token, terr := newToken(24)
		if terr != nil {
			return false, terr
		}
		row, cerr := queries.CreateAssignment(ctx, dbassessment.CreateAssignmentParams{
			CampaignID:  camp.ID,
			StudentID:   entry.StudentID,
			AccessToken: token,
		})
		if cerr != nil {
			if isUniqueViolation(cerr) {
				return false, ErrAssignmentExists
			}
			return false, cerr
		}
		assignmentID = row.ID
	}

	payload := entry.Answers
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}

	respRow, err := queries.CreateProctoredResponse(ctx, dbassessment.CreateProctoredResponseParams{
		AssignmentID:       assignmentID,
		CampaignID:         camp.ID,
		StudentID:          entry.StudentID,
		Payload:            payload,
		ProctoredByUserID:  sql.NullInt64{Int64: proctorUserID, Valid: proctorUserID > 0},
		EntryMode:          mode,
	})
	if err != nil {
		return false, err
	}

	bandRow, err := queries.BandForScore(ctx, dbassessment.BandForScoreParams{
		ScaleCode: camp.ScaleCode, Column2: *entry.RawScore,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrScoreOutOfRange
		}
		return false, err
	}

	if _, err := queries.CreateScore(ctx, dbassessment.CreateScoreParams{
		ResponseID:  respRow.ID,
		CampaignID:  camp.ID,
		StudentID:   entry.StudentID,
		RawScore:    *entry.RawScore,
		BandCode:    bandRow.Code,
		BandOrdinal: bandRow.Ordinal,
	}); err != nil {
		return false, err
	}

	if _, err := queries.MarkAssignmentSubmitted(ctx, assignmentID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return false, err
		}
	}

	return hadAssignment, nil
}

type GradingRosterRow struct {
	StudentID         int64
	AssignmentID      *int64
	HasScore          bool
	RawScore          *int32
	BandCode          string
	BandOrdinal       *int32
	EntryMode         string
	SubmittedAt       *time.Time
	ProctoredByUserID *int64
}

func ListGradingRoster(ctx context.Context, campaignID int64, studentIDs []int64) ([]*GradingRosterRow, error) {
	if campaignID <= 0 {
		return nil, ErrInvalidProctoredInput
	}
	if len(studentIDs) == 0 {
		return []*GradingRosterRow{}, nil
	}

	rows, err := queries.ListGradingRosterRows(ctx, dbassessment.ListGradingRosterRowsParams{
		CampaignID: campaignID,
		StudentIds: studentIDs,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*GradingRosterRow, 0, len(rows))
	for _, r := range rows {
		row := &GradingRosterRow{StudentID: r.StudentID}
		if r.AssignmentID.Valid {
			v := r.AssignmentID.Int64
			row.AssignmentID = &v
		}
		if r.ScoreID.Valid {
			row.HasScore = true
		}
		if r.RawScore.Valid {
			v := r.RawScore.Int32
			row.RawScore = &v
		}
		if r.BandCode.Valid {
			row.BandCode = r.BandCode.String
		}
		if r.BandOrdinal.Valid {
			v := r.BandOrdinal.Int32
			row.BandOrdinal = &v
		}
		if r.EntryMode.Valid {
			row.EntryMode = r.EntryMode.String
		}
		if r.SubmittedAt.Valid {
			t := r.SubmittedAt.Time
			row.SubmittedAt = &t
		}
		if r.ProctoredByUserID.Valid {
			v := r.ProctoredByUserID.Int64
			row.ProctoredByUserID = &v
		}
		out = append(out, row)
	}
	return out, nil
}
