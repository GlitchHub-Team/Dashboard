package real_time_data

import (
	sensorProfile "backend/internal/sensor/profile"
	"encoding/json"
	"time"
)

type RealTimeRawSample struct {
	Profile   sensorProfile.SensorProfile
	Timestamp time.Time
	Data      json.RawMessage
}

type RealTimeError struct {
	err error
	Timestamp time.Time
}

func NewErrClientDisconnected(timestamp time.Time) RealTimeError {
	return RealTimeError{
		err: ErrClientDisconnected,
		Timestamp: timestamp,
	}
}

func (e RealTimeError) Error() string {
	return e.err.Error()
}

func (e RealTimeError) Unwrap() error {
	return e.err
}