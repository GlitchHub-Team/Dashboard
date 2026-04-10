package tenant

import (
	"backend/internal/infra/transport/http/dto"
)

// Request DTO ========================================================================================

// Create --------------------------------------------------------------------------
type CreateTenantDTO struct {
	dto.TenantNameField
	CanImpersonate bool `json:"can_impersonate"`
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
	CanImpersonate bool `json:"can_impersonate"`
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

func NewTenantListResponseDTO(tenantList []Tenant, total uint) TenantListResponseDTO {
	var tenantDtos []TenantResponseDTO

	for _, t := range tenantList {
		tenantDtos = append(tenantDtos, NewTenantResponseDTO(t))
	}

	return TenantListResponseDTO{
		Tenants: tenantDtos,
		ListInfo: dto.ListInfo{
			Count: uint(len(tenantList)),
			Total: total,
		},
	}
}

type allTenants_SingleTenantResponseDTO struct {
	dto.TenantIdField
	dto.TenantNameField
}

type AllTenantsResponseDTO struct {
	Tenants []allTenants_SingleTenantResponseDTO `json:"tenants"`
}

func NewAllTenantsResponseDTO(tenantList []Tenant) AllTenantsResponseDTO {
	var tenantDtos []allTenants_SingleTenantResponseDTO
	for _, tenant := range tenantList {
		tenantDtos = append(tenantDtos, allTenants_SingleTenantResponseDTO{
			TenantIdField:   dto.TenantIdField{TenantId: tenant.Id.String()},
			TenantNameField: dto.TenantNameField{TenantName: tenant.Name},
		})
	}

	return AllTenantsResponseDTO{
		Tenants: tenantDtos,
	}
}
