import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';
import { environment } from '../../../environments/environment';
import { TenantDataAdapter } from './tenant-data.adapter';

export interface TenantConfig {
  name: string;
}

@Injectable({ providedIn: 'root' })
export class TenantApiClientService {
  private readonly http = inject(HttpClient);
  private readonly tenantAdapter = inject(TenantDataAdapter);
  private readonly apiUrl = `${environment.apiUrl}/tenants`;

  public getTenant(): Observable<Tenant[]> {
    return this.http.get<RawTenantConfig[]>(this.apiUrl).pipe(
      map(data => this.tenantAdapter.adaptArray(data))
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