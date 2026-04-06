import { Injectable } from '@angular/core';
import { Observable, of, throwError, delay } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';

import { PaginatedTenantResponse } from '../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../models/tenant/tenant-backend.model';
import { TenantConfig } from '../models/tenant/tenant-config.model';

@Injectable({ providedIn: 'root' })
export class TenantApiClientMockService {
  private mockTenants: TenantBackend[] = [
    { tenant_id: 'tenant-1', tenant_name: 'Tenant 1', can_impersonate: true },
    { tenant_id: 'tenant-2', tenant_name: 'Tenant 2', can_impersonate: false },
    { tenant_id: 'tenant-3', tenant_name: 'Tenant 3', can_impersonate: true },
    { tenant_id: 'tenant-4', tenant_name: 'Tenant 4', can_impersonate: true },
    { tenant_id: 'tenant-5', tenant_name: 'Tenant 5', can_impersonate: false },
    { tenant_id: 'tenant-6', tenant_name: 'Tenant 6', can_impersonate: true },
  ];

  public getTenant(id: string): Observable<TenantBackend> {
    const shouldFail = false;

    if (shouldFail) {
      return throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            statusText: 'Bad Request',
            error: { error: 'tenant already exists' },
          }),
      ).pipe(delay(500));
    }
    const tenant = this.mockTenants.find((t) => t.tenant_id === id);
    return of(tenant!).pipe(delay(500));
  }

  public getTenants(page = 1, limit = 10): Observable<PaginatedTenantResponse<TenantBackend>> {
    const shouldFail = false;

    if (shouldFail) {
      return throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            statusText: 'Bad Request',
            error: { error: 'tenant already exists' },
          }),
      ).pipe(delay(500));
    }

    const total = this.mockTenants.length;
    const start = (page - 1) * limit;
    const tenants = this.mockTenants.slice(start, start + limit);
    return of({ count: tenants.length, total, tenants }).pipe(delay(500));
  }

  public createTenant(config: TenantConfig): Observable<TenantBackend> {
    const shouldFail = false;

    if (shouldFail) {
      return throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            statusText: 'Bad Request',
            error: { error: 'tenant already exists' },
          }),
      ).pipe(delay(500));
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
    const shouldFail = false;

    if (shouldFail) {
      return throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            statusText: 'Bad Request',
            error: { error: 'LOL' },
          }),
      ).pipe(delay(500));
    }

    this.mockTenants = this.mockTenants.filter((tenant) => tenant.tenant_id !== id);
    return of(void 0).pipe(delay(500));
  }
}