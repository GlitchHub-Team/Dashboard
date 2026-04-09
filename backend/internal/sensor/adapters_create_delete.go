package sensor

import (
	"time"
	profile "backend/internal/sensor/profile"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DbSensorAdapter struct {
	log  *zap.Logger
	repo DatabaseRepository
}

type SendCmdAdapter struct {
	log  *zap.Logger
	repo MessageBrokerRepository
}

func NewDbSensorAdapter(log *zap.Logger, repository DatabaseRepository) *DbSensorAdapter {
	return &DbSensorAdapter{
		log:  log,
		repo: repository,
	}
}

func NewSendCmdAdapter(log *zap.Logger, repository MessageBrokerRepository) *SendCmdAdapter {
	return &SendCmdAdapter{
		log:  log,
		repo: repository,
	}
}

// Compile-time checks
var (
	_ CreateSensorPort    = (*DbSensorAdapter)(nil)
	_ DeleteSensorPort    = (*DbSensorAdapter)(nil)
	_ CreateSensorCmdPort = (*SendCmdAdapter)(nil)
	_ DeleteSensorCmdPort = (*SendCmdAdapter)(nil)
)

func (adapter *DbSensorAdapter) CreateSensor(
	sensorId uuid.UUID, 
	gatewayId uuid.UUID, 
	name string, 
	interval time.Duration, 
	profile profile.SensorProfile,
) (Sensor, error) {
	entity := &SensorEntity{
		ID:        sensorId.String(),
		GatewayID: gatewayId.String(),
		Name:      name,
		Interval:  interval.Milliseconds(),
		Profile:   string(profile),
	}
	err := adapter.repo.CreateSensor(entity)
	if err != nil {
		adapter.log.Error("Failed to create sensor", zap.Error(err))
		return Sensor{}, err
	}

	sensor, err := SensorEntityToDomain(entity)
	return sensor, err
}

func (adapter *DbSensorAdapter) DeleteSensor(sensorId uuid.UUID) (Sensor, error) {
	entity := &SensorEntity{
		ID: sensorId.String(),
	}
	err := adapter.repo.DeleteSensor(entity)
	if err != nil {
		adapter.log.Error("Failed to delete sensor", zap.Error(err))
		return Sensor{}, err
	}
	sensor, err := SensorEntityToDomain(entity)
	return sensor, err
}

func (adapter *SendCmdAdapter) SendCreateSensorCmd(
	sensorId uuid.UUID, 
	gatewayId uuid.UUID, 
	interval time.Duration, 
	profile profile.SensorProfile,
) error {
	cmd := &CreateSensorCmdEntity{
		SensorId:  sensorId.String(),
		GatewayId: gatewayId.String(),
		Interval:  interval.Milliseconds(),
		Profile:   string(profile),
	}
	err := adapter.repo.SendCreateSensorCmd(cmd)
	if err != nil {
		adapter.log.Error("Failed to send create sensor command", zap.Error(err))
		return err
	}
	return nil
}

func (adapter *SendCmdAdapter) SendDeleteSensorCmd(sensorId uuid.UUID, gatewayId uuid.UUID) error {
	cmd := &DeleteSensorCmdEntity{
		SensorId:  sensorId.String(),
		GatewayId: gatewayId.String(),
	}
	err := adapter.repo.SendDeleteSensorCmd(cmd)
	if err != nil {
		adapter.log.Error("Failed to send delete sensor command", zap.Error(err))
		return err
	}
	return nil
}
