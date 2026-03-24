export interface TenantBackend {
  tenant_id: string;
  name: string;
  can_impersonate: boolean;
}
