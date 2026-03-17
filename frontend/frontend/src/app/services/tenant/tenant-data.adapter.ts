import { Injectable } from '@angular/core';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';

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
}