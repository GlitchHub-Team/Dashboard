import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedGatewayResponse } from '../models/gateway/paginated-gateway-response.model';

export abstract class GatewayAdapter {
  abstract fromDTO(dto: unknown): Gateway;
  abstract fromPaginatedDTO(
    response: PaginatedGatewayResponse<unknown>,
  ): PaginatedGatewayResponse<Gateway>;
}
