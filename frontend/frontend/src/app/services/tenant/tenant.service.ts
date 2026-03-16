import { Injectable, inject, signal } from '@angular/core';
import { Observable } from 'rxjs';
import { Tenant } from '../../models/tenant.model';
import { RawTenantConfig } from '../../models/raw-tenant-config.model';
import { TenantApiClientService } from './tenant-api-client.service';

@Injectable({ providedIn: 'root' })
export class TenantService {
  private tenantApiClient = inject(TenantApiClientService);

  loading = signal<boolean>(false);
  error = signal<string | null>(null);
  tenantList = signal<Tenant[]>([]);

  retrieveTenant(): void {
    this.loading.set(true);
    this.error.set(null);

    this.tenantApiClient.getTenant().subscribe({
      next: (tenants: Tenant[]) => {
        this.tenantList.set(tenants);
        this.loading.set(false);
      },
      error: (err: any) => {
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
        error: (err: any) => {
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
        error: (err: any) => {
          this.error.set(err.message || 'Failed to delete tenant');
          this.loading.set(false);
          observer.error(err);
        },
      });
    });
  }
}