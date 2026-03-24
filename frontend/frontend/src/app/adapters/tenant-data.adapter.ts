import { Injectable } from '@angular/core';
import { Tenant } from '../models/tenant.model';
import { RawTenantConfig } from '../models/raw-tenant-config.model';

export interface RawPaginatedTenantResponse {
  content?: RawTenantConfig[];
  items?: RawTenantConfig[];
  data?: RawTenantConfig[];
  totalElements?: number;
  totalCount?: number;
  total?: number;
}

@Injectable({ providedIn: 'root' })
export class TenantDataAdapter {
  public adapt(input: RawTenantConfig): Tenant {
    return {
      name: input.name || '',
    };
  }

  public adaptArray(items: RawTenantConfig[]): Tenant[] {
    return items ? items.map(item => this.adapt(item)) : [];
  }

  public adaptPaginated(response: RawPaginatedTenantResponse): { items: Tenant[]; totalCount: number } {
    return {
      items: this.adaptArray(response.content || response.items || response.data || []),
      totalCount: response.totalElements || response.totalCount || response.total || 0,
    };
  }
}