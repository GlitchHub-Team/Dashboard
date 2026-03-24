package user

import (
	"errors"
)

var (
	ErrUserAlreadyExists = errors.New("user already present")
	ErrUserNotFound      = errors.New("user not found")
	ErrUnknownRole       = errors.New("unknown role")
	ErrCannotSendEmail   = errors.New("cannot create user: cannot send confirmation email")
	ErrEmptyPassword     = errors.New("cannot set empty password")
	ErrSamePassword      = errors.New("cannot set new password equal to old one")
	ErrInvalidUser       = errors.New("cannot get user with ID 0")
)
