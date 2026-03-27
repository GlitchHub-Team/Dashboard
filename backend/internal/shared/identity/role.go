package identity

type UserRole string

const (
	ROLE_TENANT_USER  UserRole = "tenant_user"
	ROLE_TENANT_ADMIN UserRole = "tenant_admin"
	ROLE_SUPER_ADMIN  UserRole = "super_admin"
)
