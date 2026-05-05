package ai

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/ai/dbai"
)

var db = sqldb.NewDatabase("ai", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbai.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbai.New(stdDB)
}()
