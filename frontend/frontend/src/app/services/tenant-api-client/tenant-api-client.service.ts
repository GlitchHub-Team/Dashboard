import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { map, Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';
import { Tenant } from '../../models/tenant/tenant.model';
import { TenantApiAdapter } from '../../adapters/tenant/tenant-api.adapter';
import { TenantApiClientAdapter } from './tenant-api-client-adapter.service';

@Injectable({
  providedIn: 'root',
})
export class TenantApiClientService extends TenantApiClientAdapter {
  private readonly http = inject(HttpClient);
  private readonly mapper = inject(TenantApiAdapter);
  private readonly apiUrl = `${environment.apiUrl}`;

  public getTenant(id: string): Observable<Tenant> {
    return this.http
      .get<TenantBackend>(`${this.apiUrl}/tenant/${id}`)
      .pipe(map((dto) => this.mapper.fromDTO(dto)));
  }

  public getTenants(
    page: number,
    limit: number,
  ): Observable<PaginatedTenantResponse<Tenant>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http
      .get<PaginatedTenantResponse<TenantBackend>>(`${this.apiUrl}/tenants`, { params })
      .pipe(map((response) => this.mapper.fromPaginatedDTO(response)));
  }

  public getAllTenants(): Observable<Tenant[]> {
    return this.http
      .get<{ tenants: TenantBackend[] }>(`${this.apiUrl}/all_tenants`)
      .pipe(map((response) => response.tenants.map((dto) => this.mapper.fromDTO(dto))));
  }

  public createTenant(config: TenantConfig): Observable<Tenant> {
    return this.http
      .post<TenantBackend>(`${this.apiUrl}/tenant`, {
        tenant_name: config.name,
        can_impersonate: config.canImpersonate,
      })
      .pipe(map((dto) => this.mapper.fromDTO(dto)));
  }

  public deleteTenant(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/tenant/${id}`);
  }
}