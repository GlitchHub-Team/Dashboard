package tenant

import (
	"backend/internal/common/dto"
)

type CreateTenantDTO struct {
	dto.TenantNameField
	CanImpersonate bool `json:"canimpersonate"`
}

type DeleteTenantDTO struct {
	dto.TenantIdField
}

type GetTenantDTO struct {
	dto.TenantIdField
}

type GetTenantListDTO struct {
	dto.Pagination
}

type GetTenantByUserDTO struct {
	dto.TenantIdField
}

type TenantResponseDTO struct {
	dto.TenantIdField
}

type TenantResponseListDTO struct {
	Count   int      `form:"tenant_name" json:"tenant_name" binding:"required"`
	Total   int      `form:"total" json:"total" binding:"required"`
	Tenants []Tenant ` form:"tenants" json:"tenants" binding:"required"`
}
