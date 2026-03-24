import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { PaginatedResponse } from '../../models/paginated-response.model';
import { TenantBackend } from '../../models/tenant/tenant-backend.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';

@Injectable({
  providedIn: 'root',
})
export class TenantApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  public getTenant(page: number, limit: number): Observable<PaginatedResponse<TenantBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedResponse<TenantBackend>>(`${this.apiUrl}/tenants`, { params });
  }

  public createTenant(config: TenantConfig): Observable<TenantBackend> {
    return this.http.post<TenantBackend>(`${this.apiUrl}/tenant`, config);
  }

  public deleteTenant(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/tenant/${id}`);
  }
}
