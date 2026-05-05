package assessment

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/assessment/dbassessment"
)

var db = sqldb.NewDatabase("assessment", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbassessment.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbassessment.New(stdDB)
}()
