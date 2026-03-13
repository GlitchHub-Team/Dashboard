package dto

type Pagination struct {
	Page  int `json:"page" binding:"required"`
	Limit int `json:"limit" binding:"required max=200"`
}

type ListInfo struct {
	Count int `json:"count" binding:"required"`
	Total int `json:"total" binding:"required"`
}

// User
type EmailField struct {
	Email string `json:"email" binding:"required email"`
}

type UsernameField struct {
	Username string `json:"username" binding:"required"`
}

type UserIdField struct {
	UserId int `json:"user_id" binding:"required"`
}

type UserRoleField struct {
	UserRole string `json:"user_role" binding:"required oneof=tenant_user tenant_admin super_admin"`
}

// Tenant
type TenantIdField struct {
	TenantId string `json:"tenant_id" binding:"required uuid4"`
}
