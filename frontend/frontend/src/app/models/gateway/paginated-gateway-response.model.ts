import { PaginatedResponse } from '../paginated-response.model';

export interface PaginatedGatewayResponse<T> extends PaginatedResponse {
  gateways: T[];
}
