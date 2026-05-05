package enrollment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/enrollment"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

func uniqueIDs() (studentID, periodID int64) {
	now := time.Now().UnixNano()
	return now, now/1000 + 1
}

func TestCreateEnrollment_Success(t *testing.T) {
	ctx := context.Background()
	studentID, periodID := uniqueIDs()

	e, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID:     studentID,
		InstitutionID: 100,
		PeriodID:      periodID,
		EnrolledOn:    mustDate(t, "2025-09-01"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if e.Status != enrollment.StatusActive {
		t.Fatalf("status=%q, want active", e.Status)
	}

	events, err := enrollment.ListEvents(ctx, e.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 || events[0].Kind != enrollment.EventCreated {
		t.Fatalf("expected 1 'created' event, got %+v", events)
	}
}

func TestCreateEnrollment_Validation(t *testing.T) {
	ctx := context.Background()
	cases := []enrollment.CreateEnrollmentParams{
		{InstitutionID: 1, PeriodID: 1, EnrolledOn: mustDate(t, "2025-01-01")},
		{StudentID: 1, PeriodID: 1, EnrolledOn: mustDate(t, "2025-01-01")},
		{StudentID: 1, InstitutionID: 1, EnrolledOn: mustDate(t, "2025-01-01")},
		{StudentID: 1, InstitutionID: 1, PeriodID: 1},
	}
	for i, c := range cases {
		if _, err := enrollment.CreateEnrollment(ctx, c); !errors.Is(err, enrollment.ErrInvalidEnrollmentInput) {
			t.Fatalf("case %d: err=%v, want ErrInvalidEnrollmentInput", i, err)
		}
	}
}

func TestCreateEnrollment_DoubleActiveBlocked(t *testing.T) {
	ctx := context.Background()
	studentID, periodID := uniqueIDs()

	if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 100, PeriodID: periodID, EnrolledOn: mustDate(t, "2025-09-01"),
	}); err != nil {
		t.Fatalf("first: %v", err)
	}

	_, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 200, PeriodID: periodID, EnrolledOn: mustDate(t, "2025-09-02"),
	})
	if !errors.Is(err, enrollment.ErrAlreadyActiveEnrollment) {
		t.Fatalf("err=%v, want ErrAlreadyActiveEnrollment", err)
	}
}

func TestDropEnrollment_RecordsEvent(t *testing.T) {
	ctx := context.Background()
	studentID, periodID := uniqueIDs()

	e, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 100, PeriodID: periodID, EnrolledOn: mustDate(t, "2025-09-01"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	dropped, err := enrollment.DropEnrollment(ctx, enrollment.DropParams{
		EnrollmentID: e.ID, EndedOn: mustDate(t, "2026-01-15"), Note: "moved abroad",
	})
	if err != nil {
		t.Fatalf("drop: %v", err)
	}
	if dropped.Status != enrollment.StatusDropped {
		t.Fatalf("status=%q, want dropped", dropped.Status)
	}
	if dropped.EndedOn == nil {
		t.Fatalf("ended_on not set")
	}

	if _, err := enrollment.DropEnrollment(ctx, enrollment.DropParams{EnrollmentID: e.ID, EndedOn: mustDate(t, "2026-02-01")}); !errors.Is(err, enrollment.ErrEnrollmentNotActive) {
		t.Fatalf("re-drop err=%v, want ErrEnrollmentNotActive", err)
	}

	events, _ := enrollment.ListEvents(ctx, e.ID)
	kinds := []string{}
	for _, ev := range events {
		kinds = append(kinds, ev.Kind)
	}
	if len(kinds) != 2 || kinds[0] != enrollment.EventDropped || kinds[1] != enrollment.EventCreated {
		t.Fatalf("event order = %v, want [dropped created]", kinds)
	}
}

func TestDropEnrollment_FreesSlotForNewActive(t *testing.T) {
	ctx := context.Background()
	studentID, periodID := uniqueIDs()

	e, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 100, PeriodID: periodID, EnrolledOn: mustDate(t, "2025-09-01"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := enrollment.DropEnrollment(ctx, enrollment.DropParams{EnrollmentID: e.ID, EndedOn: mustDate(t, "2026-01-15")}); err != nil {
		t.Fatalf("drop: %v", err)
	}

	if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 200, PeriodID: periodID, EnrolledOn: mustDate(t, "2026-01-20"),
	}); err != nil {
		t.Fatalf("re-enroll after drop: %v", err)
	}
}

