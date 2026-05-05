package imports

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/imports/dbimports"
)

var db = sqldb.NewDatabase("imports", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbimports.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbimports.New(stdDB)
}()
