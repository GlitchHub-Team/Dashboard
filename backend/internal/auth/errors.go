package auth

import "errors"

var (
	ErrAccountNotConfirmed     = errors.New("account not confirmed")
	ErrWrongCredentials        = errors.New("wrong credentials")
	ErrTokenNotFound           = errors.New("token not found")
	ErrTokenExpired            = errors.New("token expired")
	ErrAccountAlreadyConfirmed = errors.New("account already confirmed")
)
