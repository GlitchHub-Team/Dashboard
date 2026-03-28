package sensor

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/port_getByTenantId.go -package=mocks . GetSensorsByTenantIdPort

type GetSensorByTenantIdService struct {
	getSensorsByTenantPort GetSensorsByTenantIdPort
}

type GetSensorsByTenantIdPort interface {
	GetSensorsByTenant(tenantId uuid.UUID, page int, limit int) ([]Sensor, uint, error)
}

func NewGetSensorByTenantIdService(getSensorsByTenantPort GetSensorsByTenantIdPort) *GetSensorByTenantIdService {
	return &GetSensorByTenantIdService{
		getSensorsByTenantPort: getSensorsByTenantPort,
	}
}

func (s *GetSensorByTenantIdService) GetSensorsByTenant(cmd GetSensorsByTenantCommand) ([]Sensor, uint, error) {
	// Controllo se il tenantId appartiene al requester o se il requester è super admin
	if !cmd.IsSuperAdmin() {
		// Se non è admin, deve avere un tenantId e deve coincidere con quello richiesto
		if cmd.RequesterTenantId == nil || *cmd.RequesterTenantId != cmd.TenantId {
			return nil, 0, identity.ErrUnauthorizedAccess
		}
	}

	sensors, total, err := s.getSensorsByTenantPort.GetSensorsByTenant(cmd.TenantId, cmd.Page, cmd.Limit)
	if err != nil {
		return nil, 0, err
	}

	return sensors, total, nil
}
