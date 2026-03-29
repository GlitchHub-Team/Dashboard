package sensor

import (
	"backend/internal/gateway"
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/port_interrupt_resume.go -package=mocks . SendInterruptCmdPort,SendResumeCmdPort,UpdatedSensorStatusPort

type InterruptSensorService struct {
	sendInterruptCmdPort    SendInterruptCmdPort
	getSensorPort           GetSensorByIdPort
	getGatewayPort          gateway.GetGatewayPort
	updatedSensorStatusPort UpdatedSensorStatusPort
}

type ResumeSensorService struct {
	sendResumeCmdPort       SendResumeCmdPort
	getSensorPort           GetSensorByIdPort
	getGatewayPort          gateway.GetGatewayPort
	updatedSensorStatusPort UpdatedSensorStatusPort
}

type SendResumeCmdPort interface {
	SendResume(sensorId uuid.UUID) error
}

type SendInterruptCmdPort interface {
	SendInterrupt(sensorId uuid.UUID) error
}

type UpdatedSensorStatusPort interface {
	UpdatedSensorStatus(sensorId uuid.UUID, status SensorStatus) error
}

func NewInterruptSensorService(sendInterruptCmdPort SendInterruptCmdPort, getSensorPort GetSensorByIdPort, getGatewayPort gateway.GetGatewayPort, updatedSensorStatusPort UpdatedSensorStatusPort) *InterruptSensorService {
	return &InterruptSensorService{
		sendInterruptCmdPort:    sendInterruptCmdPort,
		getSensorPort:           getSensorPort,
		getGatewayPort:          getGatewayPort,
		updatedSensorStatusPort: updatedSensorStatusPort,
	}
}

func (s *InterruptSensorService) InterruptSensor(cmd InterruptSensorCommand) error {
	// Controllo che il sensore esista
	sensor, err := s.getSensorPort.GetSensorById(cmd.SensorId)
	if err != nil {
		return err
	}

	if sensor.IsZero() {
		return ErrSensorNotFound
	}

	if sensor.Status != Active {
		return ErrSensorNotActive
	}

	// Controllo che il gateway esista
	gat, err := s.getGatewayPort.GetById(sensor.GatewayId.String())
	if err != nil {
		return err
	}

	if gat.IsZero() {
		return gateway.ErrGatewayNotFound
	}

	// Controllo che l'utente sia super admin
	if !cmd.IsSuperAdmin() {
		if cmd.RequesterTenantId == nil || !gat.BelongsToTenant(*cmd.RequesterTenantId) {
			return identity.ErrUnauthorizedAccess
		}
	}

	// Invia comando di interruzione sensore al gateway simulato
	err = s.sendInterruptCmdPort.SendInterrupt(cmd.SensorId)
	if err != nil {
		return err
	}

	// Aggiorna lo stato del sensore a Inactive
	return s.updatedSensorStatusPort.UpdatedSensorStatus(cmd.SensorId, Inactive)
}

func NewResumeSensorService(sendResumeCmdPort SendResumeCmdPort, getSensorPort GetSensorByIdPort, getGatewayPort gateway.GetGatewayPort, updatedSensorStatusPort UpdatedSensorStatusPort) *ResumeSensorService {
	return &ResumeSensorService{
		sendResumeCmdPort:       sendResumeCmdPort,
		getSensorPort:           getSensorPort,
		getGatewayPort:          getGatewayPort,
		updatedSensorStatusPort: updatedSensorStatusPort,
	}
}

func (s *ResumeSensorService) ResumeSensor(cmd ResumeSensorCommand) error {
	// Controllo che il sensore esista
	sensor, err := s.getSensorPort.GetSensorById(cmd.SensorId)
	if err != nil {
		return err
	}

	if sensor.IsZero() {
		return ErrSensorNotFound
	}

	if sensor.Status != Inactive {
		return ErrSensorNotInactive
	}

	// Controllo che il gateway esista
	gat, err := s.getGatewayPort.GetById(sensor.GatewayId.String())
	if err != nil {
		return err
	}

	if gat.IsZero() {
		return gateway.ErrGatewayNotFound
	}

	// Controllo che l'utente sia super admin
	if !cmd.IsSuperAdmin() {
		if cmd.RequesterTenantId == nil || !gat.BelongsToTenant(*cmd.RequesterTenantId) {
			return identity.ErrUnauthorizedAccess
		}
	}

	// Invia comando di ripresa sensore al gateway simulato
	err = s.sendResumeCmdPort.SendResume(cmd.SensorId)
	if err != nil {
		return err
	}

	// Aggiorna lo stato del sensore a Active
	return s.updatedSensorStatusPort.UpdatedSensorStatus(cmd.SensorId, Active)
}
