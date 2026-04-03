package sqlite

import (
	"path/filepath"
	"testing"
)

func TestMigrateFreshDB(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"_migrations", "schema_info", "users", "photos", "tag_groups", "tags", "photo_tags"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("expected table %q to exist: %v", table, err)
		}
	}

	// Verify migration version
	var version int
	if err := db.QueryRow("SELECT MAX(version) FROM _migrations").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	if version != 4 {
		t.Errorf("expected version 4, got %d", version)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := OpenAndMigrate(dbPath)
	if err != nil {
		t.Fatalf("first OpenAndMigrate: %v", err)
	}

	// Run again — should be a no-op
	if err := Migrate(db); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if count != 4 {
		t.Errorf("expected 4 migration records, got %d", count)
	}
	db.Close()
}

func TestMigratePartial(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	// Simulate having already applied first 2 migrations manually
	db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	// Apply migration 1 content manually
	db.Exec(`CREATE TABLE IF NOT EXISTS schema_info (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	db.Exec("INSERT INTO _migrations (version) VALUES (1)")
	// Apply migration 2 content manually
	db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email TEXT,
		role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'viewer')),
		display_name TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME
	)`)
	db.Exec("INSERT INTO _migrations (version) VALUES (2)")

	// Now run Migrate — should only apply 3 and 4
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var version int
	if err := db.QueryRow("SELECT MAX(version) FROM _migrations").Scan(&version); err != nil {
		t.Fatalf("query version: %v", err)
	}
	if version != 4 {
		t.Errorf("expected version 4, got %d", version)
	}

	// Verify photos table exists (from migration 3)
	var name string
	if err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='photos'").Scan(&name); err != nil {
		t.Error("expected photos table to exist")
	}

	db.Close()
}
