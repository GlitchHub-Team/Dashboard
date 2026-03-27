import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedGatewayResponse } from '../models/gateway/paginated-gateway-response.model';

export abstract class GatewayAdapter {
  abstract fromDTO(dto: GatewayBackend): Gateway;
  abstract fromPaginatedDTO(
    response: PaginatedGatewayResponse<GatewayBackend>,
  ): PaginatedGatewayResponse<Gateway>;
}
