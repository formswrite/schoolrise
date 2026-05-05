package tenancy_test

import (
	"context"
	"errors"
	"testing"

	"encore.app/apps/tenancy"
)

type fixture struct {
	region      *tenancy.Node
	prefA       *tenancy.Node
	prefB       *tenancy.Node
	delegA      *tenancy.Node
	institution *tenancy.Node
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	ctx := context.Background()

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Region",
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}

	prefA, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID, Level: tenancy.LevelPrefecture, Code: uniqueCode(t), Label: "Pref A",
	})
	if err != nil {
		t.Fatalf("prefA: %v", err)
	}

	prefB, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID, Level: tenancy.LevelPrefecture, Code: uniqueCode(t), Label: "Pref B",
	})
	if err != nil {
		t.Fatalf("prefB: %v", err)
	}

	delegA, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &prefA.ID, Level: tenancy.LevelDelegation, Code: uniqueCode(t), Label: "Deleg A",
	})
	if err != nil {
		t.Fatalf("delegA: %v", err)
	}

	institution, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &delegA.ID, Level: tenancy.LevelInstitution, Code: uniqueCode(t), Label: "School",
	})
	if err != nil {
		t.Fatalf("institution: %v", err)
	}

	return &fixture{region: region, prefA: prefA, prefB: prefB, delegA: delegA, institution: institution}
}

func TestListChildren_OnlyDirectChildren(t *testing.T) {

	f := newFixture(t)

	children, err := tenancy.ListChildren(context.Background(), &f.region.ID)
	if err != nil {
		t.Fatalf("ListChildren: %v", err)
	}

	gotIDs := map[int64]struct{}{}
	for _, c := range children {
		gotIDs[c.ID] = struct{}{}
	}

	if _, ok := gotIDs[f.prefA.ID]; !ok {
		t.Errorf("prefA missing from children of region")
	}

	if _, ok := gotIDs[f.prefB.ID]; !ok {
		t.Errorf("prefB missing from children of region")
	}

	if _, ok := gotIDs[f.delegA.ID]; ok {
		t.Errorf("delegation should not appear in direct children of region")
	}
}

func TestListRoots_NoParent(t *testing.T) {

	f := newFixture(t)

	roots, err := tenancy.ListChildren(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListChildren(nil): %v", err)
	}

	found := false
	for _, r := range roots {
		if r.ID == f.region.ID {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("seeded region not in root list")
	}
}

func TestDescendantsOf_FullSubtree(t *testing.T) {

	f := newFixture(t)

	descendants, err := tenancy.DescendantsOf(context.Background(), f.region.ID)
	if err != nil {
		t.Fatalf("DescendantsOf: %v", err)
	}

	gotIDs := map[int64]int{}
	for _, d := range descendants {
		gotIDs[d.Node.ID] = d.Depth
	}

	if gotIDs[f.region.ID] != 0 {
		t.Errorf("self should be at depth 0, got %d", gotIDs[f.region.ID])
	}

	if gotIDs[f.prefA.ID] != 1 {
		t.Errorf("prefA depth = %d, want 1", gotIDs[f.prefA.ID])
	}

	if gotIDs[f.delegA.ID] != 2 {
		t.Errorf("delegA depth = %d, want 2", gotIDs[f.delegA.ID])
	}

	if gotIDs[f.institution.ID] != 3 {
		t.Errorf("institution depth = %d, want 3", gotIDs[f.institution.ID])
	}
}

func TestSoftDelete_BlocksWhenChildrenExist(t *testing.T) {

	f := newFixture(t)

	err := tenancy.SoftDeleteNode(context.Background(), f.region.ID)
	if !errors.Is(err, tenancy.ErrNodeHasChildren) {
		t.Errorf("expected ErrNodeHasChildren, got %v", err)
	}
}

func TestSoftDelete_AllowsLeafDeletion(t *testing.T) {

	f := newFixture(t)

	if err := tenancy.SoftDeleteNode(context.Background(), f.institution.ID); err != nil {
		t.Fatalf("SoftDeleteNode leaf: %v", err)
	}

	_, err := tenancy.GetNodeByID(context.Background(), f.institution.ID)
	if !errors.Is(err, tenancy.ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound after soft delete, got %v", err)
	}
}
