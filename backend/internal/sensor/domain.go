package sensor

import (
	"time"
	profile "backend/internal/sensor/profile"
	"github.com/google/uuid"
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
	Profile   profile.SensorProfile
	GatewayId uuid.UUID
	Status    SensorStatus
}

func (s Sensor) IsZero() bool {
	return s == (Sensor{})
}
