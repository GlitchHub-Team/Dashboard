package gateway

import (
	"backend/internal/infra/database"
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/gateway/mocks/save_remove_get_port.go -package=mocks . SaveGatewayPort,RemoveGatewayPort,GetGatewayPort,GetGatewaysPort
type GatewayPostgreAdapter struct {
	repo GatewayRepository
	log  *zap.Logger
}

type SaveGatewayPort interface {
	Save(g Gateway) (Gateway, error)
}

type RemoveGatewayPort interface {
	Remove(gatewayId uuid.UUID) (Gateway, error)
}

type GetGatewayPort interface {
	GetById(gatewayId uuid.UUID) (Gateway, error)
	GetGatewayByTenantID(tenantId uuid.UUID, gatewayId uuid.UUID) (Gateway, error)
}

type GetGatewaysPort interface {
	GetByTenantId(tenantId uuid.UUID, page int, limit int) ([]Gateway, uint, error)
	GetAll(page int, limit int) ([]Gateway, uint, error)
}

func NewGatewayPostgreAdapter(repository GatewayRepository, log *zap.Logger) *GatewayPostgreAdapter {
	adapter := GatewayPostgreAdapter{
		repo: repository,
		log:  log,
	}
	return &adapter
}

func (a *GatewayPostgreAdapter) Save(g Gateway) (Gateway, error) {
	if err := a.repo.SaveGateway(g); err != nil {
		return Gateway{}, err
	}
	return g, nil
}

func (a *GatewayPostgreAdapter) Remove(gatewayId uuid.UUID) (Gateway, error) {
	existing, err := a.repo.GetGatewayById(gatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if err := a.repo.DeleteGateway(existing); err != nil {
		return Gateway{}, err
	}

	return existing, nil
}

/*   ================================   */

func (a *GatewayPostgreAdapter) GetById(gatewayId uuid.UUID) (Gateway, error) {
	entity, err := a.repo.GetGatewayById(gatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	gw, err := GatewayEntityToDomain(&entity)

	return gw, nil
}

func (a *GatewayPostgreAdapter) GetGatewayByTenantID(tenantId uuid.UUID, gatewayId uuid.UUID) (Gateway, error) {
	entity, err := a.repo.GetGatewayByTenantID(tenantId.String(), gatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	gw, err := GatewayEntityToDomain(&entity)

	return gw, nil
}

func (a *GatewayPostgreAdapter) GetByTenantId(tenantId uuid.UUID, page int, limit int) ([]Gateway, uint, error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	gateways, count, err := a.repo.GetGatewaysByTenantId(tenantId.String(), offset, limit)
	if err != nil {
		return nil, 0, err
	}
	gws, err := database.MapEntityListToDomain(gateways, GatewayEntityToDomain)

	return gws, count, nil
}

func (a *GatewayPostgreAdapter) GetAll(page int, limit int) ([]Gateway, uint, error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	gateways, count, err := a.repo.GetAllGateways(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	gws, err := database.MapEntityListToDomain(gateways, GatewayEntityToDomain)
	return gws, count, nil
}

var (
	_ SaveGatewayPort   = (*GatewayPostgreAdapter)(nil)
	_ RemoveGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort    = (*GatewayPostgreAdapter)(nil)
	_ GetGatewaysPort   = (*GatewayPostgreAdapter)(nil)
)
