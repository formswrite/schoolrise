package academics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/academics"
)

func TestCreateNiveau_Success(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")
	n, err := academics.CreateNiveau(ctx, academics.CreateNiveauParams{
		Code: "CE1-" + suffix, Label: "Cours Élémentaire 1", SortOrder: 10,
	})
	if err != nil {
		t.Fatalf("CreateNiveau: %v", err)
	}
	if n.ID == 0 || n.SortOrder != 10 {
		t.Fatalf("bad niveau: %+v", n)
	}
	t.Cleanup(func() { _ = academics.DeleteNiveau(ctx, n.ID) })
}

func TestCreateNiveau_Validation(t *testing.T) {
	ctx := context.Background()
	if _, err := academics.CreateNiveau(ctx, academics.CreateNiveauParams{Code: " ", Label: "x"}); !errors.Is(err, academics.ErrInvalidNiveauInput) {
		t.Fatalf("empty code: err = %v", err)
	}
	if _, err := academics.CreateNiveau(ctx, academics.CreateNiveauParams{Code: "x", Label: ""}); !errors.Is(err, academics.ErrInvalidNiveauInput) {
		t.Fatalf("empty label: err = %v", err)
	}
}

func TestCreateNiveau_DuplicateCode(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")
	p := academics.CreateNiveauParams{Code: "dup-niv-" + suffix, Label: "Dup", SortOrder: 1}

	first, err := academics.CreateNiveau(ctx, p)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeleteNiveau(ctx, first.ID) })

	if _, err := academics.CreateNiveau(ctx, p); !errors.Is(err, academics.ErrNiveauCodeTaken) {
		t.Fatalf("err = %v, want ErrNiveauCodeTaken", err)
	}
}

func TestListNiveaux_SortedByOrderThenCode(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")

	created := []*academics.Niveau{}
	specs := []academics.CreateNiveauParams{
		{Code: "z-" + suffix, Label: "Z", SortOrder: 1},
		{Code: "a-" + suffix, Label: "A", SortOrder: 1},
		{Code: "m-" + suffix, Label: "M", SortOrder: 0},
	}
	for _, s := range specs {
		n, err := academics.CreateNiveau(ctx, s)
		if err != nil {
			t.Fatalf("create %s: %v", s.Code, err)
		}
		created = append(created, n)
	}
	t.Cleanup(func() {
		for _, n := range created {
			_ = academics.DeleteNiveau(ctx, n.ID)
		}
	})

	all, err := academics.ListNiveaux(ctx)
	if err != nil {
		t.Fatalf("ListNiveaux: %v", err)
	}

	pos := map[string]int{}
	for i, n := range all {
		pos[n.Code] = i
	}
	if pos["m-"+suffix] >= pos["a-"+suffix] {
		t.Fatalf("expected m (sort_order=0) before a (sort_order=1); positions=%v", pos)
	}
	if pos["a-"+suffix] >= pos["z-"+suffix] {
		t.Fatalf("expected a before z within sort_order=1; positions=%v", pos)
	}
}

func TestDeleteNiveau_NotFoundErrors(t *testing.T) {
	ctx := context.Background()
	if err := academics.DeleteNiveau(ctx, 999999999); !errors.Is(err, academics.ErrNiveauNotFound) {
		t.Fatalf("err = %v, want ErrNiveauNotFound", err)
	}
}
