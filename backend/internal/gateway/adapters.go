package gateway

import (
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

// TODO: fix hexagonal sbagliato
func (a *GatewayPostgreAdapter) GetById(gatewayId string) (Gateway, error) {
	entity, err := a.repo.GetGatewayById(gatewayId)
	if err != nil {
		return Gateway{}, err
	}
	return entity, nil
}

// TODO: fix hexagonal sbagliato
func (a *GatewayPostgreAdapter) GetByTenantId(tenantId string) ([]Gateway, error) {
	gateways, err := a.repo.GetGatewaysByTenantId(tenantId)
	if err != nil {
		return nil, err
	}
	// gateways := make([]Gateway, len(entities))
	// for i := range entities {
	// 	gateways[i] = entities[i].ToGateway()
	// }
	return gateways, nil
}

// TODO: fix hexagonal sbagliato
func (a *GatewayPostgreAdapter) GetAll() ([]Gateway, error) {
	gateways, err := a.repo.GetAllGateways()
	if err != nil {
		return nil, err
	}
	// gateways := make([]Gateway, len(entities))
	// for i := range entities {
	// 	gateways[i] = entities[i].ToGateway()
	// }
	return gateways, nil
}

var (
	_ CreateGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ SaveGatewayPort   = (*GatewayPostgreAdapter)(nil)
	_ DeleteGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort    = (*GatewayPostgreAdapter)(nil)
	_ GetGatewaysPort   = (*GatewayPostgreAdapter)(nil)
)
