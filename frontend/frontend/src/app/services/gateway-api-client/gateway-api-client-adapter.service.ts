import { Observable } from 'rxjs';

import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../../models/gateway/paginated-gateway-response.model';

export abstract class GatewayApiClientAdapter {
  abstract getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<Gateway>>;

  abstract getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<Gateway>>;

  abstract addNewGateway(config: GatewayConfig): Observable<Gateway>;

  abstract deleteGateway(gatewayId: string): Observable<void>;
}