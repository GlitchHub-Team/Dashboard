import { Injectable } from '@angular/core';
import { TenantAdapter } from './tenant.adapter';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { Tenant } from '../../models/tenant/tenant.model';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';

@Injectable({ providedIn: 'root' })
export class TenantApiAdapter extends TenantAdapter {
  fromDTO(dto: TenantBackend): Tenant {
    return {
      id: dto.tenant_id,
      name: dto.name,
      canImpersonate: dto.can_impersonate,
    };
  }

  fromPaginatedDTO(
    response: PaginatedTenantResponse<TenantBackend>,
  ): PaginatedTenantResponse<Tenant> {
    return {
      count: response.count,
      total: response.total,
      tenants: response.tenants.map((dto) => this.fromDTO(dto)),
    };
  }
}
