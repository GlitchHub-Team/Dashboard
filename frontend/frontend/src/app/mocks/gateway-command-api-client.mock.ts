import { Injectable } from '@angular/core';
import { delay, Observable, of } from 'rxjs';

import { GatewayBackend } from '../models/gateway/gateway-backend.model';

@Injectable({ providedIn: 'root' })
export class GatewayCommandApiClientMockService {
  private readonly mockGateway: GatewayBackend = {
    gateway_id: 'gateway-01',
    tenant_id: 'tenant-01',
    name: 'Main Lobby Gateway',
    status: 'active',
    interval: 60,
  };

  public commissionGateway(_gatewayId: string): Observable<GatewayBackend> {
    return of({ ...this.mockGateway, gateway_id: _gatewayId, status: 'active' }).pipe(delay(500));
  }

  public decommissionGateway(_gatewayId: string): Observable<void> {
    return of(void 0).pipe(delay(500));
  }

  public resetGateway(_gatewayId: string): Observable<void> {
    return of(void 0).pipe(delay(500));
  }

  public rebootGateway(_gatewayId: string): Observable<void> {
    return of(void 0).pipe(delay(500));
  }
}
