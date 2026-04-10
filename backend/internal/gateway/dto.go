package gateway

import (
	"backend/internal/infra/transport/http/dto"
)

// Request
type createGatewayDTO struct {
	dto.GatewayNameField
	dto.GatewayIntervalField
}

type getGatewayByIdDTO struct {
	dto.GatewayIdField
}

type getGatewaysByTenantDTO struct {
	dto.TenantIdField
	dto.Pagination
}

type commissionGatewayDTO struct {
	dto.TenantIdField
	dto.CommissionTokenField
}

// Response
type gatewayResponseDTO struct {
	dto.GatewayIdField
	dto.GatewayNameField
	dto.TenantIdField
	Status           GatewayStatus `json:"status"`
	Interval         int64         `json:"interval"`
	PublicIdentifier *string       `json:"publicIdentifier"`
}

type gatewayCommandResponseDTO struct {
	Result string `json:"result"`
}

type createGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
	Interval  int64  `json:"interval"`
}

type deleteGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
}

type commissionGatewayCommandPayloadDTO struct {
	GatewayId       string `json:"gatewayId"`
	TenantId        string `json:"tenantId"`
	CommissionToken string `json:"commissionedToken"`
}

type decommissionGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
}

type interruptGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
}

type resumeGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
}

type resetGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
}

type rebootGatewayCommandPayloadDTO struct {
	GatewayId string `json:"gatewayId"`
}
