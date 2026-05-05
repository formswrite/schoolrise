package tenancy_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"encore.app/apps/tenancy"
)

func TestListLevels_ReturnsSeededLevels(t *testing.T) {

	levels, err := tenancy.ListLevels(context.Background())
	if err != nil {
		t.Fatalf("ListLevels: %v", err)
	}

	codes := make([]string, 0, len(levels))
	for _, l := range levels {
		codes = append(codes, l.Code)
	}

	sort.Strings(codes)

	want := []string{"class", "delegation", "group", "institution", "prefecture", "region"}

	if len(codes) != len(want) {
		t.Fatalf("got %d levels, want %d (codes=%v)", len(codes), len(want), codes)
	}

	for i, c := range codes {
		if c != want[i] {
			t.Errorf("codes[%d] = %q, want %q", i, c, want[i])
		}
	}
}

func TestGetLevel_ReturnsLevelDef(t *testing.T) {

	def, err := tenancy.GetLevel(context.Background(), tenancy.LevelPrefecture)
	if err != nil {
		t.Fatalf("GetLevel: %v", err)
	}

	if def.Code != tenancy.LevelPrefecture {
		t.Errorf("Code = %q, want %q", def.Code, tenancy.LevelPrefecture)
	}

	if def.ParentLevel != tenancy.LevelRegion {
		t.Errorf("ParentLevel = %q, want %q", def.ParentLevel, tenancy.LevelRegion)
	}

	if def.Depth != 1 {
		t.Errorf("Depth = %d, want 1", def.Depth)
	}
}

func TestGetLevel_RootHasNoParent(t *testing.T) {

	def, err := tenancy.GetLevel(context.Background(), tenancy.LevelRegion)
	if err != nil {
		t.Fatalf("GetLevel: %v", err)
	}

	if def.ParentLevel != "" {
		t.Errorf("root parent should be empty, got %q", def.ParentLevel)
	}

	if def.Depth != 0 {
		t.Errorf("root depth should be 0, got %d", def.Depth)
	}
}

func TestGetLevel_UnknownLevel(t *testing.T) {

	_, err := tenancy.GetLevel(context.Background(), "no-such-level-zzz")
	if !errors.Is(err, tenancy.ErrUnknownLevel) {
		t.Errorf("expected ErrUnknownLevel, got %v", err)
	}
}

//nolint:paralleltest // ReloadLevels mutates global registry
func TestReloadLevels_ReflectsDBChanges(t *testing.T) {
	if err := tenancy.ReloadLevels(context.Background()); err != nil {
		t.Fatalf("ReloadLevels: %v", err)
	}

	def, err := tenancy.GetLevel(context.Background(), tenancy.LevelInstitution)
	if err != nil {
		t.Fatalf("GetLevel after reload: %v", err)
	}

	if def.ParentLevel != tenancy.LevelDelegation {
		t.Errorf("ParentLevel = %q, want %q", def.ParentLevel, tenancy.LevelDelegation)
	}
}
