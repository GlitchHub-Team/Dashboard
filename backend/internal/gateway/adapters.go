package gateway

import "github.com/google/uuid"

type SaveGatewayPort interface {
	Save(g Gateway) error
}

type RemoveGatewayPort interface {
	Remove(g Gateway) error
}

type GetGatewayPort interface {
	GetById(id uuid.UUID) (Gateway, error)
	GetByTenantId() error
	GetAll() error
}

type GatewayPostgreAdapter struct {
	repository GatewayRepository
}

func NewGatewayPostgreAdapter(repository GatewayRepository) (SaveGatewayPort, RemoveGatewayPort, GetGatewayPort) {
	adapter := &GatewayPostgreAdapter{
		repository: repository,
	}
	return adapter, adapter, adapter
}

func (adapter *GatewayPostgreAdapter) Save(g Gateway) error {
	return adapter.repository.Save(g)
}

var ()

func (adapter *GatewayPostgreAdapter) Remove(g Gateway) error {
	// ...
	return nil
}

func (a *GatewayPostgreAdapter) GetById(id uuid.UUID) (Gateway, error) {
	return a.repository.GetById(id)
}

func (a *GatewayPostgreAdapter) GetByTenantId() error {
	return nil
}

func (a *GatewayPostgreAdapter) GetAll() error {
	return nil
}

var (
	_ SaveGatewayPort   = (*GatewayPostgreAdapter)(nil)
	_ RemoveGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort    = (*GatewayPostgreAdapter)(nil)
)
