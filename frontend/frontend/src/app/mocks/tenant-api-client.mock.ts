import { Injectable } from '@angular/core';
import { Observable, of, throwError, delay, switchMap, timer } from 'rxjs';

import { PaginatedTenantResponse } from '../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../models/tenant/tenant-backend.model';
import { TenantConfig } from '../models/tenant/tenant-config.model';
import { ApiError } from '../models/api-error.model';

@Injectable({ providedIn: 'root' })
export class TenantApiClientMockService {
  private readonly shouldFailGetTenant = false;
  private readonly shouldFailGetTenants = false;
  private readonly shouldFailCreateTenant = false;
  private readonly shouldFailDeleteTenant = false;

  private mockTenants: TenantBackend[] = [
    { tenant_id: 'tenant-1', tenant_name: 'Tenant 1', can_impersonate: true },
    { tenant_id: 'tenant-2', tenant_name: 'Tenant 2', can_impersonate: false },
    { tenant_id: 'tenant-3', tenant_name: 'Tenant 3', can_impersonate: true },
    { tenant_id: 'tenant-4', tenant_name: 'Tenant 4', can_impersonate: true },
    { tenant_id: 'tenant-5', tenant_name: 'Tenant 5', can_impersonate: false },
    { tenant_id: 'tenant-6', tenant_name: 'Tenant 6', can_impersonate: true },
    { tenant_id: 'tenant-7', tenant_name: 'Tenant 7', can_impersonate: true },
    { tenant_id: 'tenant-8', tenant_name: 'Tenant 8', can_impersonate: false },
    { tenant_id: 'tenant-9', tenant_name: 'Tenant 9', can_impersonate: true },
    { tenant_id: 'tenant-10', tenant_name: 'Tenant 10', can_impersonate: true },
    { tenant_id: 'tenant-11', tenant_name: 'Tenant 11', can_impersonate: false },
    { tenant_id: 'tenant-12', tenant_name: 'Tenant 12', can_impersonate: true },
  ];

  public getTenant(id: string): Observable<TenantBackend> {
    if (this.shouldFailGetTenant) {
      return this.delayedError(400, 'Failed to fetch tenant');
    }

    const tenant = this.mockTenants.find((t) => t.tenant_id === id);
    return of(tenant!).pipe(delay(500));
  }

  public getTenants(page = 1, limit = 10): Observable<PaginatedTenantResponse<TenantBackend>> {
    if (this.shouldFailGetTenants) {
      return this.delayedError(400, 'Failed to fetch tenants');
    }

    const total = this.mockTenants.length;
    const start = (page - 1) * limit;
    const tenants = this.mockTenants.slice(start, start + limit);
    return of({ count: tenants.length, total, tenants }).pipe(delay(500));
  }

  public createTenant(config: TenantConfig): Observable<TenantBackend> {
    if (this.shouldFailCreateTenant) {
      return this.delayedError(400, 'Failed to create tenant');
    }

    const newTenant: TenantBackend = {
      tenant_id: `tenant-${(this.mockTenants.length + 1).toString()}`,
      tenant_name: config.name,
      can_impersonate: config.canImpersonate,
    };
    this.mockTenants.push(newTenant);
    return of(newTenant).pipe(delay(500));
  }

  public deleteTenant(id: string): Observable<void> {
    if (this.shouldFailDeleteTenant) {
      return this.delayedError(400, 'Failed to delete tenant');
    }

    this.mockTenants = this.mockTenants.filter((tenant) => tenant.tenant_id !== id);
    return of(void 0).pipe(delay(500));
  }

  private delayedError(status: number, message: string): Observable<never> {
    return timer(500).pipe(switchMap(() => throwError(() => ({ status, message }) as ApiError)));
  }
}
