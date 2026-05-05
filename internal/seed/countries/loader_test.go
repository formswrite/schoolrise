package countries_test

import (
	"errors"
	"testing"

	"encore.app/internal/seed/countries"
)

func TestList_IncludesTemplate(t *testing.T) {
	t.Parallel()

	packs, err := countries.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(packs) == 0 {
		t.Fatal("no packs embedded")
	}

	found := false
	for _, p := range packs {
		if p.Code == "TEMPLATE" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("template pack not found in embedded packs")
	}
}

func TestGet_ReturnsTemplate(t *testing.T) {
	t.Parallel()

	pack, err := countries.Get("TEMPLATE")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if pack.Code != "TEMPLATE" {
		t.Errorf("Code = %q, want TEMPLATE", pack.Code)
	}

	if pack.DefaultLocale != "en" {
		t.Errorf("DefaultLocale = %q, want en", pack.DefaultLocale)
	}

	if len(pack.Levels) != 0 {
		t.Errorf("template should have no levels, got %d", len(pack.Levels))
	}
}

func TestGet_UnknownPack(t *testing.T) {
	t.Parallel()

	_, err := countries.Get("DOES-NOT-EXIST")
	if !errors.Is(err, countries.ErrPackNotFound) {
		t.Errorf("expected ErrPackNotFound, got %v", err)
	}
}

func TestValidate_RejectsLevelWithUnknownParent(t *testing.T) {
	t.Parallel()

	pack := countries.Pack{
		Code:          "TEST",
		Name:          "Test",
		DefaultLocale: "en",
		Levels: []countries.Level{
			{Code: "school", Label: "School", Parent: "ghost", Depth: 1, Sort: 0},
		},
	}

	err := pack.Validate()
	if !errors.Is(err, countries.ErrInvalidPack) {
		t.Errorf("expected ErrInvalidPack, got %v", err)
	}
}

func TestValidate_RejectsCycle(t *testing.T) {
	t.Parallel()

	pack := countries.Pack{
		Code:          "TEST",
		Name:          "Test",
		DefaultLocale: "en",
		Levels: []countries.Level{
			{Code: "a", Label: "A", Parent: "b", Depth: 0, Sort: 0},
			{Code: "b", Label: "B", Parent: "a", Depth: 1, Sort: 1},
		},
	}

	err := pack.Validate()
	if !errors.Is(err, countries.ErrInvalidPack) {
		t.Errorf("expected ErrInvalidPack on cycle, got %v", err)
	}
}

func TestValidate_RejectsDuplicateLevelCode(t *testing.T) {
	t.Parallel()

	pack := countries.Pack{
		Code:          "TEST",
		Name:          "Test",
		DefaultLocale: "en",
		Levels: []countries.Level{
			{Code: "a", Label: "A", Parent: "", Depth: 0, Sort: 0},
			{Code: "a", Label: "A2", Parent: "", Depth: 0, Sort: 1},
		},
	}

	err := pack.Validate()
	if !errors.Is(err, countries.ErrInvalidPack) {
		t.Errorf("expected ErrInvalidPack on duplicate code, got %v", err)
	}
}

func TestValidate_AcceptsEmptyTemplate(t *testing.T) {
	t.Parallel()

	pack := countries.Pack{
		Code:          "TEMPLATE",
		Name:          "Empty",
		DefaultLocale: "en",
		Levels:        []countries.Level{},
		SeedNodes:     []countries.SeedNode{},
	}

	if err := pack.Validate(); err != nil {
		t.Errorf("empty template should be valid, got %v", err)
	}
}

func TestValidate_AcceptsLinearHierarchy(t *testing.T) {
	t.Parallel()

	pack := countries.Pack{
		Code:          "TEST",
		Name:          "Test",
		DefaultLocale: "en",
		Levels: []countries.Level{
			{Code: "region", Label: "Region", Parent: "", Depth: 0, Sort: 0},
			{Code: "school", Label: "School", Parent: "region", Depth: 1, Sort: 1},
		},
	}

	if err := pack.Validate(); err != nil {
		t.Errorf("linear hierarchy should be valid, got %v", err)
	}
}
