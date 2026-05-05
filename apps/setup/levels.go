package setup

import (
	"context"
	"encoding/json"
	"errors"

	"encore.dev/beta/errs"

	"encore.app/apps/tenancy"
	"encore.app/internal/seed/countries"
)

type SetLevelsRequest struct {
	SessionToken string
	PackCode     string
	Levels       []SetLevelsLevel
}

type SetLevelsLevel struct {
	Code   string
	Label  string
	Parent string
	Depth  int
	Sort   int
}

//encore:api public method=POST path=/v1/setup/levels
func (s *Service) SetLevelsAPI(ctx context.Context, req *SetLevelsRequest) error {
	if err := requireSession(ctx, req.SessionToken); err != nil {
		return err
	}

	defs, err := resolveLevelsRequest(req)
	if err != nil {
		return err
	}

	if len(defs) == 0 {
		return &errs.Error{Code: errs.InvalidArgument, Message: "at least one level required"}
	}

	if err := tenancy.ApplyLevels(ctx, defs); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not apply levels"}
	}

	payload, _ := json.Marshal(defs)
	if err := markStepComplete(ctx, "levels", payload); err != nil {
		return &errs.Error{Code: errs.Internal, Message: "could not mark progress"}
	}

	return nil
}

func resolveLevelsRequest(req *SetLevelsRequest) ([]tenancy.LevelDef, error) {
	if req.PackCode != "" {
		pack, err := countries.Get(req.PackCode)
		if err != nil {
			if errors.Is(err, countries.ErrPackNotFound) {
				return nil, &errs.Error{Code: errs.NotFound, Message: "country pack not found"}
			}

			return nil, &errs.Error{Code: errs.Internal, Message: "could not load country pack"}
		}

		out := make([]tenancy.LevelDef, 0, len(pack.Levels))
		for _, l := range pack.Levels {
			out = append(out, tenancy.LevelDef{
				Code:        l.Code,
				Label:       l.Label,
				ParentLevel: l.Parent,
				Depth:       l.Depth,
				SortOrder:   l.Sort,
			})
		}

		return out, nil
	}

	out := make([]tenancy.LevelDef, 0, len(req.Levels))
	for _, l := range req.Levels {
		out = append(out, tenancy.LevelDef{
			Code:        l.Code,
			Label:       l.Label,
			ParentLevel: l.Parent,
			Depth:       l.Depth,
			SortOrder:   l.Sort,
		})
	}

	return out, nil
}
