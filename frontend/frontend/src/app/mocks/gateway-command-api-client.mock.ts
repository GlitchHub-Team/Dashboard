import { Injectable } from '@angular/core';
import { delay, Observable, of, switchMap, throwError, timer } from 'rxjs';

import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { ApiError } from '../models/api-error.model';

@Injectable({ providedIn: 'root' })
export class GatewayCommandApiClientMockService {
  private readonly shouldFailCommission = false;
  private readonly shouldFailDecommission = false;
  private readonly shouldFailReset = false;
  private readonly shouldFailReboot = false;

  private readonly mockGateway: GatewayBackend = {
    gateway_id: 'gateway-01',
    tenant_id: 'tenant-01',
    name: 'Lobby principale Gateway',
    status: 'attivo',
    interval: 60,
  };

  public commissionGateway(_gatewayId: string): Observable<GatewayBackend> {
    if (this.shouldFailCommission) {
      return this.delayedError(400, 'Failed to commission gateway');
    }
    return of({ ...this.mockGateway, gateway_id: _gatewayId, status: 'attivo' }).pipe(delay(500));
  }

  public decommissionGateway(_gatewayId: string): Observable<void> {
    if (this.shouldFailDecommission) {
      return this.delayedError(400, 'Failed to decommission gateway');
    }
    return of(void 0).pipe(delay(500));
  }

  public resetGateway(_gatewayId: string): Observable<void> {
    if (this.shouldFailReset) {
      return this.delayedError(400, 'Failed to reset gateway');
    }
    return of(void 0).pipe(delay(500));
  }

  public rebootGateway(_gatewayId: string): Observable<void> {
    if (this.shouldFailReboot) {
      return this.delayedError(400, 'Failed to reboot gateway');
    }
    return of(void 0).pipe(delay(500));
  }

  private delayedError(status: number, message: string): Observable<never> {
    return timer(500).pipe(switchMap(() => throwError(() => ({ status, message }) as ApiError)));
  }
}
