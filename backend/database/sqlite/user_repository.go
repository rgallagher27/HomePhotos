package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rgallagher/homephotos/domain/user"
)

type UserRepository struct {
	q *Queries
}

func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{q: New(db)}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	row, err := r.q.CreateUser(ctx, CreateUserParams{
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Email:        toNullString(u.Email),
		Role:         string(u.Role),
		DisplayName:  toNullString(u.DisplayName),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, user.ErrDuplicateUsername
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return rowToUser(row), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return rowToUser(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return rowToUser(row), nil
}

func (r *UserRepository) List(ctx context.Context) ([]user.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]user.User, len(rows))
	for i, row := range rows {
		users[i] = user.User{
			ID:          row.ID,
			Username:    row.Username,
			Email:       fromNullString(row.Email),
			Role:        user.Role(row.Role),
			DisplayName: fromNullString(row.DisplayName),
			CreatedAt:   row.CreatedAt,
			LastLogin:   fromNullTime(row.LastLogin),
		}
	}
	return users, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role user.Role) error {
	return r.q.UpdateUserRole(ctx, UpdateUserRoleParams{
		Role: string(role),
		ID:   id,
	})
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
	return r.q.UpdateLastLogin(ctx, id)
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}

func rowToUser(row User) *user.User {
	return &user.User{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Email:        fromNullString(row.Email),
		Role:         user.Role(row.Role),
		DisplayName:  fromNullString(row.DisplayName),
		CreatedAt:    row.CreatedAt,
		LastLogin:    fromNullTime(row.LastLogin),
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
