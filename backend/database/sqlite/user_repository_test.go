package sqlite_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/domain/user"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)

	entries, err := os.ReadDir("migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		data, err := os.ReadFile(filepath.Join("migrations", f))
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			t.Fatalf("exec %s: %v", f, err)
		}
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	u := &user.User{
		Username:     "alice",
		PasswordHash: "hashed",
		Email:        "alice@example.com",
		Role:         user.RoleAdmin,
		DisplayName:  "Alice",
	}

	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if created.Username != "alice" {
		t.Errorf("username = %q, want %q", created.Username, "alice")
	}
	if created.Role != user.RoleAdmin {
		t.Errorf("role = %q, want %q", created.Role, user.RoleAdmin)
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestUserRepository_CreateDuplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	u := &user.User{
		Username:     "alice",
		PasswordHash: "hashed",
		Role:         user.RoleViewer,
	}
	if _, err := repo.Create(ctx, u); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create(ctx, u)
	if err != user.ErrDuplicateUsername {
		t.Errorf("err = %v, want ErrDuplicateUsername", err)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	// Not found
	_, err := repo.GetByUsername(ctx, "nobody")
	if err != user.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}

	// Found
	u := &user.User{Username: "bob", PasswordHash: "hashed", Role: user.RoleViewer}
	if _, err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByUsername(ctx, "bob")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Username != "bob" {
		t.Errorf("username = %q, want %q", got.Username, "bob")
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	// Not found
	_, err := repo.GetByID(ctx, 999)
	if err != user.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}

	// Found
	u := &user.User{Username: "charlie", PasswordHash: "hashed", Role: user.RoleViewer}
	created, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Username != "charlie" {
		t.Errorf("username = %q, want %q", got.Username, "charlie")
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	// Empty
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("len = %d, want 0", len(users))
	}

	// Two users
	repo.Create(ctx, &user.User{Username: "a", PasswordHash: "h", Role: user.RoleAdmin})
	repo.Create(ctx, &user.User{Username: "b", PasswordHash: "h", Role: user.RoleViewer})

	users, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len = %d, want 2", len(users))
	}
}

func TestUserRepository_UpdateRole(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	created, _ := repo.Create(ctx, &user.User{Username: "dave", PasswordHash: "h", Role: user.RoleViewer})

	if err := repo.UpdateRole(ctx, created.ID, user.RoleAdmin); err != nil {
		t.Fatalf("update role: %v", err)
	}

	got, _ := repo.GetByID(ctx, created.ID)
	if got.Role != user.RoleAdmin {
		t.Errorf("role = %q, want %q", got.Role, user.RoleAdmin)
	}
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	created, _ := repo.Create(ctx, &user.User{Username: "eve", PasswordHash: "h", Role: user.RoleViewer})
	if created.LastLogin != nil {
		t.Error("expected nil last_login initially")
	}

	if err := repo.UpdateLastLogin(ctx, created.ID); err != nil {
		t.Fatalf("update last login: %v", err)
	}

	got, _ := repo.GetByID(ctx, created.ID)
	if got.LastLogin == nil {
		t.Error("expected non-nil last_login after update")
	}
}

func TestUserRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewUserRepository(db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	repo.Create(ctx, &user.User{Username: "f", PasswordHash: "h", Role: user.RoleViewer})
	repo.Create(ctx, &user.User{Username: "g", PasswordHash: "h", Role: user.RoleViewer})

	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}
