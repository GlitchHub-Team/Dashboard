import { Injectable } from '@angular/core';
import { delay, Observable, of } from 'rxjs';

import { Gateway } from '../models/gateway.model';
import { GatewayStatus } from '../models/gateway-status.enum';

@Injectable({
  providedIn: 'root',
})
export class GatewayServiceMock {
  private readonly mockGateways: Record<string, Gateway[]> = {
    'tenant-01': [
      {
        id: 'gateway-01',
        name: 'Main Lobby Gateway',
        status: GatewayStatus.ONLINE,
      },
      {
        id: 'gateway-02',
        name: 'ICU Gateway',
        status: GatewayStatus.ONLINE,
      },
      {
        id: 'gateway-03',
        name: 'Ward B Gateway',
        status: GatewayStatus.OFFLINE,
      },
    ],
    'tenant-02': [
      {
        id: 'gateway-04',
        name: 'Emergency Room Gateway',
        status: GatewayStatus.ONLINE,
      },
      {
        id: 'gateway-05',
        name: 'Lab Gateway',
        status: GatewayStatus.OFFLINE,
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        id: 'gateway-06',
        name: 'Pharmacy Gateway',
        status: GatewayStatus.ONLINE,
      },
    ],
    'tenant-05': [],
  };

  public getGatewayList(): Observable<Gateway[]> {
    const allGateways = Object.values(this.mockGateways).flat();
    return of(allGateways).pipe(delay(800));
  }

  public getGatewayListByTenant(tenantId: string): Observable<Gateway[]> {
    const gateways = this.mockGateways[tenantId] ?? [];
    return of(gateways).pipe(delay(800));
  }
}
