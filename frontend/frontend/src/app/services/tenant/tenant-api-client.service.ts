import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';
import { environment } from '../../../environments/environment';

@Injectable({ providedIn: 'root' })
export class TenantApiClientService {
  private http = inject(HttpClient);
  private apiUrl = `${environment.apiUrl}/tenants`;

  getTenant(): Observable<Tenant[]> {
    return this.http.get<Tenant[]>(this.apiUrl);
  }

  createTenant(config: RawTenantConfig): Observable<Tenant> {
    return this.http.post<Tenant>(this.apiUrl, config);
  }

  deleteTenant(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}