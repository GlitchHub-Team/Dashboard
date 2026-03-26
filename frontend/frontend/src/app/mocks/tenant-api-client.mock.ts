import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';

import { PaginatedResponse } from '../models/paginated-response.model';
import { TenantBackend } from '../models/tenant/tenant-backend.model';
import { TenantConfig } from '../models/tenant/tenant-config.model';

@Injectable({ providedIn: 'root' })
export class TenantApiClientMockService {
  private mockTenants: TenantBackend[] = [
    { tenant_id: 'tenant-1', name: 'Tenant 1', can_impersonate: true },
    { tenant_id: 'tenant-2', name: 'Tenant 2', can_impersonate: false },
    { tenant_id: 'tenant-3', name: 'Tenant 3', can_impersonate: true },
    { tenant_id: 'tenant-4', name: 'Tenant 4', can_impersonate: true },
    { tenant_id: 'tenant-5', name: 'Tenant 5', can_impersonate: false },
    { tenant_id: 'tenant-6', name: 'Tenant 6', can_impersonate: true },
  ];

  public getTenant(id: string): Observable<TenantBackend> {
    const tenant = this.mockTenants.find((t) => t.tenant_id === id);
    return of(tenant!).pipe(delay(500));
  }

  public getTenants(page = 0, limit = 10): Observable<PaginatedResponse<TenantBackend>> {
    const total = this.mockTenants.length;
    const data = this.mockTenants.slice(page * limit, (page + 1) * limit);
    return of({ count: data.length, total, data }).pipe(delay(500));
  }

  public createTenant(config: TenantConfig): Observable<TenantBackend> {
    const newTenant: TenantBackend = {
      tenant_id: `tenant-${(this.mockTenants.length + 1).toString()}`,
      name: config.name,
      can_impersonate: config.canImpersonate,
    };
    this.mockTenants.push(newTenant);
    return of(newTenant).pipe(delay(500));
  }

  public deleteTenant(id: string): Observable<void> {
    this.mockTenants = this.mockTenants.filter((tenant) => tenant.tenant_id !== id);
    return of(void 0).pipe(delay(500));
  }
}
