package profile

import "errors"

type SensorProfile string

const (
	ECG_CUSTOM            SensorProfile = "ecg_custom"
	ENVIRONMENTAL_SENSING SensorProfile = "environmental_sensing"
	HEALTH_THERMOMETER    SensorProfile = "health_thermometer"
	HEART_RATE            SensorProfile = "heart_rate"
	PULSE_OXIMETER        SensorProfile = "pulse_oximeter"
)

var ErrUnknownProfile = errors.New("unknown profile")