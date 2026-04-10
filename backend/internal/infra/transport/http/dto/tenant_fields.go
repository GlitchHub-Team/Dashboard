package dto

type TenantIdField struct {
	TenantId string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"required,uuid"`
}

type TenantIdField_NotRequired struct {
	TenantId *string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"excluded_if=UserRole super_admin,omitnil,uuid4"`
}

type TenantNameField struct {
	TenantName string `uri:"name" form:"name" json:"name" binding:"required"`
}
