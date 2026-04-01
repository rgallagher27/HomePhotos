package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrRegistrationClosed  = errors.New("registration is closed")
)
