package hierarchy_test

import (
	"testing"

	"encore.app/internal/hierarchy"
)

func TestIsAncestor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		rows         []hierarchy.ClosureRow
		ancestorID   int64
		descendantID int64
		expected     bool
	}{
		{
			name:         "empty rows",
			rows:         []hierarchy.ClosureRow{},
			ancestorID:   1,
			descendantID: 2,
			expected:     false,
		},
		{
			name: "exact match present",
			rows: []hierarchy.ClosureRow{
				{AncestorID: 1, DescendantID: 2, Depth: 1},
			},
			ancestorID:   1,
			descendantID: 2,
			expected:     true,
		},
		{
			name: "ancestor matches but descendant does not",
			rows: []hierarchy.ClosureRow{
				{AncestorID: 1, DescendantID: 99, Depth: 1},
			},
			ancestorID:   1,
			descendantID: 2,
			expected:     false,
		},
		{
			name: "descendant matches but ancestor does not",
			rows: []hierarchy.ClosureRow{
				{AncestorID: 99, DescendantID: 2, Depth: 1},
			},
			ancestorID:   1,
			descendantID: 2,
			expected:     false,
		},
		{
			name: "multiple rows, one matches",
			rows: []hierarchy.ClosureRow{
				{AncestorID: 99, DescendantID: 2, Depth: 1},
				{AncestorID: 1, DescendantID: 2, Depth: 2},
				{AncestorID: 5, DescendantID: 7, Depth: 3},
			},
			ancestorID:   1,
			descendantID: 2,
			expected:     true,
		},
		{
			name: "self reference at depth zero",
			rows: []hierarchy.ClosureRow{
				{AncestorID: 5, DescendantID: 5, Depth: 0},
			},
			ancestorID:   5,
			descendantID: 5,
			expected:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := hierarchy.IsAncestor(tc.rows, tc.ancestorID, tc.descendantID)
			if got != tc.expected {
				t.Errorf("IsAncestor() = %v, want %v", got, tc.expected)
			}
		})
	}
}
