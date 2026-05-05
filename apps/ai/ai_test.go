package ai_test

import (
	"context"
	"errors"
	"testing"

	"encore.app/apps/ai"
)

func resetProvider(t *testing.T) {
	t.Helper()
	ai.SetTestRegistry(nil)
	ai.SetStubMode(true)
	t.Cleanup(func() {
		ai.SetStubMode(false)
		ai.SetTestRegistry(nil)
	})
}

func TestSuggestItems_Validation(t *testing.T) {
	resetProvider(t)
	if _, err := ai.SuggestItems(context.Background(), ai.SuggestItemsParams{Topic: "  "}); !errors.Is(err, ai.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestSuggestItems_StubModeRejectsLiveCall(t *testing.T) {
	resetProvider(t)
	_, err := ai.SuggestItems(context.Background(), ai.SuggestItemsParams{
		Topic: "Basic arithmetic", ScaleCode: "maths_5level", NiveauLabel: "CE1", Count: 2,
	})
	if !errors.Is(err, ai.ErrNoProvider) {
		t.Fatalf("err = %v, want ErrNoProvider", err)
	}
}

func TestDraftRubric_ValidationEmptyTitle(t *testing.T) {
	resetProvider(t)
	if _, err := ai.DraftRubric(context.Background(), ai.DraftRubricParams{}); !errors.Is(err, ai.ErrInvalidInput) {
		t.Fatalf("empty title: err = %v", err)
	}
}

func TestDraftRubric_ValidationEmptyBands(t *testing.T) {
	resetProvider(t)
	if _, err := ai.DraftRubric(context.Background(), ai.DraftRubricParams{QuestionTitle: "x"}); !errors.Is(err, ai.ErrInvalidInput) {
		t.Fatalf("no bands: err = %v", err)
	}
}

func TestGradeEssay_ValidationEmptyAnswer(t *testing.T) {
	resetProvider(t)
	_, err := ai.GradeEssay(context.Background(), ai.GradeEssayParams{
		QuestionTitle: "Q",
		Rubric:        []ai.RubricBand{{BandCode: "x", MinScore: 0, MaxScore: 100}},
	})
	if !errors.Is(err, ai.ErrInvalidInput) {
		t.Fatalf("empty answer: err = %v", err)
	}
}

func TestGradeEssay_ValidationEmptyRubric(t *testing.T) {
	resetProvider(t)
	_, err := ai.GradeEssay(context.Background(), ai.GradeEssayParams{
		QuestionTitle: "Q",
		StudentAnswer: "A",
	})
	if !errors.Is(err, ai.ErrInvalidInput) {
		t.Fatalf("empty rubric: err = %v", err)
	}
}

func TestGenerateDistractors_Validation(t *testing.T) {
	resetProvider(t)
	if _, err := ai.GenerateDistractors(context.Background(), ai.GenerateDistractorsParams{}); !errors.Is(err, ai.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestGenerateDistractors_StubModeRejectsLiveCall(t *testing.T) {
	resetProvider(t)
	_, err := ai.GenerateDistractors(context.Background(), ai.GenerateDistractorsParams{
		QuestionTitle: "Capital of Guinea?",
		CorrectAnswer: "Conakry",
		Count:         3,
	})
	if !errors.Is(err, ai.ErrNoProvider) {
		t.Fatalf("err = %v, want ErrNoProvider", err)
	}
}
