// mocks/gateway-service.mock.ts
import { Injectable } from '@angular/core';
import { delay, Observable, of } from 'rxjs';

import { GatewayBackend } from '../models/gateway/gateway-backend.model';
import { PaginatedResponse } from '../models/paginated-response.model';

@Injectable({
  providedIn: 'root',
})
export class GatewayApiClientServiceMock {
  private readonly mockGateways: Record<string, GatewayBackend[]> = {
    'tenant-01': [
      {
        gateway_id: 'gateway-01',
        tenant_id: 'tenant-01',
        name: 'Main Lobby Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-02',
        tenant_id: 'tenant-01',
        name: 'ICU Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-03',
        tenant_id: 'tenant-01',
        name: 'Ward B Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-04',
        tenant_id: 'tenant-01',
        name: 'Ward C Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-05',
        tenant_id: 'tenant-01',
        name: 'Radiology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-06',
        tenant_id: 'tenant-01',
        name: 'Cardiology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-07',
        tenant_id: 'tenant-01',
        name: 'Neurology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-08',
        tenant_id: 'tenant-01',
        name: 'Oncology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-09',
        tenant_id: 'tenant-01',
        name: 'Pediatrics Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-10',
        tenant_id: 'tenant-01',
        name: 'Orthopedics Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-11',
        tenant_id: 'tenant-01',
        name: 'Dermatology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-12',
        tenant_id: 'tenant-01',
        name: 'Ophthalmology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-13',
        tenant_id: 'tenant-01',
        name: 'ENT Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-14',
        tenant_id: 'tenant-01',
        name: 'Urology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-15',
        tenant_id: 'tenant-01',
        name: 'Nephrology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-16',
        tenant_id: 'tenant-01',
        name: 'Pulmonology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-17',
        tenant_id: 'tenant-01',
        name: 'Gastroenterology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-18',
        tenant_id: 'tenant-01',
        name: 'Endocrinology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-19',
        tenant_id: 'tenant-01',
        name: 'Rheumatology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-20',
        tenant_id: 'tenant-01',
        name: 'Hematology Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-21',
        tenant_id: 'tenant-01',
        name: 'Psychiatry Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-22',
        tenant_id: 'tenant-01',
        name: 'Surgery Wing Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-23',
        tenant_id: 'tenant-01',
        name: 'Recovery Room Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-24',
        tenant_id: 'tenant-01',
        name: 'Pharmacy Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-25',
        tenant_id: 'tenant-01',
        name: 'Blood Bank Gateway',
        status: 'active',
        intervals: 60,
      },
    ],
    'tenant-02': [
      {
        gateway_id: 'gateway-30',
        tenant_id: 'tenant-02',
        name: 'Emergency Room Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-31',
        tenant_id: 'tenant-02',
        name: 'Lab Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-32',
        tenant_id: 'tenant-02',
        name: 'Triage Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-33',
        tenant_id: 'tenant-02',
        name: 'Ambulance Bay Gateway',
        status: 'active',
        intervals: 60,
      },
      {
        gateway_id: 'gateway-34',
        tenant_id: 'tenant-02',
        name: 'Waiting Area Gateway',
        status: 'active',
        intervals: 60,
      },
    ],
    'tenant-03': [],
    'tenant-04': [
      {
        gateway_id: 'gateway-40',
        tenant_id: 'tenant-04',
        name: 'Clinic A Gateway',
        status: 'active',
        intervals: 60,
      },
    ],
    'tenant-05': [],
  };

  public getGatewayListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<GatewayBackend>> {
    const all = this.mockGateways[tenantId] ?? [];
    return of(this.paginate(all, page, limit)).pipe(delay(800));
  }

  public getGatewayList(
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<GatewayBackend>> {
    const all = Object.values(this.mockGateways).flat();
    return of(this.paginate(all, page, limit)).pipe(delay(800));
  }

  public addNewGateway(config: unknown): Observable<GatewayBackend> {
    return of({
      gateway_id: `gateway-${Date.now()}`,
      name: 'New Gateway',
      status: 'active',
      intervals: 60,
    }).pipe(delay(400));
  }

  public deleteGateway(id: string): Observable<void> {
    return of(undefined).pipe(delay(400));
  }

  public sendCommandToGateway(): Observable<void> {
    return of(undefined).pipe(delay(400));
  }

  private paginate<T>(items: T[], page: number, limit: number): PaginatedResponse<T> {
    const start = page * limit;
    const data = items.slice(start, start + limit);

    return {
      count: data.length,
      total: items.length,
      data,
    };
  }
}
