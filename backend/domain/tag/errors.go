package tag

import "errors"

var (
	ErrNotFound      = errors.New("tag not found")
	ErrDuplicateName = errors.New("tag name already exists")
	ErrGroupNotFound = errors.New("tag group not found")
)
