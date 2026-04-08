export interface TenantBackend {
  tenant_id: string;
  tenant_name: string;
  can_impersonate: boolean;
}
