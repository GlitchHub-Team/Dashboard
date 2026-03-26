package dto

type Pagination struct {
	Page  int `uri:"page" form:"page" json:"page" binding:"required"`
	Limit int `uri:"limit" form:"limit" json:"limit" binding:"required,max=200"`
}

type ListInfo struct {
	Count int `uri:"count" form:"count" json:"count" binding:"required"`
	Total int `uri:"total" form:"total" json:"total" binding:"required"`
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
	TenantId string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"required"`
}

type TenantNameField struct {
	TenantName string `uri:"tenant_name" form:"tenant_name" json:"tenant_name" binding:"required"`
}

// Gateway
type GatewayIdField struct {
	GatewayId string `uri:"gateway_id" form:"gateway_id" json:"gateway_id" binding:"uuid4,required"`
}

type GatewayNameField struct {
	GatewayName string `uri:"gateway_name" form:"gateway_name" json:"gateway_name" binding:"required"`
}

type GatewayCertificateField struct {
	Certificate string `uri:"certificate" form:"certificate" json:"certificate" binding:"required"`
}
