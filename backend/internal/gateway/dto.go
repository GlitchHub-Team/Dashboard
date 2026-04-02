package gateway

import (
	"backend/internal/infra/transport/http/dto"
)

// Request
type createGatewayDTO struct {
	dto.GatewayNameField
}

type deleteGatewayDTO struct {
	dto.GatewayIdField
}

/*
	type getGatewayByIdDTO struct {
		dto.GatewayIdField
	}

	type getGatewayListDTO struct {
		dto.Pagination
	}
*/
type getGatewaysByTenantDTO struct {
	dto.TenantIdField
	dto.Pagination
}

/*
	type commissionGatewayDTO struct {
		dto.TenantIdField
		dto.GatewayIdField
		dto.GatewayCertificateField
	}

	type decommissionGatewayDTO struct {
		dto.GatewayIdField
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

type setGatewayIntervalLimitDTO struct {
	dto.GatewayIdField
	IntervalLimit int `uri:"interval_limit" form:"interval_limit" json:"interval_limit" binding:"required"`
}
*/
// Response
type gatewayResponseDTO struct {
	dto.GatewayIdField
	dto.GatewayNameField
}

/*
type gatewayListResponseDTO struct {
	dto.ListInfo
	Gateways []gatewayResponseDTO `json:"gateways"`
}

type commissionGatewayResponseDTO struct {
	dto.TenantIdField
	dto.TenantNameField
	dto.GatewayCertificateField
}
*/
