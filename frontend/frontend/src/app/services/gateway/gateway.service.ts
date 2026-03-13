import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, finalize } from 'rxjs';

import { GatewayApiClientService } from '../gateway-api-client/gateway-api-client.service';
import { Gateway } from '../../models/gateway.model';
import { GatewayConfig } from '../../models/gateway-config.model';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayService {
  private readonly gatewayApi = inject(GatewayApiClientService);

  private readonly _gatewayList = signal<Gateway[]>([]);
  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);

  public readonly gatewayList = this._gatewayList.asReadonly();
  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();

  public getGatewaysByTenant(tenantId: string): void {
    this._loading.set(true);
    this._error.set(null);

    this.gatewayApi
      .getGatewayListByTenant(tenantId)
      .pipe(
        tap((list) => this._gatewayList.set(list)),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load gateways');
          throw err;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public addNewGateway(config: GatewayConfig): Observable<Gateway> {
    this._loading.set(true);
    this._error.set(null);

    return this.gatewayApi.addNewGateway(config).pipe(
      tap((newGateway) => {
        this._gatewayList.update((list) => [...list, newGateway]);
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to add gateway');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public deleteGateway(id: string): Observable<void> {
    this._loading.set(true);
    this._error.set(null);

    return this.gatewayApi.deleteGateway(id).pipe(
      tap(() => {
        this._gatewayList.update((list) => list.filter((g) => g.id !== id));
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to delete gateway');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }
}
