package tenant

import (
	"errors"

	"backend/internal/infra/database"
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
)

// Metodi struct ===============================================================================================

func (TenantEntity) TableName() string { return "tenants" }

type TenantPostgreAdapter struct {
	repo *TenantPostgreRepository
}

// Adapter ============================================================================================

//go:generate mockgen -destination=../../tests/tenant/mocks/ports.go -package=mocks . CreateTenantPort,DeleteTenantPort,GetTenantPort,GetTenantsPort,GetTenantByUserPort

func NewTenantPostgreAdapter(repository *TenantPostgreRepository) (
	CreateTenantPort,
	DeleteTenantPort,
	GetTenantPort,
	GetTenantsPort,
	GetTenantByIdPort,
) {
	adapter := &TenantPostgreAdapter{repo: repository}
	return adapter, adapter, adapter, adapter, adapter
}

// CREATE =============================================================================================

func (adapter *TenantPostgreAdapter) CreateTenant(tenant Tenant) (Tenant, error) {
	entity, _ := DomainToTenantEntity(tenant)

	err := adapter.repo.SaveTenant(entity)
	if err != nil {
		return Tenant{}, err
	}

	savedTenant, err := TenantEntityToDomain(entity)
	return savedTenant, err
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

	err = adapter.repo.DeleteTenant(oldEntity)
	if err != nil {
		return Tenant{}, err
	}

	oldTenant, err := TenantEntityToDomain(oldEntity)
	return oldTenant, err
}

// GET ================================================================================================

type GetTenantPort interface {
	GetTenant(tenantId uuid.UUID) (Tenant, error)
}

type GetTenantsPort interface {
	GetTenants(page, limit int) ([]Tenant, uint, error)
	GetAllTenants() ([]Tenant, error)
}

type GetTenantByIdPort interface {
	GetTenantById(tenantId uuid.UUID) (Tenant, error)
}

func (adapter *TenantPostgreAdapter) GetTenant(tenantId uuid.UUID) (Tenant, error) {
	entity, err := adapter.repo.GetTenant(tenantId.String())
	if err != nil {
		return Tenant{}, err
	}

	tenant, err := TenantEntityToDomain(entity)
	return tenant, err
}

func (adapter *TenantPostgreAdapter) GetTenants(page, limit int) ([]Tenant, uint, error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, total, err := adapter.repo.GetTenants(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	var tenants []Tenant
	tenants, err = database.MapEntityListToDomain(entities, TenantEntityToDomain)
	return tenants, uint(total), err
}

func (adapter *TenantPostgreAdapter) GetAllTenants() ([]Tenant, error) {
	entities, err := adapter.repo.GetAllTenants()
	if err != nil {
		return nil, err
	}

	var tenants []Tenant
	tenants, err = database.MapEntityListToDomain(entities, TenantEntityToDomain)
	return tenants, err
}

func (adapter *TenantPostgreAdapter) GetTenantById(tenantId uuid.UUID) (Tenant, error) {
	entity, err := adapter.repo.GetTenantById(tenantId.String())
	if err != nil {
		return Tenant{}, err
	}

	tenant, err := TenantEntityToDomain(entity)
	return tenant, err
}

// Compile-time checks ================================================================================
var (
	_ CreateTenantPort = (*TenantPostgreAdapter)(nil)
	_ DeleteTenantPort = (*TenantPostgreAdapter)(nil)
	_ GetTenantPort    = (*TenantPostgreAdapter)(nil)
	_ GetTenantsPort   = (*TenantPostgreAdapter)(nil)
	_ GetTenantByIdPort = (*TenantPostgreAdapter)(nil)
)
