// mocks/gateway-service.mock.ts
import { Injectable } from '@angular/core';
import { delay, Observable, of, throwError } from 'rxjs';

import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../models/gateway/paginated-gateway-response.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientServiceMock {
  private readonly mockGateways = new Map<string, GatewayBackend[]>([
    [
      'tenant-1',
      [
        {
          gateway_id: 'gateway-01',
          tenant_id: 'tenant-1',
          name: 'Lobby Principale Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-02',
          tenant_id: 'tenant-1',
          name: 'ICU Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-03',
          tenant_id: 'tenant-1',
          name: 'Reparto B Gateway',
          status: 'inattivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-04',
          tenant_id: 'tenant-1',
          name: 'Reparto C Gateway',
          status: 'inattivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-05',
          tenant_id: 'tenant-1',
          name: 'Radiologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-06',
          tenant_id: 'tenant-1',
          name: 'Cardiologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-07',
          tenant_id: 'tenant-1',
          name: 'Neurologia Gateway',
          status: 'inattivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-08',
          tenant_id: 'tenant-1',
          name: 'Oncologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-09',
          tenant_id: 'tenant-1',
          name: 'Pediatria Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-10',
          tenant_id: 'tenant-1',
          name: 'Ortopedia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-11',
          tenant_id: 'tenant-1',
          name: 'Dermatologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-12',
          tenant_id: 'tenant-1',
          name: 'Oftalmologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-13',
          tenant_id: 'tenant-1',
          name: 'ENT Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-14',
          tenant_id: 'tenant-1',
          name: 'Urologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-15',
          tenant_id: 'tenant-1',
          name: 'Nefrologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-16',
          tenant_id: 'tenant-1',
          name: 'Pneumologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-17',
          tenant_id: 'tenant-1',
          name: 'Gastroenterologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-18',
          tenant_id: 'tenant-1',
          name: 'Endocrinologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-19',
          tenant_id: 'tenant-1',
          name: 'Reumatologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-20',
          tenant_id: 'tenant-1',
          name: 'Ematologia Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-21',
          tenant_id: 'tenant-1',
          name: 'Psichiatria Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-22',
          tenant_id: 'tenant-1',
          name: 'Ala Chirurgica Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-23',
          tenant_id: 'tenant-1',
          name: 'Sala di recupero Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-24',
          tenant_id: 'tenant-1',
          name: 'Farmacia Gateway',
          status: 'inattivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-25',
          tenant_id: 'tenant-1',
          name: 'Banca del Sangue Gateway',
          status: 'attivo',
          interval: 60,
        },
      ],
    ],
    [
      'tenant-2',
      [
        {
          gateway_id: 'gateway-30',
          tenant_id: 'tenant-2',
          name: 'Pronto soccorso Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-31',
          tenant_id: 'tenant-2',
          name: 'Lab Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-32',
          tenant_id: 'tenant-2',
          name: 'Triage Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-33',
          tenant_id: 'tenant-2',
          name: 'Baia delle ambulanze Gateway',
          status: 'attivo',
          interval: 60,
        },
        {
          gateway_id: 'gateway-34',
          tenant_id: 'tenant-2',
          name: 'Area di attesa Gateway',
          status: 'attivo',
          interval: 60,
        },
      ],
    ],
    ['tenant-3', []],
    [
      'tenant-4',
      [
        {
          gateway_id: 'gateway-40',
          tenant_id: 'tenant-4',
          name: 'Clinic A Gateway',
          status: 'attivo',
          interval: 60,
        },
      ],
    ],
    ['tenant-5', []],
  ]);

  public getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<GatewayBackend>> {
    const all = this.mockGateways.get(tenantId) ?? [];
    return of(this.paginate(all, page, limit)).pipe(delay(800));
  }

  public getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<GatewayBackend>> {
    const all = Array.from(this.mockGateways.values()).flat();
    return of(this.paginate(all, page, limit)).pipe(delay(800));
  }

  public addNewGateway(config: GatewayConfig): Observable<GatewayBackend> {
    const tenantId = 'tenant-1';

    const tenantGateways = this.mockGateways.get(tenantId);

    if (!tenantGateways) {
      return throwError(() => ({
        status: 404,
        message: `Tenant ${tenantId} not found`,
      })).pipe(delay(400));
    }

    const newGateway: GatewayBackend = {
      gateway_id: `gateway-${Date.now()}`,
      tenant_id: tenantId,
      name: config.name,
      status: 'attivo',
      interval: config.interval,
    };

    tenantGateways.push(newGateway);

    return of(newGateway).pipe(delay(400));
  }

  public deleteGateway(gatewayId: string): Observable<void> {
    let found = false;

    for (const [, gateways] of this.mockGateways) {
      const index = gateways.findIndex((g) => g.gateway_id === gatewayId);
      if (index !== -1) {
        gateways.splice(index, 1);
        found = true;
        break;
      }
    }

    if (!found) {
      return throwError(() => ({
        status: 404,
        message: `Gateway ${gatewayId} not found`,
      })).pipe(delay(400));
    }

    return of(undefined).pipe(delay(400));
  }

  private paginate<T>(items: T[], page: number, limit: number): PaginatedGatewayResponse<T> {
    const start = page * limit;
    const gateways = items.slice(start, start + limit);

    return {
      count: gateways.length,
      total: items.length,
      gateways,
    };
  }
}
