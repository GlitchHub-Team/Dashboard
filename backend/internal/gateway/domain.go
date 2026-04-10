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
	DEFAULT_INTERVAL_LIMIT        time.Duration = time.Second * 5

	CREATE_GATEWAY_COMMAND_SUBJECT       = "commands.creategateway"
	DELETE_GATEWAY_COMMAND_SUBJECT       = "commands.deletegateway"
	COMMISSION_GATEWAY_COMMAND_SUBJECT   = "commands.commissiongateway"
	DECOMMISSION_GATEWAY_COMMAND_SUBJECT = "commands.decommissiongateway"
	INTERRUPT_GATEWAY_COMMAND_SUBJECT    = "commands.interruptgateway"
	RESUME_GATEWAY_COMMAND_SUBJECT       = "commands.resumegateway"
	RESET_GATEWAY_COMMAND_SUBJECT        = "commands.resetgateway"
	REBOOT_GATEWAY_COMMAND_SUBJECT       = "commands.rebootgateway"
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
