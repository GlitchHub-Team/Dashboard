import { Injectable } from '@angular/core';
import { delay, Observable, of } from 'rxjs';

import { Gateway } from '../models/gateway.model';
//import { GatewayConfig } from '../models/gateway-config.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayServiceMock {
  private readonly mockGateways: Record<string, Gateway[]> = {
    'tenant-01': [
      {
        id: 'gateway-01',
      },
      {
        id: 'gateway-02',
      },
      {
        id: 'gateway-03',
      },
    ],
    'tenant-02': [
      {
        id: 'gateway-04',
      },
      {
        id: 'gateway-05',
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        id: 'gateway-06',
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

  /*   public addNewGateway(config: GatewayConfig): Observable<Gateway> {
    const newGateway: Gateway = {
      id: config.name,
    };
    this.mockGateways.push(newGateway);
    return of(newGateway).pipe(delay(500));
  } */

  /*   public deleteGateway(id: string): Observable<void> {
    const index = this.mockGateways.findIndex((g) => g.id === id);
    if (index !== -1) {
      this.mockGateways.splice(index, 1);
    }
    return of(void 0).pipe(delay(500));
  } */

  /*  
  public sendCommand(cmd: GatewayCommand): Observable<CommandResult> {
    return of({
      success: true,
      message: `Command ${cmd.type} sent to ${cmd.gatewayId}`,
    }).pipe(delay(1000));
  } 
  */
}
