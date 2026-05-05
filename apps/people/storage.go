package people

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/people/dbpeople"
)

var db = sqldb.NewDatabase("people", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbpeople.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbpeople.New(stdDB)
}()
