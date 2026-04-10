package gateway

import (
	"time"

	"github.com/google/uuid"
)

type GatewayStatus string

const (
	GATEWAY_STATUS_ACTIVE         GatewayStatus = "active"
	GATEWAY_STATUS_INACTIVE       GatewayStatus = "inactive"
	GATEWAY_STATUS_DECOMMISSIONED GatewayStatus = "decommissioned"
)

type Gateway struct {
	Id               uuid.UUID
	Name             string
	TenantId         *uuid.UUID
	Status           GatewayStatus
	IntervalLimit    time.Duration
	PublicIdentifier *string
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
