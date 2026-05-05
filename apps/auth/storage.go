package auth

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/auth/dbauth"
)

var db = sqldb.NewDatabase("auth", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbauth.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbauth.New(stdDB)
}()
