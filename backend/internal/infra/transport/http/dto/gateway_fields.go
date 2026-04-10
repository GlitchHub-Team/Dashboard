package dto

type GatewayIdField struct {
	GatewayId string `uri:"gateway_id" form:"gateway_id" json:"gateway_id" binding:"uuid4,required"`
}

type GatewayNameField struct {
	GatewayName string `uri:"gateway_name" form:"gateway_name" json:"gateway_name" binding:"required"`
}

type CommissionTokenField struct {
	CommissionToken string `uri:"commission_token" form:"commission_token" json:"commission_token" binding:"required"`
}

type GatewayIntervalField struct {
	Interval int64 `uri:"interval" form:"interval" json:"interval" binding:"required,gt=0"`
}
