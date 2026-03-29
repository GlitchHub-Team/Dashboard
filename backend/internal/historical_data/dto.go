package historical_data

import "time"

type GetHistoricalDataQueryDTO struct {
	From  *string `form:"from" binding:"omitempty"`
	To    *string `form:"to" binding:"omitempty"`
	Limit int     `form:"limit" binding:"omitempty,min=1,max=5000"`
}

type HistoricalSampleResponseDTO struct {
	SensorId  string    `json:"sensor_id" binding:"required,uuid4"`
	GatewayId string    `json:"gateway_id" binding:"required,uuid4"`
	TenantId  string    `json:"tenant_id" binding:"required,uuid4"`
	Profile   string    `json:"profile" binding:"required"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
	Data      any       `json:"data" binding:"required"`
}

type HistoricalDataResponseDTO struct {
	Count   uint                          `json:"count" binding:"required"`
	Samples []HistoricalSampleResponseDTO `json:"samples" binding:"required,min=0"`
}

func NewHistoricalDataResponseDTO(samples []HistoricalSample) HistoricalDataResponseDTO {
	responseSamples := make([]HistoricalSampleResponseDTO, 0, len(samples))
	for _, sample := range samples {
		responseSamples = append(responseSamples, HistoricalSampleResponseDTO{
			SensorId:  sample.SensorId.String(),
			GatewayId: sample.GatewayId.String(),
			TenantId:  sample.TenantId.String(),
			Profile:   sample.Profile,
			Timestamp: sample.Timestamp,
			Data:      sample.Data,
		})
	}

	return HistoricalDataResponseDTO{
		Count:   uint(len(responseSamples)),
		Samples: responseSamples,
	}
}
