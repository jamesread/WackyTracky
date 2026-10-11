package migrate

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSQLiteMigrationsDir locates database/sqlite/migrations for in-process migrate.
func FindSQLiteMigrationsDir() (string, error) {
	for _, candidate := range sqliteMigrationCandidates() {
		if dir := usableMigrationsDir(candidate); dir != "" {
			return dir, nil
		}
	}
	return "", fmt.Errorf("sqlite migrations directory not found (set WACKYTRACKY_MIGRATIONS_DIR or run from the repo root)")
}

func sqliteMigrationCandidates() []string {
	candidates := []string{}
	if env := os.Getenv("WACKYTRACKY_MIGRATIONS_DIR"); env != "" {
		candidates = append(candidates, env)
	}
	return append(candidates,
		"database/sqlite/migrations",
		filepath.Join("..", "database", "sqlite", "migrations"),
		filepath.Join("..", "..", "database", "sqlite", "migrations"),
		"/var/app/database/sqlite/migrations",
		"/app/database/sqlite/migrations",
	)
}

func usableMigrationsDir(candidate string) string {
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return ""
	}
	if !isDir(abs) {
		return ""
	}
	if !dirHasSQLFiles(abs) {
		return ""
	}
	return abs
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func dirHasSQLFiles(abs string) bool {
	entries, err := os.ReadDir(abs)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			return true
		}
	}
	return false
}
