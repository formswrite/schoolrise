package countries

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

var (
	ErrPackNotFound = errors.New("countries: pack not found")
	ErrInvalidPack  = errors.New("countries: invalid pack")
)

//go:embed *.json
var packsFS embed.FS

type Level struct {
	Code   string `json:"code"`
	Label  string `json:"label"`
	Parent string `json:"parent"`
	Depth  int    `json:"depth"`
	Sort   int    `json:"sort"`
}

type SeedNode struct {
	Level    string     `json:"level"`
	Code     string     `json:"code"`
	Label    string     `json:"label"`
	Children []SeedNode `json:"children,omitempty"`
}

type Pack struct {
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	DefaultLocale string     `json:"default_locale"`
	Levels        []Level    `json:"levels"`
	SeedNodes     []SeedNode `json:"seed_nodes"`
}

func List() ([]Pack, error) {
	entries, err := fs.ReadDir(packsFS, ".")
	if err != nil {
		return nil, err
	}

	out := make([]Pack, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		raw, err := packsFS.ReadFile(e.Name())
		if err != nil {
			return nil, err
		}

		var pack Pack
		if err := json.Unmarshal(raw, &pack); err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrInvalidPack, e.Name(), err)
		}

		if err := pack.Validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}

		out = append(out, pack)
	}

	return out, nil
}

func Get(code string) (*Pack, error) {
	packs, err := List()
	if err != nil {
		return nil, err
	}

	for _, p := range packs {
		if p.Code == code {
			pack := p

			return &pack, nil
		}
	}

	return nil, ErrPackNotFound
}

func (p *Pack) Validate() error {
	if strings.TrimSpace(p.Code) == "" || strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: code and name required", ErrInvalidPack)
	}

	seen := map[string]Level{}

	for _, l := range p.Levels {
		if strings.TrimSpace(l.Code) == "" || strings.TrimSpace(l.Label) == "" {
			return fmt.Errorf("%w: level code and label required", ErrInvalidPack)
		}

		if _, dup := seen[l.Code]; dup {
			return fmt.Errorf("%w: duplicate level code %q", ErrInvalidPack, l.Code)
		}

		seen[l.Code] = l
	}

	for _, l := range p.Levels {
		if l.Parent == "" {
			continue
		}

		if _, ok := seen[l.Parent]; !ok {
			return fmt.Errorf("%w: level %q references unknown parent %q", ErrInvalidPack, l.Code, l.Parent)
		}
	}

	if err := detectCycle(seen); err != nil {
		return err
	}

	for _, n := range p.SeedNodes {
		if err := validateSeedNode(seen, n); err != nil {
			return err
		}
	}

	return nil
}

func detectCycle(levels map[string]Level) error {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[string]int, len(levels))

	var visit func(code string) error
	visit = func(code string) error {
		switch state[code] {
		case visiting:
			return fmt.Errorf("%w: cycle detected at level %q", ErrInvalidPack, code)
		case visited:
			return nil
		}

		state[code] = visiting

		l := levels[code]
		if l.Parent != "" {
			if err := visit(l.Parent); err != nil {
				return err
			}
		}

		state[code] = visited

		return nil
	}

	for code := range levels {
		if err := visit(code); err != nil {
			return err
		}
	}

	return nil
}

func validateSeedNode(levels map[string]Level, n SeedNode) error {
	if strings.TrimSpace(n.Code) == "" || strings.TrimSpace(n.Label) == "" {
		return fmt.Errorf("%w: seed node code and label required", ErrInvalidPack)
	}

	if _, ok := levels[n.Level]; !ok {
		return fmt.Errorf("%w: seed node %q references unknown level %q", ErrInvalidPack, n.Code, n.Level)
	}

	for _, c := range n.Children {
		if err := validateSeedNode(levels, c); err != nil {
			return err
		}
	}

	return nil
}
