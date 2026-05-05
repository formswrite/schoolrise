package hierarchy

type Node struct {
	ID       int64
	ParentID *int64
	Level    string
	Code     string
	Label    string
}

type ClosureRow struct {
	AncestorID   int64
	DescendantID int64
	Depth        int
}

func IsAncestor(rows []ClosureRow, ancestorID, descendantID int64) bool {
	for _, r := range rows {
		if r.AncestorID == ancestorID && r.DescendantID == descendantID {
			return true
		}
	}

	return false
}
