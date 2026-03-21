import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';
import { environment } from '../../../environments/environment';
import { TenantDataAdapter, RawPaginatedTenantResponse } from '../../adapters/tenant-data.adapter';

export interface TenantConfig {
  name: string;
}

@Injectable({ providedIn: 'root' })
export class TenantApiClientService {
  private readonly http = inject(HttpClient);
  private readonly tenantAdapter = inject(TenantDataAdapter);
  private readonly apiUrl = `${environment.apiUrl}/tenants`;

  public getTenant(page = 0, size = 10): Observable<{ items: Tenant[]; totalCount: number }> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('size', size.toString());

    return this.http.get<RawPaginatedTenantResponse>(this.apiUrl, { params }).pipe(
      map(response => this.tenantAdapter.adaptPaginated(response))
    );
  }

  public createTenant(config: TenantConfig): Observable<Tenant> {
    return this.http.post<RawTenantConfig>(this.apiUrl, config).pipe(
      map(data => this.tenantAdapter.adapt(data))
    );
  }

  public deleteTenant(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}