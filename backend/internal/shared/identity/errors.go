package identity

import "errors"

var (
	ErrUnauthorizedAccess = errors.New("cannot access resource")
	ErrUnknownRole        = errors.New("unknown role")
	ErrInvalidUser        = errors.New("invalid user")
)
