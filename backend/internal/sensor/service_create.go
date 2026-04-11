package sensor

import (
	"time"

	"backend/internal/gateway"
	profile "backend/internal/sensor/profile"
	"backend/internal/shared/identity"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/port_create.go -package=mocks . CreateSensorPort,CreateSensorCmdPort

type CreateSensorPort interface {
	CreateSensor(sensorId uuid.UUID, gatewayId uuid.UUID, name string, interval time.Duration, profile profile.SensorProfile) (Sensor, error)
}

type CreateSensorCmdPort interface {
	SendCreateSensorCmd(sensord uuid.UUID, gatewayId uuid.UUID, interval time.Duration, profile profile.SensorProfile) error
}

type CreateSensorService struct {
	createSensorPort        CreateSensorPort
	sendCreateSensorCmdPort CreateSensorCmdPort
	getGatewayPort          gateway.GetGatewayPort
}

func NewCreateSensorService(
	createSensorPort CreateSensorPort,
	sendCreateSensorCmdPort CreateSensorCmdPort,
	getGatewayPort gateway.GetGatewayPort,
) *CreateSensorService {
	return &CreateSensorService{
		createSensorPort:        createSensorPort,
		sendCreateSensorCmdPort: sendCreateSensorCmdPort,
		getGatewayPort:          getGatewayPort,
	}
}

func (s *CreateSensorService) CreateSensor(cmd CreateSensorCommand) (Sensor, error) {
	// Controllo che il gateway esista
	gat, err := s.getGatewayPort.GetById(cmd.GatewayId)
	if err != nil {
		return Sensor{}, err
	}

	if gat.IsZero() {
		return Sensor{}, gateway.ErrGatewayNotFound
	}

	// Controllo che l'utente sia super admin
	if !cmd.IsSuperAdmin() {
		return Sensor{}, identity.ErrUnauthorizedAccess
	}

	sensorId := uuid.New()

	// Invia comando di creazione sensore al gateway simulato
	err = s.sendCreateSensorCmdPort.SendCreateSensorCmd(sensorId, cmd.GatewayId, cmd.Interval, cmd.Profile)
	if err != nil {
		return Sensor{}, err
	}
	// Salva il sensore nel database
	return s.createSensorPort.CreateSensor(sensorId, cmd.GatewayId, cmd.Name, cmd.Interval, cmd.Profile)
}

// Compile-time checks
var (
	_ CreateSensorUseCase = (*CreateSensorService)(nil)
)
