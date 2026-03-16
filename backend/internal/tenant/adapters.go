package tenant

import (
	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/tenant/mocks/ports.go -package=mocks . GetTenantPort,GetTenantsPort

type GetTenantPort interface {
	GetTenant(tenantId uuid.UUID) (Tenant, error)
}

type GetTenantsPort interface {
	GetTenants() ([]Tenant, error)
}

type GetTenantsPostgreAdapter struct {
	repository *TenantPostgreRepository
}
func NewGetTenantsPostgreAdapter(repository *TenantPostgreRepository,) (GetTenantPort, GetTenantsPort) {
	adapter := &GetTenantsPostgreAdapter{
		repository: repository,
	}
	return adapter, adapter
}


func (adapter *GetTenantsPostgreAdapter) mapTenantEntity(entity TenantEntity) (*Tenant, error) {

	tenantId, err := uuid.Parse(entity.ID)
	if err != nil {
		return nil, err
	}
	return &Tenant{
		Id: tenantId,
		Name: entity.Name,
		CanImpersonate: entity.CanImpersonate,
	}, nil
}


func (adapter *GetTenantsPostgreAdapter) GetTenant(tenantId uuid.UUID) (Tenant, error) {
	tenantEntity, err := adapter.repository.GetTenant(tenantId.String())
	if err != nil {
		return Tenant{}, err
	}

	tenant, err := tenantEntity.toTenant()
	return tenant, err
}

func (adapter *GetTenantsPostgreAdapter) GetTenants() ([]Tenant, error) {
	tenantEntities, err := adapter.repository.GetAllTenants()
	if err != nil {
		return nil, err
	}

	var tenants []Tenant

	for _, entity := range tenantEntities {
		tenant, err := adapter.mapTenantEntity(entity)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, *tenant)
	}

	return tenants, nil
}

