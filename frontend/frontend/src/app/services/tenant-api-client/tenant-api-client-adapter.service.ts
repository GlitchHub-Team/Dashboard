import { Observable } from 'rxjs';

import { Tenant } from '../../models/tenant/tenant.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';

export abstract class TenantApiClientAdapter {
  abstract getTenant(id: string): Observable<Tenant>;
  abstract getTenants(page: number, limit: number): Observable<PaginatedTenantResponse<Tenant>>;
  abstract getAllTenants(): Observable<Tenant[]>;
  abstract createTenant(config: TenantConfig): Observable<Tenant>;
  abstract deleteTenant(id: string): Observable<void>;
}