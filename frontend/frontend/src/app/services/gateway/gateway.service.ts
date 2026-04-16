import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, EMPTY, finalize } from 'rxjs';

import { GatewayApiClientAdapter } from '../gateway-api-client/gateway-api-client-adapter.service';
import { GatewayCommandApiClientAdapter } from '../gateway-command-api-client/gateway-command-api-client-adapter.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayService {
  private readonly gatewayApi = inject(GatewayApiClientAdapter);
  private readonly gatewayCommandApi = inject(GatewayCommandApiClientAdapter);

  private readonly _gatewayList = signal<Gateway[]>([]);
  private readonly _total = signal(0);
  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _pageIndex = signal(0);
  private readonly _limit = signal(10);
  private readonly _currentTenantId = signal<string | null>(null);

  public readonly gatewayList = this._gatewayList.asReadonly();
  public readonly total = this._total.asReadonly();
  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly pageIndex = this._pageIndex.asReadonly();
  public readonly limit = this._limit.asReadonly();

  public getGatewaysByTenant(tenantId: string, page: number, limit: number): void {
    this._currentTenantId.set(tenantId);
    this._pageIndex.set(page);
    this._limit.set(limit);
    this.setGettingGatewaysState();

    this.gatewayApi
      .getGatewayListByTenant(tenantId, page + 1, limit)
      .pipe(
        tap((result) => {
          this._gatewayList.set(result.gateways);
          this._total.set(result.total);
        }),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load gateways');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public getGateways(page: number, limit: number): void {
    this._currentTenantId.set(null);
    this._pageIndex.set(page);
    this._limit.set(limit);
    this.setGettingGatewaysState();

    this.gatewayApi
      .getGatewayList(page + 1, limit)
      .pipe(
        tap((result) => {
          this._gatewayList.set(result.gateways);
          this._total.set(result.total);
        }),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load gateways');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public addNewGateway(config: GatewayConfig): Observable<Gateway> {
    return this.gatewayApi.addNewGateway(config);
  }

  public deleteGateway(id: string): Observable<void> {
    this.setLoadingState();

    return this.gatewayApi.deleteGateway(id).pipe(
      tap(() => this.refetchCurrentPage()),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to delete gateway');
        this._loading.set(false);
        return EMPTY;
      }),
    );
  }

  public commissionGateway(id: string, tenantId: string, token: string): Observable<Gateway> {
    return this.gatewayCommandApi.commissionGateway(id, tenantId, token).pipe(
      tap(() => this.refetchCurrentPage()),
    );
  }

  public decommissionGateway(id: string): Observable<void> {
    return this.gatewayCommandApi.decommissionGateway(id).pipe(
      tap(() => this.refetchCurrentPage()),
    );
  }

  public resetGateway(id: string): Observable<void> {
    return this.gatewayCommandApi.resetGateway(id);
  }

  public rebootGateway(id: string): Observable<void> {
    return this.gatewayCommandApi.rebootGateway(id);
  }

  public interruptGateway(id: string): Observable<void> {
    return this.gatewayCommandApi.interruptGateway(id).pipe(
      tap(() => this.refetchCurrentPage()),
    );
  }

  public resumeGateway(id: string): Observable<void> {
    return this.gatewayCommandApi.resumeGateway(id).pipe(
      tap(() => this.refetchCurrentPage()),
    );
  }

  public changePage(page: number, limit: number): void {
    const tenantId = this._currentTenantId();
    if (tenantId) {
      this.getGatewaysByTenant(tenantId, page, limit);
    } else {
      this.getGateways(page, limit);
    }
  }

  private refetchCurrentPage(): void {
    this.changePage(this._pageIndex(), this._limit());
  }

  private setGettingGatewaysState(): void {
    this._gatewayList.set([]);
    this._loading.set(true);
    this._error.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }
}