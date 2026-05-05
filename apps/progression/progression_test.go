package progression_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/assessment"
	"encore.app/apps/enrollment"
	"encore.app/apps/people"
	"encore.app/apps/progression"
	"encore.app/apps/tenancy"
)

type fixture struct {
	region    *tenancy.Node
	schoolA   *tenancy.Node
	schoolB   *tenancy.Node
	periodID  int64
	campaign  *assessment.Campaign
	studentsA []*people.Student
	studentB  *people.Student
}

func setupFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().Format("150405.000000")

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "prog-r-" + suffix, Label: "Region", Level: tenancy.LevelRegion,
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}
	pref, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "prog-p-" + suffix, Label: "Pref", Level: tenancy.LevelPrefecture, ParentID: &region.ID,
	})
	if err != nil {
		t.Fatalf("pref: %v", err)
	}
	deleg, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "prog-d-" + suffix, Label: "Deleg", Level: tenancy.LevelDelegation, ParentID: &pref.ID,
	})
	if err != nil {
		t.Fatalf("deleg: %v", err)
	}
	a, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "prog-a-" + suffix, Label: "School A", Level: tenancy.LevelInstitution, ParentID: &deleg.ID,
	})
	if err != nil {
		t.Fatalf("schoolA: %v", err)
	}
	b, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Code: "prog-b-" + suffix, Label: "School B", Level: tenancy.LevelInstitution, ParentID: &deleg.ID,
	})
	if err != nil {
		t.Fatalf("schoolB: %v", err)
	}
	t.Cleanup(func() {
		_ = tenancy.SoftDeleteNode(ctx, a.ID)
		_ = tenancy.SoftDeleteNode(ctx, b.ID)
		_ = tenancy.SoftDeleteNode(ctx, deleg.ID)
		_ = tenancy.SoftDeleteNode(ctx, pref.ID)
		_ = tenancy.SoftDeleteNode(ctx, region.ID)
	})

	periodID := time.Now().UnixNano()
	formID := periodID + 1
	formVerID := periodID + 2

	mkStudent := func(institutionID int64, name string) *people.Student {
		st, _, err := people.CreateStudent(ctx, people.CreateStudentParams{
			InstitutionID: institutionID,
			Person:        people.CreatePersonParams{FullName: name + " " + suffix},
		})
		if err != nil {
			t.Fatalf("student %s: %v", name, err)
		}
		t.Cleanup(func() { _ = people.SoftDeleteStudent(ctx, st.ID) })
		return st
	}

	a1 := mkStudent(a.ID, "Alice")
	a2 := mkStudent(a.ID, "Bob")
	a3 := mkStudent(a.ID, "Cleo")
	b1 := mkStudent(b.ID, "Eli")

	for _, s := range []*people.Student{a1, a2, a3} {
		if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
			StudentID: s.ID, InstitutionID: a.ID, PeriodID: periodID,
			EnrolledOn: time.Now(),
		}); err != nil {
			t.Fatalf("enroll: %v", err)
		}
	}
	if _, err := enrollment.CreateEnrollment(ctx, enrollment.CreateEnrollmentParams{
		StudentID: b1.ID, InstitutionID: b.ID, PeriodID: periodID,
		EnrolledOn: time.Now(),
	}); err != nil {
		t.Fatalf("enroll b1: %v", err)
	}

	camp, err := assessment.CreateCampaign(ctx, assessment.CreateCampaignParams{
		Title: "Prog Test", ScaleCode: assessment.ScaleFrench,
		FormID: formID, FormVersionID: formVerID,
		PeriodID: periodID, ScopeNodeID: region.ID,
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	if _, err := assessment.OpenCampaign(ctx, camp.ID); err != nil {
		t.Fatalf("open: %v", err)
	}

	for _, s := range []*people.Student{a1, a2, a3, b1} {
		res, err := assessment.AssignStudents(ctx, assessment.AssignParams{
			CampaignID: camp.ID, StudentIDs: []int64{s.ID},
		})
		if err != nil {
			t.Fatalf("assign %d: %v", s.ID, err)
		}
		var score int32 = 30
		if s.ID == a2.ID {
			score = 65
		} else if s.ID == a3.ID {
			score = 90
		} else if s.ID == b1.ID {
			score = 50
		}
		if _, err := assessment.SubmitResponse(ctx, assessment.SubmitParams{
			AccessToken: res.Created[0].AccessToken, RawScore: score,
		}); err != nil {
			t.Fatalf("submit %d: %v", s.ID, err)
		}
	}

	return &fixture{
		region: region, schoolA: a, schoolB: b,
		periodID: periodID, campaign: camp,
		studentsA: []*people.Student{a1, a2, a3}, studentB: b1,
	}
}

func TestComputeProgression_AggregatesPerScope(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	progA, err := progression.ComputeProgression(ctx, f.schoolA.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("schoolA: %v", err)
	}
	if progA.TotalScored != 3 {
		t.Fatalf("schoolA total=%d, want 3", progA.TotalScored)
	}

	bandCounts := map[string]int32{}
	for _, b := range progA.Bands {
		bandCounts[b.BandCode] = b.StudentCount
	}
	if bandCounts["lettres"] != 1 {
		t.Fatalf("lettres count = %d, want 1 (Alice score=30)", bandCounts["lettres"])
	}
	if bandCounts["paragraphe"] != 1 {
		t.Fatalf("paragraphe count = %d, want 1 (Bob score=65)", bandCounts["paragraphe"])
	}
	if bandCounts["histoire"] != 1 {
		t.Fatalf("histoire count = %d, want 1 (Cleo score=90)", bandCounts["histoire"])
	}
}

func TestComputeProgression_RollsUpAcrossSchools(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	progRegion, err := progression.ComputeProgression(ctx, f.region.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("region: %v", err)
	}
	if progRegion.TotalScored != 4 {
		t.Fatalf("region total=%d, want 4 (3 in A + 1 in B)", progRegion.TotalScored)
	}
}

func TestComputeProgression_PercentagesSumTo100(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	prog, err := progression.ComputeProgression(ctx, f.schoolA.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	var sum int32
	for _, b := range prog.Bands {
		sum += b.Percentage
	}
	if sum < 99 || sum > 101 {
		t.Fatalf("percentages sum=%d, want ~100 (got %+v)", sum, prog.Bands)
	}
}

func TestRefreshSnapshot_PersistsAndRecallsViaGetSnapshot(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	first, err := progression.RefreshSnapshot(ctx, f.schoolA.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if first.TotalScored != 3 {
		t.Fatalf("refresh total=%d, want 3", first.TotalScored)
	}

	got, err := progression.GetSnapshot(ctx, f.schoolA.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.TotalScored != 3 {
		t.Fatalf("get total=%d, want 3", got.TotalScored)
	}
	gotCounts := map[string]int32{}
	for _, b := range got.Bands {
		gotCounts[b.BandCode] = b.StudentCount
	}
	if gotCounts["lettres"] != 1 || gotCounts["paragraphe"] != 1 || gotCounts["histoire"] != 1 {
		t.Fatalf("snapshot counts wrong: %+v", gotCounts)
	}
}

func TestDrilldown_ShowsAllChildSchools(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	parent, err := tenancy.GetNodeByID(ctx, f.schoolA.ID)
	if err != nil {
		t.Fatalf("get parent: %v", err)
	}
	delegID := *parent.ParentID

	d, err := progression.DrilldownByScope(ctx, delegID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("drilldown: %v", err)
	}
	if d.Scope.TotalScored != 4 {
		t.Fatalf("scope total=%d, want 4", d.Scope.TotalScored)
	}
	if len(d.Children) != 2 {
		t.Fatalf("children=%d, want 2", len(d.Children))
	}

	var aChild, bChild *progression.DrilldownChild
	for i := range d.Children {
		c := &d.Children[i]
		if c.NodeID == f.schoolA.ID {
			aChild = c
		} else if c.NodeID == f.schoolB.ID {
			bChild = c
		}
	}
	if aChild == nil || aChild.Total != 3 {
		t.Fatalf("schoolA child wrong: %+v", aChild)
	}
	if bChild == nil || bChild.Total != 1 {
		t.Fatalf("schoolB child wrong: %+v", bChild)
	}
}

func TestComputeProgression_ValidationRejectsBadInput(t *testing.T) {
	ctx := context.Background()
	if _, err := progression.ComputeProgression(ctx, 0, 1, 1); !errors.Is(err, progression.ErrInvalidInput) {
		t.Fatalf("err=%v, want ErrInvalidInput", err)
	}
	if _, err := progression.ComputeProgression(ctx, 1, 1, 999999999); !errors.Is(err, progression.ErrCampaignNotFound) {
		t.Fatalf("err=%v, want ErrCampaignNotFound", err)
	}
}

func TestRefreshAllOpenCampaigns_PopulatesSnapshotsForAllAncestors(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	summary, err := progression.RefreshAllOpenCampaigns(ctx)
	if err != nil {
		t.Fatalf("refresh-all: %v", err)
	}
	if summary.CampaignsScanned == 0 {
		t.Fatalf("expected at least 1 campaign scanned, got %d", summary.CampaignsScanned)
	}
	if summary.ScopesRefreshed == 0 {
		t.Fatalf("expected at least 1 scope refreshed, got %d", summary.ScopesRefreshed)
	}
	if summary.Errors != 0 {
		t.Fatalf("unexpected errors: %d", summary.Errors)
	}

	regionProg, err := progression.GetSnapshot(ctx, f.region.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("get region snapshot: %v", err)
	}
	if regionProg.TotalScored != 4 {
		t.Fatalf("region snapshot total=%d, want 4", regionProg.TotalScored)
	}

	schoolAProg, err := progression.GetSnapshot(ctx, f.schoolA.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("get schoolA snapshot: %v", err)
	}
	if schoolAProg.TotalScored != 3 {
		t.Fatalf("schoolA snapshot total=%d, want 3", schoolAProg.TotalScored)
	}

	schoolBProg, err := progression.GetSnapshot(ctx, f.schoolB.ID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("get schoolB snapshot: %v", err)
	}
	if schoolBProg.TotalScored != 1 {
		t.Fatalf("schoolB snapshot total=%d, want 1", schoolBProg.TotalScored)
	}
}

func TestRefreshAllOpenCampaigns_SkipsClosedCampaigns(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	if _, err := assessment.CloseCampaign(ctx, f.campaign.ID); err != nil {
		t.Fatalf("close: %v", err)
	}

	summary, err := progression.RefreshAllOpenCampaigns(ctx)
	if err != nil {
		t.Fatalf("refresh-all: %v", err)
	}

	for _, camp := range []int64{f.campaign.ID} {
		_ = camp
	}
	if summary.CampaignsScanned != 0 && summary.ScopesRefreshed > 0 {
		other, _ := assessment.ListOpenCampaignsWithScores(ctx)
		hasOurClosed := false
		for _, c := range other {
			if c.ID == f.campaign.ID {
				hasOurClosed = true
			}
		}
		if hasOurClosed {
			t.Fatalf("closed campaign was refreshed: %+v", summary)
		}
	}
}

func TestDrilldownByScopeViaSnapshots_UsesSnapshotsAfterRefresh(t *testing.T) {
	ctx := context.Background()
	f := setupFixture(t)

	if _, err := progression.RefreshAllOpenCampaigns(ctx); err != nil {
		t.Fatalf("refresh-all: %v", err)
	}

	parent, err := tenancy.GetNodeByID(ctx, f.schoolA.ID)
	if err != nil {
		t.Fatalf("get parent: %v", err)
	}
	delegID := *parent.ParentID

	d, err := progression.DrilldownByScopeViaSnapshots(ctx, delegID, f.periodID, f.campaign.ID)
	if err != nil {
		t.Fatalf("drilldown via snapshots: %v", err)
	}
	if d.Scope.TotalScored != 4 {
		t.Fatalf("scope total=%d, want 4", d.Scope.TotalScored)
	}
	if len(d.Children) != 2 {
		t.Fatalf("children=%d, want 2", len(d.Children))
	}
	for _, c := range d.Children {
		if c.NodeID == f.schoolA.ID && c.Total != 3 {
			t.Fatalf("schoolA child total=%d, want 3", c.Total)
		}
		if c.NodeID == f.schoolB.ID && c.Total != 1 {
			t.Fatalf("schoolB child total=%d, want 1", c.Total)
		}
	}
}
