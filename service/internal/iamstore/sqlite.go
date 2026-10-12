package iamstore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	iamsqlite "github.com/jamesread/armature-iam/store/sqlite"
	// Pure-Go SQLite so release builds with CGO_ENABLED=0 (goreleaser) can open IAM.
	_ "modernc.org/sqlite"
)

// SQLite persists IAM data (users, sessions, API keys, RBAC) separately from task storage.
type SQLite struct {
	*iamsqlite.SQLite
	db *sql.DB
}

// Open opens (or creates) a SQLite database at dbPath.
func Open(dbPath string) (*SQLite, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	dsn := "file:" + filepath.ToSlash(dbPath) + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	configureSQLitePool(db)
	if err := applySQLitePragmas(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLite{db: db, SQLite: iamsqlite.New(db)}, nil
}

func configureSQLitePool(db *sql.DB) {
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(-1)
}

func applySQLitePragmas(db *sql.DB) error {
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return err
	}
	if _, err := db.Exec(`PRAGMA synchronous=NORMAL`); err != nil {
		return err
	}
	return db.Ping()
}

// OpenMemory opens an in-memory IAM database with schema applied (for tests).
func OpenMemory() (*SQLite, error) {
	db, err := sql.Open("sqlite", "file::memory:?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := iamsqlite.ApplySchema(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLite{db: db, SQLite: iamsqlite.New(db)}, nil
}

// SQLDB exposes the underlying connection for migrations.
func (s *SQLite) SQLDB() *sql.DB {
	return s.db
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLite) HasMigration(ctx context.Context, id string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM migrations WHERE id = ?`, id).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("query migrations: %w", err)
	}
	return n > 0, nil
}

func (s *SQLite) LatestMigration(ctx context.Context) (string, error) {
	var id sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM migrations ORDER BY applied_at DESC, id DESC LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("latest migration: %w", err)
	}
	if !id.Valid {
		return "", nil
	}
	return id.String, nil
}
