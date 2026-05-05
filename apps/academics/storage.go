package academics

import (
	"context"
	"database/sql"

	"encore.dev/storage/sqldb"

	"encore.app/apps/academics/dbacademics"
)

var db = sqldb.NewDatabase("academics", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = func() *dbacademics.Queries {
	var n int
	_ = db.QueryRow(context.Background(), "SELECT 1").Scan(&n)

	connStr := sqldb.RegisterStdlibDriver(db)

	stdDB, err := sql.Open("encore", connStr)
	if err != nil {
		panic(err)
	}

	return dbacademics.New(stdDB)
}()
