import { Injectable } from '@angular/core';
import { TenantAdapter } from './tenant.adapter';
import { TenantBackend } from '../models/tenant/tenant-backend.model';
import { Tenant } from '../models/tenant/tenant.model';
import { PaginatedResponse } from '../models/paginated-response.model';

@Injectable({ providedIn: 'root' })
export class TenantApiAdapter extends TenantAdapter {
  fromDTO(dto: TenantBackend): Tenant {
    return {
      id: dto.tenant_id,
      name: dto.name,
      canImpersonate: dto.can_impersonate,
    };
  }

  fromPaginatedDTO(response: PaginatedResponse<TenantBackend>): PaginatedResponse<Tenant> {
    return {
      count: response.count,
      total: response.total,
      data: response.data.map((dto) => this.fromDTO(dto)),
    };
  }
}
