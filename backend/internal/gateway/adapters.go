package gateway

import (
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
	GetById(gatewayId string) (Gateway, error)
}

type GetGatewaysPort interface {
	GetByTenantId(tenantId string) ([]Gateway, error)
	GetAll() ([]Gateway, error)
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

func (a *GatewayPostgreAdapter) GetById(gatewayId string) (Gateway, error) {
	entity, err := a.repo.GetGatewayById(gatewayId)
	if err != nil {
		return Gateway{}, err
	}
	return entity, nil
}

func (a *GatewayPostgreAdapter) GetByTenantId(tenantId string) ([]Gateway, error) {
	gateways, err := a.repo.GetGatewaysByTenantId(tenantId)
	if err != nil {
		return nil, err
	}
	return gateways, nil
}

func (a *GatewayPostgreAdapter) GetAll() ([]Gateway, error) {
	gateways, err := a.repo.GetAllGateways()
	if err != nil {
		return nil, err
	}
	return gateways, nil
}

var (
	_ SaveGatewayPort   = (*GatewayPostgreAdapter)(nil)
	_ RemoveGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort    = (*GatewayPostgreAdapter)(nil)
	_ GetGatewaysPort   = (*GatewayPostgreAdapter)(nil)
)