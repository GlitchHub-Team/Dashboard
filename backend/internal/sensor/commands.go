package sensor

import (
	"time"

	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type CreateSensorCommand struct {
	identity.Requester
	Name      string
	Interval  time.Duration
	Profile   SensorProfile
	GatewayId uuid.UUID
}

type DeleteSensorCommand struct {
	identity.Requester
	SensorId uuid.UUID
}

type GetSensorCommand struct {
	identity.Requester
	SensorId uuid.UUID
}

type GetSensorsByGatewayCommand struct {
	identity.Requester
	Page      int
	Limit     int
	GatewayId uuid.UUID
}

type GetSensorsByTenantCommand struct {
	identity.Requester
	Page     int
	Limit    int
	TenantId uuid.UUID
}

type InterruptSensorCommand struct {
	identity.Requester
	SensorId uuid.UUID
}

type ResumeSensorCommand struct {
	identity.Requester
	SensorId uuid.UUID
}
