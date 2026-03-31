package tenant

import (
	"errors"

	"github.com/google/uuid"
)

// ROBE ===============================================================================================

func (TenantEntity) TableName() string { return "tenants" }

func (entity *TenantEntity) fromTenant(tenant Tenant) {
	if tenant.Id != uuid.Nil {
		entity.ID = tenant.Id.String()
	} else {
		entity.ID = uuid.New().String()
	}
	entity.Name = tenant.Name
	entity.CanImpersonate = tenant.CanImpersonate
}

func (entity *TenantEntity) toTenant() (Tenant, error) {
	if entity.ID == "" {
		return Tenant{}, nil
	}

	id, err := uuid.Parse(entity.ID)
	if err != nil {
		return Tenant{}, err
	}

	return Tenant{
		Id:             id,
		Name:           entity.Name,
		CanImpersonate: entity.CanImpersonate,
	}, nil
}

type TenantPostgreAdapter struct {
	repo *TenantPostgreRepository
}

// Repository =========================================================================================

// Adapter ============================================================================================

//go:generate mockgen -destination=../../tests/tenant/mocks/ports.go -package=mocks . CreateTenantPort,DeleteTenantPort,GetTenantPort,GetTenantsPort,GetTenantByUserPort

func NewTenantPostgreAdapter(repository *TenantPostgreRepository) (
	CreateTenantPort,
	DeleteTenantPort,
	GetTenantPort,
	GetTenantsPort,
	GetTenantByUserPort,
) {
	adapter := &TenantPostgreAdapter{repo: repository}
	return adapter, adapter, adapter, adapter, adapter
}

// CREATE =============================================================================================

func (adapter *TenantPostgreAdapter) CreateTenant(tenant Tenant) (Tenant, error) {
	entity := &TenantEntity{}

	entity.fromTenant(tenant)

	err := adapter.repo.SaveTenant(entity)
	if err != nil {
		return Tenant{}, err
	}

	return entity.toTenant()
}

// DELETE =============================================================================================

type DeleteTenantPort interface {
	DeleteTenant(tenantId uuid.UUID) (Tenant, error)
}

func (adapter *TenantPostgreAdapter) DeleteTenant(tenantId uuid.UUID) (Tenant, error) {
	oldEntity, err := adapter.repo.GetTenant(tenantId.String())
	if err != nil {
		return Tenant{}, err
	}

	if oldEntity.ID == "" {
		return Tenant{}, errors.New("tenant not found")
	}

	oldTenant, err := adapter.repo.DeleteTenant(oldEntity)
	if err != nil {
		return Tenant{}, err
	}

	return oldTenant, nil
}

// GET ================================================================================================

type GetTenantPort interface {
	GetTenant(tenantId uuid.UUID) (Tenant, error)
}

type GetTenantsPort interface {
	GetTenants() ([]Tenant, error)
}

type GetTenantByUserPort interface {
	GetTenantByUser(userId uuid.UUID) (Tenant, error)
}

func (adapter *TenantPostgreAdapter) GetTenant(tenantId uuid.UUID) (Tenant, error) {
	entity, err := adapter.repo.GetTenant(tenantId.String())
	if err != nil {
		return Tenant{}, err
	}
	return entity.toTenant()
}

func (adapter *TenantPostgreAdapter) GetTenants() ([]Tenant, error) {
	entities, err := adapter.repo.GetAllTenants()
	if err != nil {
		return nil, err
	}

	var tenants []Tenant
	for _, entity := range entities {
		tenant, err := entity.toTenant()
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}

	return tenants, nil
}

func (adapter *TenantPostgreAdapter) GetTenantByUser(userId uuid.UUID) (Tenant, error) {
	entity, err := adapter.repo.GetTenantByUser(userId.String())
	if err != nil {
		return Tenant{}, err
	}
	return entity.toTenant()
}

// Compile-time checks ================================================================================
var (
	_ CreateTenantPort = (*TenantPostgreAdapter)(nil)
	_ DeleteTenantPort = (*TenantPostgreAdapter)(nil)
	_ GetTenantPort    = (*TenantPostgreAdapter)(nil)
	_ GetTenantsPort   = (*TenantPostgreAdapter)(nil)
)
