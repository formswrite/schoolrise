package tenancy_test

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"

	encauth "encore.dev/beta/auth"
	"encore.dev/et"

	"encore.app/apps/auth"
	"encore.app/apps/tenancy"
)

var apiUserCounter atomic.Int64

func authedCtx(t *testing.T) context.Context {
	t.Helper()

	n := apiUserCounter.Add(1)

	user, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email:    fmt.Sprintf("api-user-%d-%s@local.test", n, t.Name()),
		Password: "Pass1234!",
		FullName: "API Tester",
		Role:     "admin",
	})
	if err != nil {
		t.Fatalf("seed api user: %v", err)
	}

	if err := auth.AssignRole(context.Background(), user.ID, "admin", nil); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	et.OverrideAuthInfo(encauth.UID(strconv.FormatInt(user.ID, 10)), &auth.AuthData{UserID: user.ID, Role: "admin"})

	return context.Background()
}

func newAPIService(t *testing.T) *tenancy.Service {
	t.Helper()
	return &tenancy.Service{}
}

func TestListLevels_ReturnsRegisteredLevels(t *testing.T) {

	if err := tenancy.ReloadLevels(context.Background()); err != nil {
		t.Fatalf("ReloadLevels: %v", err)
	}

	resp, err := newAPIService(t).ListLevelsAPI(authedCtx(t))
	if err != nil {
		t.Fatalf("ListLevels: %v", err)
	}

	if len(resp.Levels) == 0 {
		t.Error("expected at least one level")
	}

	for _, l := range resp.Levels {
		if l.Code == "" {
			t.Errorf("level has empty code: %+v", l)
		}
	}
}

func TestCreateNode_RoundTrip(t *testing.T) {

	created, err := newAPIService(t).CreateNodeAPI(authedCtx(t), &tenancy.CreateNodeAPIRequest{
		Level: tenancy.LevelRegion,
		Code:  uniqueCode(t),
		Label: "API Region",
	})
	if err != nil {
		t.Fatalf("CreateNode: %v", err)
	}

	if created.Node.ID == 0 {
		t.Fatal("ID not assigned")
	}

	fetched, err := newAPIService(t).GetNodeAPI(authedCtx(t), created.Node.ID)
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}

	if fetched.Node.Label != "API Region" {
		t.Errorf("Label = %q, want %q", fetched.Node.Label, "API Region")
	}
}

func TestListNodes_FilterByParent(t *testing.T) {

	region, err := newAPIService(t).CreateNodeAPI(authedCtx(t), &tenancy.CreateNodeAPIRequest{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Filter Region",
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}

	child, err := newAPIService(t).CreateNodeAPI(authedCtx(t), &tenancy.CreateNodeAPIRequest{
		ParentID: &region.Node.ID,
		Level:    tenancy.LevelPrefecture,
		Code:     uniqueCode(t),
		Label:    "Filter Prefecture",
	})
	if err != nil {
		t.Fatalf("prefecture: %v", err)
	}

	resp, err := newAPIService(t).ListNodesAPI(authedCtx(t), &tenancy.ListNodesRequest{
		ParentID: region.Node.ID,
	})
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}

	found := false
	for _, n := range resp.Nodes {
		if n.ID == child.Node.ID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("child node %d not in parent's listing", child.Node.ID)
	}
}

func TestDeleteNode_SoftDeletesLeaf(t *testing.T) {

	region, err := newAPIService(t).CreateNodeAPI(authedCtx(t), &tenancy.CreateNodeAPIRequest{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Delete Region",
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}

	if err := newAPIService(t).DeleteNodeAPI(authedCtx(t), region.Node.ID); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}

	_, err = newAPIService(t).GetNodeAPI(authedCtx(t), region.Node.ID)
	if err == nil {
		t.Error("expected NotFound after delete")
	}
}

func TestGetNode_NotFound(t *testing.T) {

	_, err := newAPIService(t).GetNodeAPI(authedCtx(t), 999_999_998)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateNode_RejectsInvalidLevel(t *testing.T) {

	_, err := newAPIService(t).CreateNodeAPI(authedCtx(t), &tenancy.CreateNodeAPIRequest{
		Level: "no-such-level",
		Code:  uniqueCode(t),
		Label: "X",
	})
	if err == nil {
		t.Fatal("expected error on unknown level")
	}
}

func TestScopedUser_OnlySeesDescendants(t *testing.T) {

	adminCtx := authedCtx(t)

	region, err := newAPIService(t).CreateNodeAPI(adminCtx, &tenancy.CreateNodeAPIRequest{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Scope Region A",
	})
	if err != nil {
		t.Fatalf("region: %v", err)
	}

	otherRegion, err := newAPIService(t).CreateNodeAPI(adminCtx, &tenancy.CreateNodeAPIRequest{
		Level: tenancy.LevelRegion, Code: uniqueCode(t), Label: "Scope Region B",
	})
	if err != nil {
		t.Fatalf("otherRegion: %v", err)
	}

	prefA, err := newAPIService(t).CreateNodeAPI(adminCtx, &tenancy.CreateNodeAPIRequest{
		ParentID: &region.Node.ID, Level: tenancy.LevelPrefecture, Code: uniqueCode(t), Label: "Pref under A",
	})
	if err != nil {
		t.Fatalf("prefA: %v", err)
	}

	n := apiUserCounter.Add(1)
	scopedUser, err := auth.CreateUser(context.Background(), auth.CreateUserParams{
		Email:    fmt.Sprintf("scoped-%d@local.test", n),
		Password: "Pass1234!",
		FullName: "Scoped Inspector",
		Role:     "inspector",
	})
	if err != nil {
		t.Fatalf("scoped user: %v", err)
	}

	if err := auth.AssignRole(context.Background(), scopedUser.ID, "inspector", &region.Node.ID); err != nil {
		t.Fatalf("AssignRole scoped: %v", err)
	}

	et.OverrideAuthInfo(encauth.UID(strconv.FormatInt(scopedUser.ID, 10)), &auth.AuthData{UserID: scopedUser.ID, Role: "inspector"})

	if _, err := newAPIService(t).GetNodeAPI(context.Background(), prefA.Node.ID); err != nil {
		t.Errorf("scoped user should see descendant prefA: %v", err)
	}

	if _, err := newAPIService(t).GetNodeAPI(context.Background(), otherRegion.Node.ID); err == nil {
		t.Error("scoped user should NOT see other region")
	}
}
