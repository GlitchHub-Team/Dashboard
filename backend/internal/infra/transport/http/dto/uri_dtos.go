package dto

type TenantUriDTO struct {
	TenantIdField
}

type TenantMemberUriDTO struct {
	TenantIdField
	UserIdField
}

type SuperAdminUriDTO struct {
	UserIdField
}
