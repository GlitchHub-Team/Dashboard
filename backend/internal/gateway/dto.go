package gateway

import (
	"backend/internal/infra/transport/http/dto"
)

// Request
type createGatewayDTO struct {
	dto.GatewayNameField
	dto.GatewayIntervalField
}

type getGatewayListDTO struct {
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
	PublicIdentifier *string       `json:"public_identifier"`
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

// Response

type gatewayListResponseDTO struct {
	dto.ListInfo
	Gateways []gatewayResponseDTO `json:"gateways"`
}

func NewGatewayResponseDTO(gateway Gateway) gatewayResponseDTO {
	response := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: tenantIDString(gateway.TenantId)},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit.Milliseconds(),
		PublicIdentifier: gateway.PublicIdentifier,
	}

	return response
}

func NewGatewayListResponseDTO(gatewayList []Gateway, total uint) gatewayListResponseDTO {
	var gatewayDtos []gatewayResponseDTO

	for _, gateway := range gatewayList {
		gatewayDtos = append(gatewayDtos, NewGatewayResponseDTO(gateway))
	}

	return gatewayListResponseDTO{
		Gateways: gatewayDtos,
		ListInfo: dto.ListInfo{
			Count: uint(len(gatewayList)),
			Total: total,
		},
	}
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
