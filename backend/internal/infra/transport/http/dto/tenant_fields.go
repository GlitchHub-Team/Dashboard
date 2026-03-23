package dto

type TenantIdField struct {
	TenantId string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"uuid4,required"`
}

type TenantIdField_NotRequired struct {
	TenantId string `uri:"tenant_id" form:"tenant_id" json:"tenant_id" binding:"uuid4"`
}