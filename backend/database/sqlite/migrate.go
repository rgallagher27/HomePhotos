package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// Migrate applies all pending migrations to the database.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("create _migrations table: %w", err)
	}

	var current int
	if err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM _migrations").Scan(&current); err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, name := range upFiles {
		var version int
		if _, err := fmt.Sscanf(name, "%06d", &version); err != nil {
			return fmt.Errorf("parse version from %s: %w", name, err)
		}
		if version <= current {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", name, err)
		}

		if _, err := tx.Exec("INSERT INTO _migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}

		slog.Info("applied migration", "file", name, "version", version)
	}

	return nil
}

// OpenAndMigrate opens a SQLite database and applies all pending migrations.
func OpenAndMigrate(path string) (*sql.DB, error) {
	db, err := Open(path)
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}
