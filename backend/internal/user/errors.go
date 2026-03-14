package user

import (
	"errors"
)

var errUserAlreadyExists = errors.New("User already present")
var errInexistentUser = errors.New("User does not exist")