package dto

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
