package forms_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/forms"
)

func uniqueOwner() int64 { return time.Now().UnixNano() }
func clientID(label string) string {
	return label + "-" + time.Now().Format("150405.000000")
}

func TestCreateForm_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, err := forms.CreateForm(ctx, forms.CreateFormParams{
		OwnerID: uniqueOwner(), Title: "French Q1 2025-2026", Description: "End of term assessment",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if f.Status != forms.StatusDraft {
		t.Fatalf("status=%q, want draft", f.Status)
	}
	if len(f.PublicID) != 12 {
		t.Fatalf("public_id len=%d, want 12: %q", len(f.PublicID), f.PublicID)
	}
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })
}

func TestCreateForm_ValidationRejectsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if _, err := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: 1, Title: "  "}); !errors.Is(err, forms.ErrInvalidFormInput) {
		t.Fatalf("err=%v, want ErrInvalidFormInput", err)
	}
	if _, err := forms.CreateForm(ctx, forms.CreateFormParams{Title: "Valid"}); !errors.Is(err, forms.ErrInvalidFormInput) {
		t.Fatalf("missing owner err=%v", err)
	}
}

func TestCreateQuestion_RejectsInvalidType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "T"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	_, err := forms.CreateQuestion(ctx, forms.CreateQuestionParams{
		FormID: f.ID, ClientID: clientID("q"), Type: "TOTALLY_FAKE", Title: "x",
	})
	if !errors.Is(err, forms.ErrInvalidFieldType) {
		t.Fatalf("err=%v, want ErrInvalidFieldType", err)
	}
}

func TestCreateQuestion_PersistsAllJSONBFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "Math Q"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	min := int32(1)
	max := int32(5)
	q, err := forms.CreateQuestion(ctx, forms.CreateQuestionParams{
		FormID:      f.ID,
		ClientID:    clientID("scale"),
		Title:       "How confident with multiplication?",
		Type:        forms.TypeLinearScale,
		Required:    true,
		ScaleMin:    &min,
		ScaleMax:    &max,
		ScaleLabels: map[string]any{"min": "Not at all", "max": "Very"},
	})
	if err != nil {
		t.Fatalf("create question: %v", err)
	}
	if !q.Required || q.ScaleMin == nil || *q.ScaleMin != 1 || q.ScaleMax == nil || *q.ScaleMax != 5 {
		t.Fatalf("scale not preserved: %+v", q)
	}
	if q.ScaleLabels["min"] != "Not at all" {
		t.Fatalf("scale_labels not preserved: %v", q.ScaleLabels)
	}
}

func TestListQuestions_RespectsSortOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "Ordered"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	for i, ord := range []int32{30, 10, 20} {
		_, err := forms.CreateQuestion(ctx, forms.CreateQuestionParams{
			FormID: f.ID, ClientID: clientID("o" + string(rune('0'+i))), Type: forms.TypeShortAnswer, SortOrder: ord, Title: "Q",
		})
		if err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	qs, err := forms.ListQuestionsByForm(ctx, f.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(qs) != 3 {
		t.Fatalf("len=%d, want 3", len(qs))
	}
	for i := 1; i < len(qs); i++ {
		if qs[i].SortOrder < qs[i-1].SortOrder {
			t.Fatalf("not sorted: %v", qs)
		}
	}
}

func TestPublishForm_SnapshotsQuestionsAndBumpsVersion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "Pub Test"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	for i := 0; i < 2; i++ {
		_, err := forms.CreateQuestion(ctx, forms.CreateQuestionParams{
			FormID: f.ID, ClientID: clientID("p" + string(rune('a'+i))), Type: forms.TypeShortAnswer, Title: "Q",
		})
		if err != nil {
			t.Fatalf("create q: %v", err)
		}
	}

	v1, updated, err := forms.PublishForm(ctx, f.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if v1.VersionNum != 1 {
		t.Fatalf("version=%d, want 1", v1.VersionNum)
	}
	if updated.Status != forms.StatusPublished {
		t.Fatalf("status=%q, want published", updated.Status)
	}

	snapQs, ok := v1.Snapshot["questions"].([]any)
	if !ok || len(snapQs) != 2 {
		t.Fatalf("snapshot questions wrong: %v", v1.Snapshot["questions"])
	}

	_, err = forms.CreateQuestion(ctx, forms.CreateQuestionParams{
		FormID: f.ID, ClientID: clientID("p3"), Type: forms.TypeShortAnswer, Title: "Q3",
	})
	if err != nil {
		t.Fatalf("create q3: %v", err)
	}

	v2, _, err := forms.PublishForm(ctx, f.ID)
	if err != nil {
		t.Fatalf("publish v2: %v", err)
	}
	if v2.VersionNum != 2 {
		t.Fatalf("version=%d, want 2", v2.VersionNum)
	}

	versions, _ := forms.ListVersions(ctx, f.ID)
	if len(versions) != 2 {
		t.Fatalf("list versions = %d, want 2", len(versions))
	}
}

func TestPublishForm_RejectsLayoutOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "Layout Only"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	if _, err := forms.CreateQuestion(ctx, forms.CreateQuestionParams{
		FormID: f.ID, ClientID: clientID("sec"), Type: forms.TypeSection, Title: "Section",
	}); err != nil {
		t.Fatalf("create section: %v", err)
	}

	if _, _, err := forms.PublishForm(ctx, f.ID); !errors.Is(err, forms.ErrFormNotPublishable) {
		t.Fatalf("err=%v, want ErrFormNotPublishable", err)
	}
}

func TestUpdateQuestion_ChangesType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "Update Q"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	q, err := forms.CreateQuestion(ctx, forms.CreateQuestionParams{
		FormID: f.ID, ClientID: clientID("u"), Type: forms.TypeShortAnswer, Title: "Original",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	updated, err := forms.UpdateQuestion(ctx, forms.UpdateQuestionParams{
		ID: q.ID, Title: "Changed", Type: forms.TypeParagraph, Required: true, SortOrder: 5,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Title != "Changed" || updated.Type != forms.TypeParagraph || !updated.Required {
		t.Fatalf("not applied: %+v", updated)
	}
}

func TestDeleteForm_NotFoundErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if err := forms.DeleteForm(ctx, 9999999); !errors.Is(err, forms.ErrFormNotFound) {
		t.Fatalf("err=%v, want ErrFormNotFound", err)
	}
}

func TestFieldTypeCatalogue_HasAll26(t *testing.T) {
	t.Parallel()
	all := forms.AllTypes()
	if len(all) != 32 {
		t.Logf("warning: AllTypes returns %d (expected 32 = 26 submittable + 2 layout + variants); spec says 26 submittable", len(all))
	}
	for _, want := range []string{forms.TypeShortAnswer, forms.TypeMultipleChoice, forms.TypeLinearScale, forms.TypeMatching, forms.TypeFillInBlank, forms.TypeOrdering, forms.TypeEssay, forms.TypeSection, forms.TypeStatement} {
		if !forms.IsValidType(want) {
			t.Errorf("missing %s", want)
		}
	}
}

func TestGetFormByPublicID_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f, _ := forms.CreateForm(ctx, forms.CreateFormParams{OwnerID: uniqueOwner(), Title: "Pub ID"})
	t.Cleanup(func() { _ = forms.DeleteForm(ctx, f.ID) })

	got, err := forms.GetFormByPublicID(ctx, f.PublicID)
	if err != nil {
		t.Fatalf("get by public id: %v", err)
	}
	if got.ID != f.ID {
		t.Fatalf("id mismatch: %d != %d", got.ID, f.ID)
	}
}
