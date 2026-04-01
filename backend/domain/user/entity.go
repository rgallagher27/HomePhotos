package user

import "time"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Email        string
	Role         Role
	DisplayName  string
	CreatedAt    time.Time
	LastLogin    *time.Time
}
