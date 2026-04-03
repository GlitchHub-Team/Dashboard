package dto

type GatewayIdField struct {
	GatewayId string `uri:"gateway_id" form:"gateway_id" json:"gateway_id" binding:"uuid4,required"`
}

type GatewayNameField struct {
	GatewayName string `uri:"gateway_name" form:"gateway_name" json:"gateway_name" binding:"required"`
}

type GatewayCertificateField struct {
	Certificate string `uri:"certificate" form:"certificate" json:"certificate" binding:"required"`
}
