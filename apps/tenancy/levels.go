package tenancy

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"

	"encore.app/apps/tenancy/dbtenancy"
)

var ErrUnknownLevel = errors.New("tenancy: unknown level")

const (
	LevelRegion      = "region"
	LevelPrefecture  = "prefecture"
	LevelDelegation  = "delegation"
	LevelInstitution = "institution"
	LevelClass       = "class"
	LevelGroup       = "group"
)

type LevelDef struct {
	Code        string
	Label       string
	ParentLevel string
	Depth       int
	SortOrder   int
}

var (
	levelsLoaded atomic.Bool
	levelsMu     sync.RWMutex
	levelsByCode = map[string]LevelDef{}
)

func ensureLevelsLoaded(ctx context.Context) error {
	if levelsLoaded.Load() {
		return nil
	}

	return ReloadLevels(ctx)
}

func ReloadLevels(ctx context.Context) error {
	rows, err := queries.ListHierarchyLevels(ctx)
	if err != nil {
		return err
	}

	next := make(map[string]LevelDef, len(rows))
	for _, r := range rows {
		next[r.Code] = LevelDef{
			Code:        r.Code,
			Label:       r.Label,
			ParentLevel: nullStringValue(r.ParentLevelCode),
			Depth:       int(r.Depth),
			SortOrder:   int(r.SortOrder),
		}
	}

	levelsMu.Lock()
	levelsByCode = next
	levelsMu.Unlock()
	levelsLoaded.Store(true)

	return nil
}

func ListLevels(ctx context.Context) ([]LevelDef, error) {
	if err := ensureLevelsLoaded(ctx); err != nil {
		return nil, err
	}

	levelsMu.RLock()
	defer levelsMu.RUnlock()

	out := make([]LevelDef, 0, len(levelsByCode))
	for _, l := range levelsByCode {
		out = append(out, l)
	}

	return out, nil
}

func GetLevel(ctx context.Context, code string) (LevelDef, error) {
	if err := ensureLevelsLoaded(ctx); err != nil {
		return LevelDef{}, err
	}

	levelsMu.RLock()
	defer levelsMu.RUnlock()

	def, ok := levelsByCode[code]
	if !ok {
		return LevelDef{}, ErrUnknownLevel
	}

	return def, nil
}

func ApplyLevels(ctx context.Context, defs []LevelDef) error {
	for _, d := range defs {
		var parent sql.NullString
		if d.ParentLevel != "" {
			parent = sql.NullString{String: d.ParentLevel, Valid: true}
		}

		if err := queries.UpsertHierarchyLevel(ctx, dbtenancy.UpsertHierarchyLevelParams{
			Code:            d.Code,
			Label:           d.Label,
			ParentLevelCode: parent,
			Depth:           int32(d.Depth),
			SortOrder:       int32(d.SortOrder),
		}); err != nil {
			return err
		}
	}

	return ReloadLevels(ctx)
}

func nullStringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}

	return ""
}
