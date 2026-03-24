package dto

type TenantIdField struct {
	TenantId string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"required,uuid4"`
}

type TenantIdField_NotRequired struct {
	TenantId *string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"excluded_if=UserRole super_admin,omitnil,uuid4"`
}
