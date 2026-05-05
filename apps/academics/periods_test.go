package academics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/apps/academics"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return v
}

func TestCreatePeriod_Success(t *testing.T) {
	ctx := context.Background()
	p, err := academics.CreatePeriod(ctx, academics.CreatePeriodParams{
		Code:     "2025-2026",
		Label:    "Year 2025-2026",
		StartsOn: mustDate(t, "2025-09-01"),
		EndsOn:   mustDate(t, "2026-06-30"),
	})
	if err != nil {
		t.Fatalf("CreatePeriod: %v", err)
	}
	if p.ID == 0 || p.Code != "2025-2026" || p.IsCurrent {
		t.Fatalf("unexpected period: %+v", p)
	}
	t.Cleanup(func() { _ = academics.DeletePeriod(ctx, p.ID) })
}

func TestCreatePeriod_Validation(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		p    academics.CreatePeriodParams
		want error
	}{
		{"empty code", academics.CreatePeriodParams{Code: " ", Label: "x", StartsOn: mustDate(t, "2025-01-01"), EndsOn: mustDate(t, "2025-12-31")}, academics.ErrInvalidPeriodInput},
		{"empty label", academics.CreatePeriodParams{Code: "x", Label: "", StartsOn: mustDate(t, "2025-01-01"), EndsOn: mustDate(t, "2025-12-31")}, academics.ErrInvalidPeriodInput},
		{"missing starts_on", academics.CreatePeriodParams{Code: "x", Label: "y", EndsOn: mustDate(t, "2025-12-31")}, academics.ErrInvalidPeriodInput},
		{"end before start", academics.CreatePeriodParams{Code: "x", Label: "y", StartsOn: mustDate(t, "2025-12-31"), EndsOn: mustDate(t, "2025-01-01")}, academics.ErrPeriodDateRange},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := academics.CreatePeriod(ctx, tc.p)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestCreatePeriod_DuplicateCode(t *testing.T) {
	ctx := context.Background()
	base := academics.CreatePeriodParams{
		Code:     "dup-period-" + time.Now().Format("150405.000"),
		Label:    "Dup",
		StartsOn: mustDate(t, "2030-09-01"),
		EndsOn:   mustDate(t, "2031-06-30"),
	}
	first, err := academics.CreatePeriod(ctx, base)
	if err != nil {
		t.Fatalf("first CreatePeriod: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeletePeriod(ctx, first.ID) })

	if _, err := academics.CreatePeriod(ctx, base); !errors.Is(err, academics.ErrPeriodCodeTaken) {
		t.Fatalf("err = %v, want ErrPeriodCodeTaken", err)
	}
}

func TestSetPeriodCurrent_OnlyOneCurrent(t *testing.T) {
	ctx := context.Background()
	suffix := time.Now().Format("150405.000")
	a, err := academics.CreatePeriod(ctx, academics.CreatePeriodParams{
		Code:      "cur-a-" + suffix,
		Label:     "A",
		StartsOn:  mustDate(t, "2032-09-01"),
		EndsOn:    mustDate(t, "2033-06-30"),
		IsCurrent: true,
	})
	if err != nil {
		t.Fatalf("create A: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeletePeriod(ctx, a.ID) })

	b, err := academics.CreatePeriod(ctx, academics.CreatePeriodParams{
		Code:      "cur-b-" + suffix,
		Label:     "B",
		StartsOn:  mustDate(t, "2033-09-01"),
		EndsOn:    mustDate(t, "2034-06-30"),
		IsCurrent: true,
	})
	if err != nil {
		t.Fatalf("create B: %v", err)
	}
	t.Cleanup(func() { _ = academics.DeletePeriod(ctx, b.ID) })

	cur, err := academics.GetCurrentPeriod(ctx)
	if err != nil {
		t.Fatalf("GetCurrentPeriod: %v", err)
	}
	if cur.ID != b.ID {
		t.Fatalf("current = %d, want %d (B)", cur.ID, b.ID)
	}

	switched, err := academics.SetPeriodCurrent(ctx, a.ID)
	if err != nil {
		t.Fatalf("SetPeriodCurrent: %v", err)
	}
	if switched.ID != a.ID || !switched.IsCurrent {
		t.Fatalf("switched = %+v", switched)
	}

	cur, err = academics.GetCurrentPeriod(ctx)
	if err != nil {
		t.Fatalf("GetCurrentPeriod after switch: %v", err)
	}
	if cur.ID != a.ID {
		t.Fatalf("current after switch = %d, want %d (A)", cur.ID, a.ID)
	}
}

func TestDeletePeriod(t *testing.T) {
	ctx := context.Background()
	p, err := academics.CreatePeriod(ctx, academics.CreatePeriodParams{
		Code:     "del-period-" + time.Now().Format("150405.000"),
		Label:    "Del",
		StartsOn: mustDate(t, "2040-09-01"),
		EndsOn:   mustDate(t, "2041-06-30"),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := academics.DeletePeriod(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := academics.GetPeriodByID(ctx, p.ID); !errors.Is(err, academics.ErrPeriodNotFound) {
		t.Fatalf("after delete err = %v, want ErrPeriodNotFound", err)
	}
}
