package tenancy

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/tenancy/dbtenancy"
)

var db = sqldb.NewDatabase("tenancy", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbtenancy.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbtenancy.New(stdDB)
}()
