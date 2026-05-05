package setup

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/setup/dbsetup"
)

var db = sqldb.NewDatabase("setup", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbsetup.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbsetup.New(stdDB)
}()
