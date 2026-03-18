// adapters/gateway-api.adapter.ts
import { Injectable } from '@angular/core';
import { GatewayAdapter } from './gateway.adapter';
import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { Gateway } from '../models/gateway/gateway.model';
import { PaginatedResponse } from '../models/paginated-response.model';

@Injectable()
export class GatewayApiAdapter extends GatewayAdapter {
  fromDTO(dto: GatewayBackend, tenantId?: string): Gateway {
    return {
      id: dto.GatewayId,
      name: dto.GatewayName,
      ...(tenantId && { tenantId }),
    };
  }

  toDTO(gateway: Partial<Gateway>): Partial<GatewayBackend> {
    return {
      GatewayName: gateway.name,
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
