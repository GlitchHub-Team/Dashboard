package gateway

import (
	"backend/internal/infra/transport/http/dto"
)

// Request
type createGatewayDTO struct {
	dto.GatewayNameField
	dto.GatewayIntervalField
}

type deleteGatewayDTO struct {
	dto.GatewayIdField
}

type getGatewayByIdDTO struct {
	dto.GatewayIdField
}

type getGatewayListDTO struct {
	dto.Pagination
}

type getGatewaysByTenantDTO struct {
	dto.TenantIdField
	Page  int
	Limit int
}

type commissionGatewayDTO struct {
	dto.TenantIdField
	dto.CommissionTokenField
}

type interruptGatewayDTO struct {
	dto.GatewayIdField
}

type resumeGatewayDTO struct {
	dto.GatewayIdField
}

type resetGatewayDTO struct {
	dto.GatewayIdField
}

type rebootGatewayDTO struct {
	dto.GatewayIdField
}

// Response
type gatewayResponseDTO struct {
	dto.GatewayIdField
	dto.GatewayNameField
	dto.TenantIdField
	Status   GatewayStatus `json:"status"`
	Interval int64         `json:"interval"`
	PublicIdentifier *string       `json:"publicIdentifier"`
}
type gatewayListResponseDTO struct {
	dto.ListInfo
	Gateways []gatewayResponseDTO `json:"gateways"`
}

type commissionGatewayResponseDTO struct {
	dto.TenantIdField
	dto.TenantNameField
}

type gatewayCommandResponseDTO struct {
	Result string `json:"result"`
}

type getGatewayBodyDTO struct {
	Page  int
	Limit int
}
