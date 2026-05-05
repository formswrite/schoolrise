package notifications

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/notifications/dbnotifications"
)

var db = sqldb.NewDatabase("notifications", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbnotifications.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbnotifications.New(stdDB)
}()
