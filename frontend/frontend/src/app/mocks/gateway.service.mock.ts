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
        tenantId: 'tenant-01',
        name: 'Main Lobby Gateway',
        status: GatewayStatus.ONLINE,
      },
      {
        id: 'gateway-02',
        tenantId: 'tenant-01',
        name: 'ICU Gateway',
        status: GatewayStatus.ONLINE,
      },
      {
        id: 'gateway-03',
        tenantId: 'tenant-01',
        name: 'Ward B Gateway',
        status: GatewayStatus.OFFLINE,
      },
    ],
    'tenant-02': [
      {
        id: 'gateway-04',
        tenantId: 'tenant-02',
        name: 'Emergency Room Gateway',
        status: GatewayStatus.ONLINE,
      },
      {
        id: 'gateway-05',
        tenantId: 'tenant-02',
        name: 'Lab Gateway',
        status: GatewayStatus.OFFLINE,
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        id: 'gateway-06',
        tenantId: 'tenant-04',
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
