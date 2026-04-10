package gateway

import (
	"github.com/google/uuid"
)

type GatewayStatus string

const (
	GATEWAY_STATUS_ACTIVE         GatewayStatus = "active"
	GATEWAY_STATUS_INACTIVE       GatewayStatus = "inactive"
	GATEWAY_STATUS_COMMISSIONED   GatewayStatus = "commissioned"
	GATEWAY_STATUS_DECOMMISSIONED GatewayStatus = "decommissioned"
	GATEWAY_STATUS_INTERRUPTED    GatewayStatus = "interrupted"
)

type Gateway struct {
	Id       uuid.UUID
	Name     string
	TenantId *uuid.UUID
	// Sensors	map[uuid.UUID]sensor.Sensor
	Status           GatewayStatus
	IntervalLimit    int64
	SigningSecret    string
	PublicIdentifier string
}

func (g Gateway) IsZero() bool {
	return g == (Gateway{})
}

func (g Gateway) IsCommissioned() bool {
	return g.TenantId != nil
}

func (g *Gateway) GetId() uuid.UUID { return g.Id }

func (g *Gateway) BelongsToTenant(userTenantId uuid.UUID) bool {
	return g.TenantId != nil && *g.TenantId == userTenantId
}
