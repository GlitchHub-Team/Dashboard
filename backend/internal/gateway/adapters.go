package gateway

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/gateway/mocks/save_remove_get_port.go -package=mocks . SaveGatewayPort,RemoveGatewayPort,GetGatewayPort

type GatewayPostgreAdapter struct {
	repo *gatewayPostgreRepository
	log  *zap.Logger
}

type SaveGatewayPort interface {
	Save(g Gateway) (Gateway, error)
}

type RemoveGatewayPort interface {
	Remove(gatewayId uuid.UUID) (Gateway, error)
}

type GetGatewayPort interface {
	GetById(gatewayId string) (Gateway, error)
	GetByTenantId(tenantId string) ([]Gateway, error)
	GetAll() ([]Gateway, error)
}

func NewGatewayPostgreAdapter(repository *gatewayPostgreRepository, log *zap.Logger) *GatewayPostgreAdapter {
	adapter := &GatewayPostgreAdapter{
		repo: repository,
		log:  log,
	}
	return adapter
}

func (a *GatewayPostgreAdapter) Save(g Gateway) (Gateway, error) {
	if err := a.repo.SaveGateway(g); err != nil {
		return Gateway{}, err
	}
	return g, nil
}

func (a *GatewayPostgreAdapter) Remove(gatewayId uuid.UUID) (Gateway, error) {
	g := Gateway{Id: gatewayId}
	if err := a.repo.DeleteGateway(g); err != nil {
		return Gateway{}, err
	}
	return g, nil
}

func (a *GatewayPostgreAdapter) GetById(gatewayId string) (Gateway, error) {
	entity, err := a.repo.GetGatewayById(gatewayId)
	if err != nil {
		return Gateway{}, err
	}
	return entity.toGateway(), nil
}

func (a *GatewayPostgreAdapter) GetByTenantId(tenantId string) ([]Gateway, error) {
	entities, err := a.repo.GetGatewaysByTenantId(tenantId)
	if err != nil {
		return nil, err
	}
	gateways := make([]Gateway, len(entities))
	for i := range entities {
		gateways[i] = entities[i].toGateway()
	}
	return gateways, nil
}

func (a *GatewayPostgreAdapter) GetAll() ([]Gateway, error) {
	entities, err := a.repo.GetAllGateways()
	if err != nil {
		return nil, err
	}
	gateways := make([]Gateway, len(entities))
	for i := range entities {
		gateways[i] = entities[i].toGateway()
	}
	return gateways, nil
}
