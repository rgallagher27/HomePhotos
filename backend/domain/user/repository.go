package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context) ([]User, error)
	UpdateRole(ctx context.Context, id int64, role Role) error
	UpdateLastLogin(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
}
