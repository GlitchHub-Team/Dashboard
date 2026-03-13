import { Injectable, inject } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';
import { TenantApiClientService } from './tenant-api-client.service';

@Injectable({ providedIn: 'root' })
export class TenantService {
  private tenantApiClient = inject(TenantApiClientService);

  loading$ = new BehaviorSubject<boolean>(false);
  error$ = new BehaviorSubject<string | null>(null);
  tenantList$ = new BehaviorSubject<Tenant[]>([]);

  retrieveTenant(): void {
    this.loading$.next(true);
    this.error$.next(null);

    this.tenantApiClient.getTenant().subscribe({
      next: (tenants: Tenant[]) => {
        this.tenantList$.next(tenants);
        this.loading$.next(false);
      },
      error: (err: any) => {
        this.error$.next(err.message || 'Failed to fetch tenants');
        this.loading$.next(false);
      },
    });
  }

  addNewTenant(config: RawTenantConfig): Observable<Tenant> {
    this.loading$.next(true);
    this.error$.next(null);

    return new Observable((observer) => {
      this.tenantApiClient.createTenant(config).subscribe({
        next: (tenant: Tenant) => {
          const currentList = this.tenantList$.value;
          this.tenantList$.next([...currentList, tenant]);
          this.loading$.next(false);
          observer.next(tenant);
          observer.complete();
        },
        error: (err: any) => {
          this.error$.next(err.message || 'Failed to create tenant');
          this.loading$.next(false);
          observer.error(err);
        },
      });
    });
  }

  removeTenant(id: string): Observable<void> {
    this.loading$.next(true);
    this.error$.next(null);

    return new Observable((observer) => {
      this.tenantApiClient.deleteTenant(id).subscribe({
        next: () => {
          const currentList = this.tenantList$.value;
          this.tenantList$.next(currentList.filter((t) => t.name !== id));
          this.loading$.next(false);
          observer.next();
          observer.complete();
        },
        error: (err: any) => {
          this.error$.next(err.message || 'Failed to delete tenant');
          this.loading$.next(false);
          observer.error(err);
        },
      });
    });
  }
}