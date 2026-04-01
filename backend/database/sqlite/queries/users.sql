-- name: CreateUser :one
INSERT INTO users (username, password_hash, email, role, display_name)
VALUES (?, ?, ?, ?, ?)
RETURNING id, username, password_hash, email, role, display_name, created_at, last_login;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, email, role, display_name, created_at, last_login
FROM users WHERE username = ?;

-- name: GetUserByID :one
SELECT id, username, password_hash, email, role, display_name, created_at, last_login
FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, username, email, role, display_name, created_at, last_login
FROM users ORDER BY created_at;

-- name: UpdateUserRole :exec
UPDATE users SET role = ? WHERE id = ?;

-- name: UpdateLastLogin :exec
UPDATE users SET last_login = datetime('now') WHERE id = ?;

-- name: CountUsers :one
SELECT count(*) FROM users;
