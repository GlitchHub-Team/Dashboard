package user

import (
	"errors"
)

var (
	ErrUserAlreadyExists = errors.New("user already present")
	ErrUserNotFound      = errors.New("user not found")
	ErrUnknownRole       = errors.New("unknown role")
	ErrCannotSendEmail   = errors.New("cannot create user: cannot send confirmation email")
)
