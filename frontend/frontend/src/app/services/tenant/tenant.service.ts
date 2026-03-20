import { Injectable, inject, signal } from '@angular/core';
import { Observable } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';
import { TenantApiClientService } from './tenant-api-client.service';

@Injectable({ providedIn: 'root' })
export class TenantService {
  private readonly tenantApiClient = inject(TenantApiClientService);

  public readonly loading = signal<boolean>(false);
  public readonly error = signal<string | null>(null);
  public readonly tenantList = signal<Tenant[]>([]);
  public readonly totalTenants = signal<number>(0);

  retrieveTenant(page = 0, size = 10): void {
    this.loading.set(true);
    this.error.set(null);

    this.tenantApiClient.getTenant(page, size).subscribe({
      next: (res) => {
        this.tenantList.set(res.items);
        this.totalTenants.set(res.totalCount);
        this.loading.set(false);
      },
      error: (err: Error) => {
        this.error.set(err.message || 'Failed to fetch tenants');
        this.loading.set(false);
      },
    });
  }

  addNewTenant(config: RawTenantConfig): Observable<Tenant> {
    this.loading.set(true);
    this.error.set(null);

    return new Observable((observer) => {
      this.tenantApiClient.createTenant(config).subscribe({
        next: (tenant: Tenant) => {
          this.tenantList.update((current) => [...current, tenant]);
          this.loading.set(false);
          observer.next(tenant);
          observer.complete();
        },
        error: (err: Error) => {
          this.error.set(err.message || 'Failed to create tenant');
          this.loading.set(false);
          observer.error(err);
        },
      });
    });
  }

  removeTenant(id: string): Observable<void> {
    this.loading.set(true);
    this.error.set(null);

    return new Observable((observer) => {
      this.tenantApiClient.deleteTenant(id).subscribe({
        next: () => {
          this.tenantList.update((current: Tenant[]) =>
            current.filter((t) => t.name !== id)
          );
          this.loading.set(false);
          observer.next();
          observer.complete();
        },
        error: (err: Error) => {
          this.error.set(err.message || 'Failed to delete tenant');
          this.loading.set(false);
          observer.error(err);
        },
      });
    });
  }
}