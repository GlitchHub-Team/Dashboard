package real_time_data

import (
	"errors"
)

var (
	ErrClientDisconnected = errors.New("client disconnected")
	ErrMappingError       = errors.New("mapping error")
)
