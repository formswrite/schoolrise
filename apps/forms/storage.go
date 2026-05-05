package forms

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/forms/dbforms"
)

var db = sqldb.NewDatabase("forms", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbforms.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbforms.New(stdDB)
}()
