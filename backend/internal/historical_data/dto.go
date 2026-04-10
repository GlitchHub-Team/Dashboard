package historical_data

import transportDto "backend/internal/infra/transport/http/dto"

type GetHistoricalDataQueryDTO struct {
	From  *string `form:"from" binding:"omitempty"`
	To    *string `form:"to" binding:"omitempty"`
	Limit int     `form:"limit" binding:"omitempty,min=1,max=5000"`
}

type HistoricalSampleResponseDTO struct {
	transportDto.SensorIdField
	transportDto.GatewayIdField
	transportDto.TenantIdField
	transportDto.TimestampField
	Profile string `json:"profile" binding:"required"`
	Data    any    `json:"data" binding:"required"`
}

type HistoricalDataResponseDTO struct {
	Count   uint                          `json:"count" binding:"required"`
	Samples []HistoricalSampleResponseDTO `json:"samples" binding:"required,min=0"`
}

func NewHistoricalDataResponseDTO(samples []HistoricalSample) (HistoricalDataResponseDTO, error) {
	responseSamples := make([]HistoricalSampleResponseDTO, 0, len(samples))
	for _, sample := range samples {
		decodedData, err := transportDto.DecodeSensorProfileData(sample.Profile, sample.Data)
		if err != nil {
			return HistoricalDataResponseDTO{}, err
		}

		responseSamples = append(responseSamples, HistoricalSampleResponseDTO{
			SensorIdField: transportDto.SensorIdField{
				SensorId: sample.SensorId.String(),
			},
			GatewayIdField: transportDto.GatewayIdField{
				GatewayId: sample.GatewayId.String(),
			},
			TenantIdField: transportDto.TenantIdField{
				TenantId: sample.TenantId.String(),
			},
			TimestampField: transportDto.TimestampField{
				Timestamp: sample.Timestamp,
			},
			Profile: string(sample.Profile),
			Data:    decodedData,
		})
	}

	return HistoricalDataResponseDTO{
		Count:   uint(len(responseSamples)),
		Samples: responseSamples,
	}, nil
}
