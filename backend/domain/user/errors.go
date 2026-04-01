package user

import "errors"

var (
	ErrNotFound          = errors.New("user not found")
	ErrDuplicateUsername = errors.New("username already exists")
)
