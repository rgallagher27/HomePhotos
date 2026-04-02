package photo

import "errors"

var (
	ErrNotFound      = errors.New("photo not found")
	ErrDuplicatePath = errors.New("file path already exists")
)
