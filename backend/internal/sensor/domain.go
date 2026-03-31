package sensor

import (
	"time"

	"github.com/google/uuid"
)

type SensorProfile string

const (
	HEART_RATE            SensorProfile = "heart_rate"
	PULSE_OXIMETER        SensorProfile = "pulse_oximeter"
	ECG_CUSTOM            SensorProfile = "ecg_custom"
	HEALTH_THERMOMETER    SensorProfile = "health_thermometer"
	ENVIRONMENTAL_SENSING SensorProfile = "environmental_sensing"
)

type SensorStatus string

const (
	Active   SensorStatus = "active"
	Inactive SensorStatus = "inactive"
)

type Sensor struct {
	Id        uuid.UUID
	Name      string
	Interval  time.Duration
	Profile   SensorProfile
	GatewayId uuid.UUID
	Status    SensorStatus
}

func (s Sensor) IsZero() bool {
	return s == (Sensor{})
}
