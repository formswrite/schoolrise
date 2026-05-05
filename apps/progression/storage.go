package progression

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/progression/dbprogression"
)

var db = sqldb.NewDatabase("progression", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbprogression.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbprogression.New(stdDB)
}()
