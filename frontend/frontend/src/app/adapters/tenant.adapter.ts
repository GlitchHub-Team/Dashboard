import { Tenant } from '../models/tenant/tenant.model';
import { PaginatedTenantResponse } from '../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../models/tenant/tenant-backend.model';

export abstract class TenantAdapter {
  abstract fromDTO(dto: unknown): Tenant;
  abstract fromPaginatedDTO(
    response: PaginatedTenantResponse<TenantBackend>,
  ): PaginatedTenantResponse<Tenant>;
}
