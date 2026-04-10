package sensor

import (
	"backend/internal/infra/transport/http/dto"
	profile "backend/internal/sensor/profile"

	"github.com/google/uuid"
)

type CreateSensorBodyDTO struct {
	Name      string                `json:"sensor_name" binding:"required"`
	Interval  int64                 `json:"data_interval" binding:"required,gte=1"`
	Profile   profile.SensorProfile `json:"profile" binding:"required,oneof=heart_rate pulse_oximeter ecg_custom health_thermometer environmental_sensing"`
	GatewayId uuid.UUID             `json:"gateway_id" binding:"required,uuid"`
}

type SensorResponseDTO struct {
	SensorId  uuid.UUID `json:"sensor_id"`
	GatewayId uuid.UUID `json:"gateway_id"`
	Name      string    `json:"sensor_name"`
	Interval  int64     `json:"data_interval" binding:"required,gte=1"`
	Status    string    `json:"status" binding:"required,oneof=active inactive"`
	Profile   string    `json:"profile"`
}

func NewSensorResponseDTO(sensor Sensor) SensorResponseDTO {
	response := SensorResponseDTO{
		SensorId:  sensor.Id,
		GatewayId: sensor.GatewayId,
		Name:      sensor.Name,
		Interval:  sensor.Interval.Milliseconds(),
		Profile:   string(sensor.Profile),
		Status:    string(sensor.Status),
	}
	return response
}

type SensorQueryDTO struct {
	dto.Pagination
}

type SensorsResponseDTO struct {
	Sensors []SensorResponseDTO `json:"sensors"`
	dto.ListInfo
}

func NewSensorsResponseDTO(sensors []Sensor, totalCount uint) SensorsResponseDTO {
	responseDtos := make([]SensorResponseDTO, len(sensors))
	for i, sensor := range sensors {
		responseDtos[i] = NewSensorResponseDTO(sensor)
	}

	paginationInfo := dto.ListInfo{
		Count: uint(len(sensors)),
		Total: totalCount,
	}

	return SensorsResponseDTO{
		Sensors:  responseDtos,
		ListInfo: paginationInfo,
	}
}
