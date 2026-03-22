import { Injectable } from '@angular/core';

import { GatewayAdapter } from './gateway.adapter';
import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { Status } from '../models/gateway-sensor-status.enum';

@Injectable()
export class GatewayApiAdapter extends GatewayAdapter {
  fromDTO(dto: GatewayBackend): Gateway {
    return {
      id: dto.gateway_id,
      name: dto.name,
      status: dto.status as Status,
      interval: dto.interval,
      tenantId: dto.tenant_id ?? undefined,
    };
  }

  fromPaginatedDTO(response: PaginatedResponse<GatewayBackend>): PaginatedResponse<Gateway> {
    return {
      count: response.count,
      total: response.total,
      data: response.data.map((dto) => this.fromDTO(dto)),
    };
  }
}
