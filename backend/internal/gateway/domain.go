package gateway

import (
	"github.com/google/uuid"
)

type GatewayStatus string

const (
	GATEWAY_STATUS_ACTIVE   GatewayStatus = "active"
	GATEWAY_STATUS_INACTIVE GatewayStatus = "inactive"
)


type Gateway struct {
	Id            uuid.UUID
	Name          string
	TenantId      *uuid.UUID
	// Sensors	map[uuid.UUID]sensor.Sensor
	Status        GatewayStatus
	IntervalLimit int64
	
}
