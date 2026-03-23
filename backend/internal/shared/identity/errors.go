package identity

import "errors"

var ErrUnauthorizedAccess = errors.New("cannot access resource")
var ErrUnknownRole = errors.New("unknown role")