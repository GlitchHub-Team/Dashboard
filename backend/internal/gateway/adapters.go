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

func NewGatewayPostgreAdapter(repository gatewayPostgreRepository) *GatewayPostgreAdapter {
	return &GatewayPostgreAdapter{
		repository: repository,
	}
}


func (adapter *GatewayPostgreAdapter) Save(g Gateway) error {
	// ...
	return nil
}