func TestTransferEnrollment_HappyPath(t *testing.T) {
	ctx := context.Background()
	studentID, periodID := uniqueIDs()

	if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 100, PeriodID: periodID, EnrolledOn: mustDate(t, "2025-09-01"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	res, err := enrollment.TransferEnrollment(ctx, enrollment.TransferParams{
		StudentID: studentID, PeriodID: periodID, ToInstitutionID: 200,
		EffectiveOn: mustDate(t, "2025-11-15"), Note: "family relocation",
	})
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if res.Closed.Status != enrollment.StatusTransferred {
		t.Fatalf("closed status=%q, want transferred", res.Closed.Status)
	}
	if res.Opened.InstitutionID != 200 || res.Opened.Status != enrollment.StatusActive {
		t.Fatalf("opened wrong: %+v", res.Opened)
	}

	current, err := enrollment.GetActiveEnrollment(ctx, studentID, periodID)
	if err != nil {
		t.Fatalf("get active: %v", err)
	}
	if current.ID != res.Opened.ID {
		t.Fatalf("active = %d, want opened=%d", current.ID, res.Opened.ID)
	}

	all, err := enrollment.ListEnrollmentsByStudent(ctx, studentID)
	if err != nil {
		t.Fatalf("list student: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("student has %d enrollments, want 2", len(all))
	}

	closedEvents, _ := enrollment.ListEvents(ctx, res.Closed.ID)
	if len(closedEvents) != 2 || closedEvents[0].Kind != enrollment.EventTransferred {
		t.Fatalf("closed events = %+v, want [transferred created]", closedEvents)
	}
	openedEvents, _ := enrollment.ListEvents(ctx, res.Opened.ID)
	if len(openedEvents) != 1 || openedEvents[0].Kind != enrollment.EventCreated {
		t.Fatalf("opened events = %+v, want [created]", openedEvents)
	}
	if openedEvents[0].FromInstitutionID == nil || *openedEvents[0].FromInstitutionID != 100 {
		t.Fatalf("opened.created event missing from_institution_id=100: %+v", openedEvents[0])
	}
}

func TestTransferEnrollment_SameInstitutionBlocked(t *testing.T) {
	ctx := context.Background()
	studentID, periodID := uniqueIDs()

	if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: studentID, InstitutionID: 100, PeriodID: periodID, EnrolledOn: mustDate(t, "2025-09-01"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err := enrollment.TransferEnrollment(ctx, enrollment.TransferParams{
		StudentID: studentID, PeriodID: periodID, ToInstitutionID: 100,
	})
	if !errors.Is(err, enrollment.ErrSameInstitutionTransfer) {
		t.Fatalf("err=%v, want ErrSameInstitutionTransfer", err)
	}
}

func TestTransferEnrollment_NoActiveEnrollment(t *testing.T) {
	ctx := context.Background()
	_, err := enrollment.TransferEnrollment(ctx, enrollment.TransferParams{
		StudentID: 999999991, PeriodID: 999999991, ToInstitutionID: 1,
	})
	if !errors.Is(err, enrollment.ErrEnrollmentNotFound) {
		t.Fatalf("err=%v, want ErrEnrollmentNotFound", err)
	}
}

func TestListEnrollmentsByInstitution_FiltersByStatus(t *testing.T) {
	ctx := context.Background()
	periodID := time.Now().UnixNano() + 7777
	institutionID := int64(7777)

	for i := 1; i <= 3; i++ {
		studentID := time.Now().UnixNano() + int64(i)
		e, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
			StudentID: studentID, InstitutionID: institutionID, PeriodID: periodID,
			EnrolledOn: mustDate(t, "2025-09-01"),
		})
		if err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
		if i == 3 {
			if _, err := enrollment.DropEnrollment(ctx, enrollment.DropParams{EnrollmentID: e.ID, EndedOn: mustDate(t, "2025-10-01")}); err != nil {
				t.Fatalf("drop: %v", err)
			}
		}
	}

	active, err := enrollment.ListEnrollmentsByInstitution(ctx, institutionID, periodID, false)
	if err != nil {
		t.Fatalf("active: %v", err)
	}
	if len(active) != 2 {
		t.Fatalf("active count = %d, want 2", len(active))
	}

	all, err := enrollment.ListEnrollmentsByInstitution(ctx, institutionID, periodID, true)
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("all count = %d, want 3", len(all))
	}
}
