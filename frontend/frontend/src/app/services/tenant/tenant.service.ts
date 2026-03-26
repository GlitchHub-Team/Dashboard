import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, EMPTY, finalize, map } from 'rxjs';

import { TenantApiAdapter } from '../../adapters/tenant-api.adapter';
import { ApiError } from '../../models/api-error.model';
import { Tenant } from '../../models/tenant/tenant.model';
import { TenantConfig } from '../../models/tenant/tenant-config.model';
import { TenantApiClientService } from '../tenant-api-client/tenant-api-client.service';

@Injectable({
  providedIn: 'root',
})
export class TenantService {
  private readonly tenantApi = inject(TenantApiClientService);
  private readonly adapter = inject(TenantApiAdapter);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _tenantList = signal<Tenant[]>([]);
  private readonly _total = signal(0);
  private readonly _pageIndex = signal(0);
  private readonly _limit = signal(10);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly tenantList = this._tenantList.asReadonly();
  public readonly total = this._total.asReadonly();
  public readonly pageIndex = this._pageIndex.asReadonly();
  public readonly limit = this._limit.asReadonly();

  public getTenant(id: string): Observable<Tenant> {
    this.setLoadingState();

    return this.tenantApi.getTenant(id).pipe(
      map((dto) => this.adapter.fromDTO(dto)),
      tap({
        error: (err: ApiError) => {
          this._error.set(err.message ?? 'Failed to fetch tenant');
        },
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public retrieveTenants(): void {
    this.setGettingTenantsState();

    this.tenantApi
      .getTenants(this.pageIndex(), this.limit())
      .pipe(
        map((response) => this.adapter.fromPaginatedDTO(response)),
        tap((result) => {
          this._tenantList.set(result.tenants);
          this._total.set(result.total);
        }),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to fetch tenants');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public changePage(pageIndex: number, limit: number): void {
    this._pageIndex.set(pageIndex);
    this._limit.set(limit);
    this.retrieveTenants();
  }

  public addNewTenant(config: TenantConfig): Observable<Tenant> {
    this.setLoadingState();

    return this.tenantApi.createTenant(config).pipe(
      map((dto) => this.adapter.fromDTO(dto)),
      tap({
        error: (err: ApiError) => {
          this._error.set(err.message ?? 'Failed to create tenant');
        },
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public removeTenant(id: string): Observable<void> {
    this.setLoadingState();

    return this.tenantApi.deleteTenant(id).pipe(
      tap(() => this.refetchCurrentPage()),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to delete tenant');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  private refetchCurrentPage(): void {
    this.changePage(this._pageIndex(), this._limit());
  }

  private setGettingTenantsState(): void {
    this._tenantList.set([]);
    this._loading.set(true);
    this._error.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }
}
