package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
var initSQL string

func Open(dbPath string) (*sql.DB, error) {
	if dbPath != ":memory:" && !isMemoryURI(dbPath) {
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create db dir: %w", err)
			}
		}
	}

	dsn := dbPath
	if dbPath != ":memory:" && !isMemoryURI(dbPath) {
		dsn = fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)
	}

	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if _, err := database.Exec("PRAGMA busy_timeout = 5000; PRAGMA foreign_keys = ON;"); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("configure sqlite pragmas: %w", err)
	}
	if _, err := database.Exec(initSQL); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("run init migration: %w", err)
	}
	return database, nil
}

func isMemoryURI(p string) bool {
	return len(p) > 7 && p[:5] == "file:" && (p == "file::memory:" || p == "file::memory:?cache=shared")
}
