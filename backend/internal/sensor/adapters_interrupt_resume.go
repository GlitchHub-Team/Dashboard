package sensor

import "github.com/google/uuid"

// Compile-time checks
var (
	_ SendResumeCmdPort      = (*SendCmdAdapter)(nil)
	_ SendInterruptCmdPort   = (*SendCmdAdapter)(nil)
	_ UpdateSensorStatusPort = (*DbSensorAdapter)(nil)
)

func (adapter *SendCmdAdapter) SendResume(sensorId uuid.UUID, gatewayId uuid.UUID) error {
	return adapter.repo.SendResumeSensorCmd(&ResumeSensorCmdEntity{
		SensorId:  sensorId.String(),
		GatewayId: gatewayId.String(),
	})
}

func (adapter *SendCmdAdapter) SendInterrupt(sensorId uuid.UUID, gatewayId uuid.UUID) error {
	return adapter.repo.SendInterruptSensorCmd(&InterruptSensorCmdEntity{
		SensorId:  sensorId.String(),
		GatewayId: gatewayId.String(),
	})
}

func (adapter *DbSensorAdapter) UpdateSensorStatus(sensor Sensor, status SensorStatus) error {
	return adapter.repo.UpdateSensor(sensor.Id.String(), string(status))
}
