package gateway

type GatewayPostgreAdapter struct {
	repository gatewayPostgreRepository
}

type SaveGatewayPort interface {
	Save(g Gateway) error
}

type RemoveGatewayPort interface {
	Remove(g Gateway) error
}

type GetGatewayPort interface {
	GetById(gatewayId string) error
	GetByTenantId(tenantId string) error
	GetAll() ([]Gateway, error)
}

func NewGatewayPostgreAdapter(repository gatewayPostgreRepository) (SaveGatewayPort, RemoveGatewayPort, GetGatewayPort) {
	adapter := &GatewayPostgreAdapter{
		repository: repository,
	}
	return adapter, adapter, adapter
}

func (adapter *GatewayPostgreAdapter) Save(g Gateway) error {
	
	return nil
}

func (adapter *GatewayPostgreAdapter) Remove(g Gateway) error {
	return nil
}

func (a *GatewayPostgreAdapter) GetById(gatewayId string) error {
	return nil
}

func (a *GatewayPostgreAdapter) GetByTenantId(tenantId string) error {
	return nil
}

func (a *GatewayPostgreAdapter) GetAll() ([]Gateway, error) {
	return nil, nil
}

var (
	_ SaveGatewayPort   = (*GatewayPostgreAdapter)(nil)
	_ RemoveGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort    = (*GatewayPostgreAdapter)(nil)
)
