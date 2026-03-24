package dto

type Pagination struct {
	// Page: pagina di dati (iniziando a contare da 1)
	Page int `uri:"page" form:"page" json:"page" binding:"min=1"`

	// Limit: quanti elementi inserire in una pagina (minimo: 10, massimo: 200)
	Limit int `uri:"limit" form:"limit" json:"limit" binding:"min=10,max=200"`
}

var DEFAULT_PAGINATION = Pagination{Page: 1, Limit: 25}

type ListInfo struct {
	Count uint `uri:"count" form:"count" json:"count" binding:"required"`
	Total uint `uri:"total" form:"total" json:"total" binding:"required"`
}

// User
type EmailField struct {
	Email string `uri:"email" form:"email" json:"email" binding:"required,email"`
}

type UsernameField struct {
	Username string `uri:"username" form:"username" json:"username" binding:"required"`
}

type UserIdField struct {
	UserId uint `uri:"user_id" form:"user_id" json:"user_id" binding:"required"`
}

type UserRoleField struct {
	UserRole string `uri:"user_role" form:"user_role" json:"user_role" binding:"required,oneof=tenant_user tenant_admin super_admin"`
}

// Tenant
type TenantIdField struct {
	TenantId string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"uuid4,required"`
}

// Uri DTOs -------------------------------------------------------------------------------------------
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
