import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';

@Injectable({ providedIn: 'root' })
export class TenantApiClientMockService {
  private mockTenants: Tenant[] = [
    { name: 'Tenant 1' },
    { name: 'Tenant 2' },
    { name: 'Tenant 3' },
  ];

  public getTenant(): Observable<Tenant[]> {
    return of([...this.mockTenants]).pipe(delay(500));
  }

  public createTenant(config: RawTenantConfig): Observable<Tenant> {
    const newTenant: Tenant = { name: config.name };
    this.mockTenants.push(newTenant);
    return of(newTenant).pipe(delay(500));
  }

  public deleteTenant(name: string): Observable<void> {
    this.mockTenants = this.mockTenants.filter((t) => t.name !== name);
    return of(void 0).pipe(delay(500));
  }
}