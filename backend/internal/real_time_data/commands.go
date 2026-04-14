package real_time_data

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type GetRealTimeDataCommand struct {
	Requester identity.Requester
	SensorId  uuid.UUID
	TenantId  uuid.UUID
}
