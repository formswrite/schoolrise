package setup

import (
	"context"
	"database/sql"

	"encore.app/apps/setup/dbsetup"
)

func markStepComplete(ctx context.Context, stepCode string, payload []byte) error {
	if len(payload) == 0 {
		payload = []byte("{}")
	}

	return queries.UpsertSetupProgressComplete(ctx, dbsetup.UpsertSetupProgressCompleteParams{
		StepCode: stepCode,
		Payload:  payload,
	})
}

func markStepSkipped(ctx context.Context, stepCode string) error {
	return queries.UpsertSetupProgressSkip(ctx, stepCode)
}

func upsertSystemSetting(ctx context.Context, key string, value []byte) error {
	return queries.UpsertSystemSetting(ctx, dbsetup.UpsertSystemSettingParams{
		Key:   key,
		Value: value,
	})
}

func errSQLNoRows() error {
	return sql.ErrNoRows
}
