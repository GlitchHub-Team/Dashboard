package gateway

import (
	"backend/internal/infra/database"
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/gateway/mocks/save_remove_get_port.go -package=mocks . CreateGatewayPort,SaveGatewayPort,DeleteGatewayPort,GetGatewayPort,GetGatewaysPort
type GatewayPostgreAdapter struct {
	repo GatewayRepository
	log  *zap.Logger
}

type SaveGatewayPort interface {
	Save(g Gateway) (Gateway, error)
}

type CreateGatewayPort interface {
	Create(g Gateway) (Gateway, error)
}

type DeleteGatewayPort interface {
	Delete(gatewayId uuid.UUID) (Gateway, error)
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
	entity := &GatewayEntity{
		ID:               g.Id.String(),
		Name:             g.Name,
		Interval:         g.IntervalLimit.Milliseconds(),
		Status:           string(g.Status),
		PublicIdentifier: g.PublicIdentifier,
	}

	if g.TenantId != nil {
		tenantIdStr := g.TenantId.String()
		entity.TenantId = &tenantIdStr
	} else {
		entity.TenantId = nil
	}

	if err := a.repo.SaveGateway(entity); err != nil {
		return Gateway{}, err
	}
	return g, nil
}

func (a *GatewayPostgreAdapter) Create(g Gateway) (Gateway, error) {
	var tenantId *string
	if g.TenantId != nil {
		tenantIdStr := g.TenantId.String()
		tenantId = &tenantIdStr
	}

	entity := &GatewayEntity{
		ID:               g.Id.String(),
		Name:             g.Name,
		Interval:         g.IntervalLimit.Milliseconds(),
		Status:           string(g.Status),
		TenantId:         tenantId,
		PublicIdentifier: g.PublicIdentifier,
	}

	return a.repo.CreateGateway(entity)
}

func (a *GatewayPostgreAdapter) Delete(gatewayId uuid.UUID) (Gateway, error) {
	entity := &GatewayEntity{
		ID: gatewayId.String(),
	}

	if err := a.repo.DeleteGateway(entity); err != nil {
		return Gateway{}, err
	}

	return entity.ToGateway(), nil
}

/*   ================================   */

func (a *GatewayPostgreAdapter) GetById(gatewayId uuid.UUID) (Gateway, error) {
	entity, err := a.repo.GetGatewayById(gatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	gw, err := GatewayEntityToDomain(&entity)
	if err != nil {
		return Gateway{}, err
	}

	return gw, nil
}

func (a *GatewayPostgreAdapter) GetGatewayByTenantID(tenantId uuid.UUID, gatewayId uuid.UUID) (Gateway, error) {
	entity, err := a.repo.GetGatewayByTenantID(tenantId.String(), gatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	gw, err := GatewayEntityToDomain(&entity)
	if err != nil {
		return Gateway{}, err
	}

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
	if err != nil {
		return nil, 0, err
	}

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
	if err != nil {
		return nil, 0, err
	}
	return gws, count, nil
}

var (
	_ CreateGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ SaveGatewayPort   = (*GatewayPostgreAdapter)(nil)
	_ DeleteGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort    = (*GatewayPostgreAdapter)(nil)
	_ GetGatewaysPort   = (*GatewayPostgreAdapter)(nil)
)
