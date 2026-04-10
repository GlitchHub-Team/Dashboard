package gateway

import (
	"time"

	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type CreateGatewayCommand struct {
	identity.Requester
	Name     string
	Interval time.Duration
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
	Size int
}

type GetGatewaysByTenantCommand struct {
	TenantId uuid.UUID
	// Page     int
	// Size     int
}

type CommissionGatewayCommand struct {
	GatewayId       uuid.UUID
	TenantId        uuid.UUID
	CommissionToken string
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
