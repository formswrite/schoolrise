package enrollment

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/enrollment/dbenrollment"
)

var db = sqldb.NewDatabase("enrollment", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbenrollment.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbenrollment.New(stdDB)
}()
