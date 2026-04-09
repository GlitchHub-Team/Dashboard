package tenant

import (
	"github.com/google/uuid"
)

func TenantEntityToDomain(entity *TenantEntity) (tenant Tenant, err error) {
	if entity == nil || entity.ID == "" {
		return
	}

	tenantId, err := uuid.Parse(entity.ID)
	if err != nil {
		return
	}

	return Tenant{
		Id:             tenantId,
		Name:           entity.Name,
		CanImpersonate: entity.CanImpersonate,
	}, nil
}

func DomainToTenantEntity(tenant Tenant) (*TenantEntity, error) {
	return &TenantEntity{
		ID:             tenant.Id.String(),
		Name:           tenant.Name,
		CanImpersonate: tenant.CanImpersonate,
	}, nil
}
