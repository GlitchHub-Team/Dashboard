package sensor

import (
	"backend/internal/gateway"
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/port_getById.go -package=mocks . GetSensorByIdPort

type GetSensorByIdPort interface {
	GetSensorById(sensorId uuid.UUID) (Sensor, error)
}

type GetSensorByIdService struct {
	getSensorByIdPort GetSensorByIdPort
	getGatewayPort    gateway.GetGatewayPort
}

func NewGetSensorByIdService(getSensorByIdPort GetSensorByIdPort, getGatewayPort gateway.GetGatewayPort) *GetSensorByIdService {
	return &GetSensorByIdService{
		getSensorByIdPort: getSensorByIdPort,
		getGatewayPort:    getGatewayPort,
	}
}

func (s *GetSensorByIdService) GetSensorById(cmd GetSensorCommand) (Sensor, error) {
	sensor, err := s.getSensorByIdPort.GetSensorById(cmd.SensorId)
	if err != nil {
		return Sensor{}, err
	}

	if sensor.IsZero() {
		return Sensor{}, ErrSensorNotFound
	}

	gat, err := s.getGatewayPort.GetById(sensor.GatewayId.String())
	if err != nil {
		return Sensor{}, err
	}

	if gat.IsZero() {
		return Sensor{}, gateway.ErrGatewayNotFound
	}

	// Se è super admin può accedere a tutti i sensori di ogni gateway, altrimenti controllo se il gateway(e conseguentemente il sensore) appartiene al tenant dell'utente
	if !cmd.IsSuperAdmin() {
		if cmd.RequesterTenantId == nil || !gat.BelongsToTenant(*cmd.RequesterTenantId) {
			return Sensor{}, identity.ErrUnauthorizedAccess
		}
	}

	return sensor, nil
}
