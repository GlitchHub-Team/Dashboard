import { Injectable } from '@angular/core';

import { GatewayAdapter } from './gateway.adapter';
import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedResponse } from '../models/paginated-response.model';
import { Status } from '../models/gateway-sensor-status.enum';

@Injectable()
export class GatewayApiAdapter extends GatewayAdapter {
  fromDTO(dto: GatewayBackend, tenantId?: string): Gateway {
    return {
      id: dto.gateway_id,
      name: dto.name,
      status: dto.status as Status,
      interval: dto.intervals,
      ...(tenantId && { tenantId }),
    };
  }

  fromPaginatedDTO(
    response: PaginatedResponse<GatewayBackend>,
    tenantId?: string,
  ): PaginatedResponse<Gateway> {
    return {
      count: response.count,
      total: response.total,
      data: response.data.map((dto) => this.fromDTO(dto, tenantId)),
    };
  }
}
