package historical_data

import (
	"time"

	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type GetSensorHistoricalDataCommand struct {
	Requester identity.Requester
	TenantId  uuid.UUID
	SensorId  uuid.UUID
	From      *time.Time
	To        *time.Time
	Limit     int
}
