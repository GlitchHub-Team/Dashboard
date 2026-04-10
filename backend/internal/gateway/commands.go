package gateway

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type CreateGatewayCommand struct {
	identity.Requester
	TenantId         uuid.UUID
	Name             string
	Certificate      string
	PublicIdentifier string
	SigningSecret    string
}

type DeleteGatewayCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type GetGatewayByIdCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type GetGatewayListCommand struct {
	Page int
	Limit int
}

type GetGatewaysByTenantCommand struct {
	TenantId uuid.UUID
	Page     int
	Limit    int
	identity.Requester
}

type GetGatewayByTenantIDCommand struct {
	TenantId  uuid.UUID
	GatewayId uuid.UUID
	identity.Requester
}

type GetAllGatewaysCommand struct {
	identity.Requester
	Page  int
	Limit int
}

type CommissionGatewayCommand struct {
	GatewayId          uuid.UUID
	TenantId           uuid.UUID
	GatewayCertificate string
	identity.Requester
}

type DecommissionGatewayCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type InterruptGatewayCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type ResumeGatewayCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type ResetGatewayCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type RebootGatewayCommand struct {
	GatewayId uuid.UUID
	identity.Requester
}

type SetGatewayIntervalLimitCommand struct {
	GatewayId     uuid.UUID
	IntervalLimit int
	identity.Requester
}
