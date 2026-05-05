package tenancy_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"encore.app/apps/tenancy"
)

var nodeCounter atomic.Uint64

func uniqueCode(t *testing.T) string {
	t.Helper()

	n := nodeCounter.Add(1)

	return fmt.Sprintf("test-%s-%d", strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")), n)
}

func TestCreateNode_Region(t *testing.T) {

	ctx := context.Background()

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion,
		Code:  uniqueCode(t),
		Label: "Test Region",
	})
	if err != nil {
		t.Fatalf("CreateNode region: %v", err)
	}

	if region.ID == 0 {
		t.Fatal("ID not assigned")
	}

	if region.ParentID != nil {
		t.Errorf("region should have no parent, got %v", *region.ParentID)
	}

	if region.Level != tenancy.LevelRegion {
		t.Errorf("Level = %q, want %q", region.Level, tenancy.LevelRegion)
	}
}

func TestCreateNode_PrefectureUnderRegion(t *testing.T) {

	ctx := context.Background()

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Region",
	})
	if err != nil {
		t.Fatalf("create region: %v", err)
	}

	prefecture, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID,
		Level:    tenancy.LevelPrefecture,
		Code:     uniqueCode(t),
		Label:    "Prefecture",
	})
	if err != nil {
		t.Fatalf("create prefecture: %v", err)
	}

	if prefecture.ParentID == nil || *prefecture.ParentID != region.ID {
		t.Errorf("prefecture.ParentID = %v, want %d", prefecture.ParentID, region.ID)
	}
}

func TestCreateNode_RejectsInvalidLevelTransition(t *testing.T) {

	ctx := context.Background()

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Region",
	})
	if err != nil {
		t.Fatalf("create region: %v", err)
	}

	_, err = tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID,
		Level:    tenancy.LevelInstitution,
		Code:     uniqueCode(t),
		Label:    "Skips levels",
	})
	if !errors.Is(err, tenancy.ErrInvalidLevelTransition) {
		t.Errorf("expected ErrInvalidLevelTransition, got %v", err)
	}
}

func TestCreateNode_RejectsRegionWithParent(t *testing.T) {

	ctx := context.Background()

	r1, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "R1",
	})
	if err != nil {
		t.Fatalf("create r1: %v", err)
	}

	_, err = tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &r1.ID,
		Level:    tenancy.LevelRegion,
		Code:     uniqueCode(t),
		Label:    "Region under region (illegal)",
	})
	if !errors.Is(err, tenancy.ErrInvalidLevelTransition) {
		t.Errorf("expected ErrInvalidLevelTransition, got %v", err)
	}
}

func TestCreateNode_RejectsDuplicateCodeUnderSameParent(t *testing.T) {

	ctx := context.Background()

	region, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Region",
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}

	code := uniqueCode(t)

	_, err = tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID, Level: tenancy.LevelPrefecture, Code: code, Label: "P1",
	})
	if err != nil {
		t.Fatalf("first prefecture: %v", err)
	}

	_, err = tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		ParentID: &region.ID, Level: tenancy.LevelPrefecture, Code: code, Label: "P2",
	})
	if !errors.Is(err, tenancy.ErrCodeAlreadyExists) {
		t.Errorf("expected ErrCodeAlreadyExists, got %v", err)
	}
}

func TestCreateNode_RejectsMissingLabel(t *testing.T) {

	_, err := tenancy.CreateNode(context.Background(), tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "",
	})
	if !errors.Is(err, tenancy.ErrInvalidNodeInput) {
		t.Errorf("expected ErrInvalidNodeInput, got %v", err)
	}
}

func TestGetNodeByID_RoundTrip(t *testing.T) {

	ctx := context.Background()

	created, err := tenancy.CreateNode(ctx, tenancy.CreateNodeParams{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Round-trip",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	fetched, err := tenancy.GetNodeByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetNodeByID: %v", err)
	}

	if fetched.ID != created.ID || fetched.Label != created.Label {
		t.Errorf("fetched %+v != created %+v", fetched, created)
	}
}

func TestGetNodeByID_NotFound(t *testing.T) {

	_, err := tenancy.GetNodeByID(context.Background(), 999_999_999)
	if !errors.Is(err, tenancy.ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}
