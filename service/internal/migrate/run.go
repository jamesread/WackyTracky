package migrate

import (
	"fmt"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
)

// RunSQLiteUp applies pending sql-migrate scripts for the IAM SQLite database.
func RunSQLiteUp(st *iamstore.SQLite, migrationsDir string) (int, error) {
	if st == nil {
		return 0, fmt.Errorf("nil iam store")
	}
	source := &migrate.FileMigrationSource{Dir: migrationsDir}
	ms := migrate.MigrationSet{TableName: "migrations"}
	return ms.Exec(st.SQLDB(), "sqlite3", source, migrate.Up)
}
