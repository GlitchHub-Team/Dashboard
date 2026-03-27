import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { PaginatedTenantResponse } from '../../models/tenant/paginated-tenant-response.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';

@Injectable({
  providedIn: 'root',
})
export class TenantApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  public getTenant(id: string): Observable<TenantBackend> {
    return this.http.get<TenantBackend>(`${this.apiUrl}/tenant/${id}`);
  }

  public getTenants(
    page: number,
    limit: number,
  ): Observable<PaginatedTenantResponse<TenantBackend>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http.get<PaginatedTenantResponse<TenantBackend>>(`${this.apiUrl}/tenants`, {
      params,
    });
  }

  public createTenant(config: TenantConfig): Observable<TenantBackend> {
    return this.http.post<TenantBackend>(`${this.apiUrl}/tenant`, config);
  }

  public deleteTenant(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/tenant/${id}`);
  }
}
