package sensor

import (
	"time"

	profile "backend/internal/sensor/profile"

	"backend/internal/gateway"
	"github.com/google/uuid"
)

func DomainToSensorEntity(s Sensor) *SensorEntity {
	return &SensorEntity{
		ID:        s.Id.String(),
		GatewayID: s.GatewayId.String(),
		Name:      s.Name,
		Interval:  s.Interval.Milliseconds(),
		Profile:   string(s.Profile),
		Status:    string(s.Status),
	}
}

func SensorEntityToDomain(entity *SensorEntity) (sensor Sensor, err error) {
	if entity == nil {
		return
	}

	sensorId, err := uuid.Parse(entity.ID)
	if err != nil {
		err = ErrSensorNotFound
		return
	}

	gatewayId, err := uuid.Parse(entity.GatewayID)
	if err != nil {
		err = gateway.ErrGatewayNotFound
		return
	}

	return Sensor{
		Id:        sensorId,
		Name:      entity.Name,
		Interval:  time.Duration(entity.Interval) * time.Millisecond,
		Profile:   profile.SensorProfile(entity.Profile),
		GatewayId: gatewayId,
		Status:    SensorStatus(entity.Status),
	}, nil
}
