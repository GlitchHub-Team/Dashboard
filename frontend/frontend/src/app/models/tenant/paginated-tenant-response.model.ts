import { PaginatedResponse } from '../paginated-response.model';

export interface PaginatedTenantResponse<T> extends PaginatedResponse {
  tenants: T[];
}
