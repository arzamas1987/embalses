package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB with SQLite-specific helpers.
type DB struct {
	*sql.DB
}

// Open creates or opens an SQLite database at the given path.
func Open(ctx context.Context, path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("wal mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("foreign keys: %w", err)
	}

	s := &DB{DB: db}
	if err := s.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return s, nil
}

// Close releases the database connection.
func (s *DB) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
}

// SeedTestKey inserts the test API key if not present.
func SeedTestKey(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO api_keys (key_hash, name, tier, daily_quota, rate_limit_per_minute)
		VALUES ('test-key-123', 'Test Key', 'free', 1000, 120)
		ON CONFLICT(key_hash) DO NOTHING
	`)
	return err
}
