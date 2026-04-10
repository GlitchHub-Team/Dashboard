import { Injectable } from '@angular/core';
import { delay, Observable, of, switchMap, throwError, timer } from 'rxjs';

import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { GatewayConfig } from '../models/gateway/gateway-config.model';
import { PaginatedGatewayResponse } from '../models/gateway/paginated-gateway-response.model';
import { ApiError } from '../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientServiceMock {
  private readonly mockGateways: GatewayBackend[] = [
    // Uncommissioned
    {
      gateway_id: 'gateway-01',
      tenant_id: 'tenant-1',
      name: 'Lobby Principale Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-01-lobby',
    },

    // Tenant 1
    {
      gateway_id: 'gateway-02',
      tenant_id: undefined,
      name: 'ICU Gateway',
      status: 'active',
      interval: 60,
    },
    {
      gateway_id: 'gateway-03',
      tenant_id: 'tenant-1',
      name: 'Reparto B Gateway',
      status: 'inactive',
      interval: 60,
      public_identifier: 'pk-gateway-03-repartob',
    },
    {
      gateway_id: 'gateway-04',
      tenant_id: 'tenant-1',
      name: 'Reparto C Gateway',
      status: 'inactive',
      interval: 60,
      public_identifier: 'pk-gateway-04-repartoc',
    },
    {
      gateway_id: 'gateway-05',
      tenant_id: 'tenant-1',
      name: 'Radiologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-05-radiologia',
    },
    {
      gateway_id: 'gateway-06',
      tenant_id: 'tenant-1',
      name: 'Cardiologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-06-cardiologia',
    },
    {
      gateway_id: 'gateway-07',
      tenant_id: 'tenant-1',
      name: 'Neurologia Gateway',
      status: 'inactive',
      interval: 60,
      public_identifier: 'pk-gateway-07-neurologia',
    },
    {
      gateway_id: 'gateway-08',
      tenant_id: 'tenant-1',
      name: 'Oncologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-08-oncologia',
    },
    {
      gateway_id: 'gateway-09',
      tenant_id: 'tenant-1',
      name: 'Pediatria Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-09-pediatria',
    },
    {
      gateway_id: 'gateway-10',
      tenant_id: 'tenant-1',
      name: 'Ortopedia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-10-ortopedia',
    },
    {
      gateway_id: 'gateway-11',
      tenant_id: 'tenant-1',
      name: 'Dermatologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-11-dermatologia',
    },
    {
      gateway_id: 'gateway-12',
      tenant_id: 'tenant-1',
      name: 'Oftalmologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-12-oftalmologia',
    },
    {
      gateway_id: 'gateway-13',
      tenant_id: 'tenant-1',
      name: 'ENT Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-13-ent',
    },
    {
      gateway_id: 'gateway-14',
      tenant_id: 'tenant-1',
      name: 'Urologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-14-urologia',
    },
    {
      gateway_id: 'gateway-15',
      tenant_id: 'tenant-1',
      name: 'Nefrologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-15-nefrologia',
    },
    {
      gateway_id: 'gateway-16',
      tenant_id: 'tenant-1',
      name: 'Pneumologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-16-pneumologia',
    },
    {
      gateway_id: 'gateway-17',
      tenant_id: 'tenant-1',
      name: 'Gastroenterologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-17-gastro',
    },
    {
      gateway_id: 'gateway-18',
      tenant_id: 'tenant-1',
      name: 'Endocrinologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-18-endocrino',
    },
    {
      gateway_id: 'gateway-19',
      tenant_id: 'tenant-1',
      name: 'Reumatologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-19-reumatologia',
    },
    {
      gateway_id: 'gateway-20',
      tenant_id: 'tenant-1',
      name: 'Ematologia Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-20-ematologia',
    },
    {
      gateway_id: 'gateway-21',
      tenant_id: 'tenant-1',
      name: 'Psichiatria Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-21-psichiatria',
    },
    {
      gateway_id: 'gateway-22',
      tenant_id: 'tenant-1',
      name: 'Ala Chirurgica Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-22-chirurgica',
    },
    {
      gateway_id: 'gateway-23',
      tenant_id: 'tenant-1',
      name: 'Sala di recupero Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-23-recupero',
    },
    {
      gateway_id: 'gateway-24',
      tenant_id: 'tenant-1',
      name: 'Farmacia Gateway',
      status: 'inactive',
      interval: 60,
      public_identifier: 'pk-gateway-24-farmacia',
    },
    {
      gateway_id: 'gateway-25',
      tenant_id: 'tenant-1',
      name: 'Banca del Sangue Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-25-sangue',
    },

    // Tenant 2
    {
      gateway_id: 'gateway-30',
      tenant_id: 'tenant-2',
      name: 'Pronto soccorso Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-30-prontosoccorso',
    },
    {
      gateway_id: 'gateway-31',
      tenant_id: 'tenant-2',
      name: 'Lab Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-31-lab',
    },
    {
      gateway_id: 'gateway-32',
      tenant_id: 'tenant-2',
      name: 'Triage Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-32-triage',
    },
    {
      gateway_id: 'gateway-33',
      tenant_id: 'tenant-2',
      name: 'Baia delle ambulanze Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-33-ambulanze',
    },
    {
      gateway_id: 'gateway-34',
      tenant_id: 'tenant-2',
      name: 'Area di attesa Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-34-attesa',
    },

    // Tenant 4
    {
      gateway_id: 'gateway-40',
      tenant_id: 'tenant-4',
      name: 'Clinic A Gateway',
      status: 'active',
      interval: 60,
      public_identifier: 'pk-gateway-40-clinica',
    },
  ];

  private readonly shouldFailGetGatewayListByTenant = false;
  private readonly shouldFailGetGatewayList = false;
  private readonly shouldFailAddNewGateway = false;
  private readonly shouldFailDeleteGateway = false;

  public getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<GatewayBackend>> {
    if (this.shouldFailGetGatewayListByTenant) {
      return this.delayedError(400, 'Failed to fetch gateways by tenant');
    }

    const filtered = this.mockGateways.filter((g) => g.tenant_id === tenantId);
    return of(this.paginate(filtered, page, limit)).pipe(delay(800));
  }

  public getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedGatewayResponse<GatewayBackend>> {
    if (this.shouldFailGetGatewayList) {
      return this.delayedError(400, 'Failed to fetch gateways');
    }

    return of(this.paginate(this.mockGateways, page, limit)).pipe(delay(800));
  }

  public addNewGateway(config: GatewayConfig): Observable<GatewayBackend> {
    if (this.shouldFailAddNewGateway) {
      return this.delayedError(400, 'Failed to create gateway');
    }

    const newGateway: GatewayBackend = {
      gateway_id: `gateway-${Date.now()}`,
      tenant_id: undefined,
      name: config.name,
      status: 'active',
      interval: config.interval,
    };

    this.mockGateways.push(newGateway);
    return of(newGateway).pipe(delay(400));
  }

  public deleteGateway(gatewayId: string): Observable<void> {
    if (this.shouldFailDeleteGateway) {
      return this.delayedError(400, 'Failed to delete gateway');
    }

    const index = this.mockGateways.findIndex((g) => g.gateway_id === gatewayId);

    if (index === -1) {
      return this.delayedError(404, `Gateway ${gatewayId} not found`);
    }

    this.mockGateways.splice(index, 1);
    return of(undefined).pipe(delay(400));
  }

  private paginate<T>(items: T[], page: number, limit: number): PaginatedGatewayResponse<T> {
    const start = (page - 1) * limit;
    const gateways = items.slice(start, start + limit);

    return {
      count: gateways.length,
      total: items.length,
      gateways,
    };
  }

  private delayedError(status: number, message: string): Observable<never> {
    return timer(500).pipe(switchMap(() => throwError(() => ({ status, message }) as ApiError)));
  }
}
