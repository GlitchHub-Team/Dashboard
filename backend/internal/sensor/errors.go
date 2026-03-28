package sensor

import "errors"

var (
	ErrSensorNotFound      = errors.New("sensor not found")
	ErrSensorAlreadyExists = errors.New("sensor already exists")
	ErrInvalidSensorID     = errors.New("invalid sensor ID")
)
