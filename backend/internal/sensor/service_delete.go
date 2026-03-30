package sensor

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/port_delete.go -package=mocks . DeleteSensorPort,DeleteSensorCmdPort

type DeleteSensorPort interface {
	DeleteSensor(sensorId uuid.UUID) (Sensor, error)
}

type DeleteSensorCmdPort interface {
	SendDeleteSensorCmd(sensord uuid.UUID, gatewayId uuid.UUID) error
}

type DeleteSensorService struct {
	deleteSensorPort    DeleteSensorPort
	getSensorByIdPort   GetSensorByIdPort
	deleteSensorCmdPort DeleteSensorCmdPort
}

func NewDeleteSensorService(deleteSensorPort DeleteSensorPort, getSensorByIdPort GetSensorByIdPort, deleteSensorCmdPort DeleteSensorCmdPort) *DeleteSensorService {
	return &DeleteSensorService{
		deleteSensorPort:    deleteSensorPort,
		getSensorByIdPort:   getSensorByIdPort,
		deleteSensorCmdPort: deleteSensorCmdPort,
	}
}

func (s *DeleteSensorService) DeleteSensor(cmd DeleteSensorCommand) (Sensor, error) {
	// Controllo che il sensore esista
	sensor, err := s.getSensorByIdPort.GetSensorById(cmd.SensorId)
	if err != nil {
		return Sensor{}, err
	}

	if sensor.IsZero() {
		return Sensor{}, ErrSensorNotFound
	}

	// Controllo che l'utente sia super admin
	if !cmd.IsSuperAdmin() {
		return Sensor{}, identity.ErrUnauthorizedAccess
	}

	// Invia comando di eliminazione sensore al gateway simulato
	err = s.deleteSensorCmdPort.SendDeleteSensorCmd(cmd.SensorId, sensor.GatewayId)
	if err != nil {
		return Sensor{}, err
	}
	return s.deleteSensorPort.DeleteSensor(cmd.SensorId)
}

// Compile-time checks
var (
	_ DeleteSensorUseCase = (*DeleteSensorService)(nil)
)
