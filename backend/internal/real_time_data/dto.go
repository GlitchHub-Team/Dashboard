package real_time_data

import "backend/internal/infra/transport/http/dto"

// Request ------------------------------------------------------------------------------------------------
type GetRealTimeDataDTO struct {
	dto.SensorIdField
	dto.TenantIdField
}

// Output (verso client) ----------------------------------------------------------------------------------
type RealTimeErrorOutDTO struct {
	Error string `json:"error"`
}