import { Tenant } from '../models/tenant/tenant.model';
import { PaginatedResponse } from '../models/paginated-response.model';

export abstract class TenantAdapter {
  abstract fromDTO(dto: unknown): Tenant;
  abstract fromPaginatedDTO(response: PaginatedResponse<unknown>): PaginatedResponse<Tenant>;
}
