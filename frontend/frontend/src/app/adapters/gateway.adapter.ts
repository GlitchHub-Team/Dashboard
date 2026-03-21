import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedResponse } from '../models/paginated-response.model';

export abstract class GatewayAdapter {
  abstract fromDTO(dto: unknown): Gateway;
  abstract fromPaginatedDTO(response: PaginatedResponse<unknown>): PaginatedResponse<Gateway>;
}
