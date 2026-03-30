import { Injectable } from '@angular/core';

import { GatewayAdapter } from './gateway.adapter';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { Gateway } from '../../models/gateway/gateway.model';
import { PaginatedGatewayResponse } from '../../models/gateway/paginated-gateway-response.model';
import { statusMapper } from '../../utils/status.utils';

@Injectable()
export class GatewayApiAdapter extends GatewayAdapter {
  fromDTO(dto: GatewayBackend): Gateway {
    return {
      id: dto.gateway_id,
      name: dto.name,
      status: statusMapper.fromBackend(dto.status),
      interval: dto.interval,
      tenantId: dto.tenant_id ?? undefined,
    };
  }

  fromPaginatedDTO(
    response: PaginatedGatewayResponse<GatewayBackend>,
  ): PaginatedGatewayResponse<Gateway> {
    return {
      count: response.count,
      total: response.total,
      gateways: response.gateways.map((dto) => this.fromDTO(dto)),
    };
  }
}
