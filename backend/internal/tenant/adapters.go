package tenant

import (
	"errors"

	"backend/internal/infra/database"
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockgen -source=adapters.go -destination=../../tests/tenant/mocks/repository_adapter.go -package=mocks TenantRepository

type TenantRepository interface {
	SaveTenant(tenant *TenantEntity) error
	GetTenant(tenantId string) (*TenantEntity, error)
	DeleteTenant(tenant *TenantEntity) error
	GetTenants(offset, limit int) ([]TenantEntity, int64, error)
	GetAllTenants() ([]TenantEntity, error)
}

type TenantPostgreAdapter struct {
	repo TenantRepository
}

var (
	_ TenantRepository = (*TenantPostgreRepository)(nil)
	_ CreateTenantPort = (*TenantPostgreAdapter)(nil)
	_ DeleteTenantPort = (*TenantPostgreAdapter)(nil)
	_ GetTenantPort    = (*TenantPostgreAdapter)(nil)
	_ GetTenantsPort   = (*TenantPostgreAdapter)(nil)
)

func NewTenantPostgreAdapter(repository TenantRepository) *TenantPostgreAdapter {
	return &TenantPostgreAdapter{
		repo: repository,
	}
}

// CREATE =============================================================================================

func (adapter *TenantPostgreAdapter) CreateTenant(tenant Tenant) (Tenant, error) {
	entity, _ := DomainToTenantEntity(tenant)

	err := adapter.repo.SaveTenant(entity)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return Tenant{}, ErrTenantAlreadyExists
		}
		return Tenant{}, err
	}

	savedTenant, err := TenantEntityToDomain(entity)
	return savedTenant, err
}

// DELETE =============================================================================================

func (adapter *TenantPostgreAdapter) DeleteTenant(tenantId uuid.UUID) (Tenant, error) {
	oldEntity := TenantEntity{
		ID: tenantId.String(),
	}

	err := adapter.repo.DeleteTenant(&oldEntity)
	if err != nil {
		return Tenant{}, err
	}

	oldTenant, err := TenantEntityToDomain(&oldEntity)
	return oldTenant, err
}

// GET ================================================================================================

func (adapter *TenantPostgreAdapter) GetTenant(tenantId uuid.UUID) (Tenant, error) {
	entity, err := adapter.repo.GetTenant(tenantId.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Tenant{}, ErrTenantNotFound
		}
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
