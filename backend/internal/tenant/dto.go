package tenant

import (
	"backend/internal/common/dto"
)

// Request DTO ========================================================================================

// Create --------------------------------------------------------------------------
type CreateTenantDTO struct {
	dto.TenantNameField
	CanImpersonate bool `json:"canimpersonate"`
}

// Delete --------------------------------------------------------------------------
type DeleteTenantDTO struct {
	dto.TenantIdField
}

// Get ------------------------------------------------------------------------------
type GetTenantDTO struct {
	dto.TenantIdField
}

type GetTenantListDTO struct {
	dto.Pagination
}

type GetTenantByUserDTO struct {
	dto.UserIdField // Ora richiede obbligatoriamente un user_id valido
}

// Response DTO =======================================================================================

type TenantResponseDTO struct {
	dto.TenantIdField
	dto.TenantNameField
	CanImpersonate bool `json:"canimpersonate"`
}

func NewTenantResponseDTO(tenant Tenant) TenantResponseDTO {
	return TenantResponseDTO{
		TenantIdField:   dto.TenantIdField{TenantId: tenant.Id.String()},
		TenantNameField: dto.TenantNameField{TenantName: tenant.Name},
		CanImpersonate:  tenant.CanImpersonate,
	}
}

type TenantListResponseDTO struct {
	dto.ListInfo
	Tenants []TenantResponseDTO `json:"tenants"`
}

func NewTenantListResponseDTO(tenantList []Tenant, total int) TenantListResponseDTO {
	var tenantDtos []TenantResponseDTO

	for _, t := range tenantList {
		tenantDtos = append(tenantDtos, NewTenantResponseDTO(t))
	}

	return TenantListResponseDTO{
		Tenants: tenantDtos,
		ListInfo: dto.ListInfo{
			Count: len(tenantList),
			Total: total,
		},
	}
}
