import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedResponse } from '../models/paginated-response.model';

export abstract class GatewayAdapter {
  abstract fromDTO(dto: unknown, tenantId?: string): Gateway;
  abstract toDTO(gateway: Partial<Gateway>): unknown;
  abstract fromPaginatedDTO(
    response: PaginatedResponse<unknown>,
    tenantId?: string,
  ): PaginatedResponse<Gateway>;
}
