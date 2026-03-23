package auth

import "errors"

var ErrAccountNotConfirmed = errors.New("account not confirmed")
var ErrWrongCredentials = errors.New("wrong credentials")
var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")
var ErrAccountAlreadyConfirmed = errors.New("account already confirmed")