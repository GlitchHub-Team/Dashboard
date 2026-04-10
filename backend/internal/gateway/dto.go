package gateway

import (
	"backend/internal/infra/transport/http/dto"
)

// Request
type createGatewayDTO struct {
	dto.GatewayNameField
	dto.GatewayIntervalField
<<<<<<< issue-130
}

type deleteGatewayDTO struct {
	dto.GatewayIdField
=======
>>>>>>> main
}

type getGatewayByIdDTO struct {
	dto.GatewayIdField
}

<<<<<<< issue-130
type getGatewayListDTO struct {
	dto.Pagination
}

=======
>>>>>>> main
type getGatewaysByTenantDTO struct {
	dto.TenantIdField
	Page  int
	Limit int
}

type commissionGatewayDTO struct {
	dto.TenantIdField
	dto.CommissionTokenField
<<<<<<< issue-130
}

type interruptGatewayDTO struct {
	dto.GatewayIdField
=======
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
>>>>>>> main
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

<<<<<<< issue-130
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
=======
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
>>>>>>> main
