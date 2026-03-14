package gateway

type SaveGatewayPort interface {
	Save(g Gateway) error
}

type RemoveGatewayPort interface {
	Remove(g Gateway) error
}

type GetGatewayPort interface {
	GetById() error
	GetByTenantId() error
	GetAll() error
}

type GatewayPostgreAdapter struct {
	repository gatewayPostgreRepository
}

func NewGatewayPostgreAdapter(repository gatewayPostgreRepository) (SaveGatewayPort, RemoveGatewayPort, GetGatewayPort) {
	adapter := &GatewayPostgreAdapter{
		repository: repository,
	}
	return adapter, adapter, adapter
}

func (adapter *GatewayPostgreAdapter) Save(g Gateway) error {
	// ...
	return nil
}

func (adapter *GatewayPostgreAdapter) Remove(g Gateway) error {
	// ...
	return nil
}

func (a *GatewayPostgreAdapter) GetById() error {
	return nil
}
func (a *GatewayPostgreAdapter) GetByTenantId() error {
	return nil
}
func (a *GatewayPostgreAdapter) GetAll() error {
	return nil
}



var (
	_ SaveGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ RemoveGatewayPort = (*GatewayPostgreAdapter)(nil)
	_ GetGatewayPort = (*GatewayPostgreAdapter)(nil)
)