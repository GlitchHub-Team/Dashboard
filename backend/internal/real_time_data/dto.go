package real_time_data

import (
	httpDto "backend/internal/infra/transport/http/dto"
	natsDto "backend/internal/infra/transport/nats/dto"
)

// Request ------------------------------------------------------------------------------------------------
type GetRealTimeDataDTO struct {
	httpDto.SensorIdField
	httpDto.TenantIdField
}

// Output (verso client) ----------------------------------------------------------------------------------
type RealTimeErrorOutDTO struct {
	Error string `json:"error"`
}

type RealTimeSampleOutDTO struct {
	natsDto.ProfileField
	natsDto.TimestampField
	Data any `json:"data" binding:"required"`
}
