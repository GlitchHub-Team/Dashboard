package real_time_data

import (
	"errors"
)

var ErrClientDisconnected = errors.New("client disconnected")
var ErrMappingError = errors.New("mapping error")