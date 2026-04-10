package sensor

import (
	"backend/internal/gateway"
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/port_getByGatewayId.go -package=mocks . GetSensorsByGatewayIdPort

type GetSensorsByGatewayIdPort interface {
	GetSensorsByGatewayId(gatewayId uuid.UUID, page int, limit int) ([]Sensor, uint, error)
}

type GetSensorsByGatewayIdService struct {
	getSensorsByGatewayIdPort GetSensorsByGatewayIdPort
	getGatewayPort            gateway.GetGatewayPort
}

func NewGetSensorsByGatewayIdService(getSensorsByGatewayIdPort GetSensorsByGatewayIdPort, getGatewayPort gateway.GetGatewayPort) *GetSensorsByGatewayIdService {
	return &GetSensorsByGatewayIdService{
		getSensorsByGatewayIdPort: getSensorsByGatewayIdPort,
		getGatewayPort:            getGatewayPort,
	}
}

func (s *GetSensorsByGatewayIdService) GetSensorsByGateway(cmd GetSensorsByGatewayCommand) ([]Sensor, uint, error) {
	// Controllo che il gateway esista
	gat, err := s.getGatewayPort.GetById(cmd.GatewayId)
	if err != nil {
		return nil, 0, err
	}

	if gat.IsZero() {
		return nil, 0, gateway.ErrGatewayNotFound
	}

	// Se è super admin può accedere a tutti i sensori di ogni gateway, altrimenti controllo se il gateway appartiene al tenant dell'utente
	if !cmd.IsSuperAdmin() {
		if cmd.RequesterTenantId == nil || !gat.BelongsToTenant(*cmd.RequesterTenantId) {
			return nil, 0, identity.ErrUnauthorizedAccess
		}
	}

	return s.getSensorsByGatewayIdPort.GetSensorsByGatewayId(cmd.GatewayId, cmd.Page, cmd.Limit)
}

// Compile-time checks
var (
	_ GetSensorsByGatewayUseCase = (*GetSensorsByGatewayIdService)(nil)
)
