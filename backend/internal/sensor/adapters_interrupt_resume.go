package sensor

import "github.com/google/uuid"

// Compile-time checks
var (
	_ SendResumeCmdPort       = (*SendCmdAdapter)(nil)
	_ SendInterruptCmdPort    = (*SendCmdAdapter)(nil)
	_ UpdatedSensorStatusPort = (*DbSensorAdapter)(nil)
)

func (adapater *SendCmdAdapter) SendResume(sensorId uuid.UUID, gatewayId uuid.UUID) error {
	return adapater.repo.SendResumeSensorCmd(&ResumeSensorCmdEntity{
		SensorId:  sensorId.String(),
		GatewayId: gatewayId.String(),
	})
}

func (adapater *SendCmdAdapter) SendInterrupt(sensorId uuid.UUID, gatewayId uuid.UUID) error {
	return adapater.repo.SendInterruptSensorCmd(&InterruptSensorCmdEntity{
		SensorId:  sensorId.String(),
		GatewayId: gatewayId.String(),
	})
}

func (adapater *DbSensorAdapter) UpdateSensorStatus(sensor Sensor, status SensorStatus) error {
	return adapater.repo.UpdateSensor(sensor.Id.String(), string(status))
}
