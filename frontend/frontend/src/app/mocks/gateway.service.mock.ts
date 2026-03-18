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
      { GatewayId: 'gateway-01', GatewayName: 'Main Lobby Gateway' },
      { GatewayId: 'gateway-02', GatewayName: 'ICU Gateway' },
      { GatewayId: 'gateway-03', GatewayName: 'Ward B Gateway' },
      { GatewayId: 'gateway-04', GatewayName: 'Ward C Gateway' },
      { GatewayId: 'gateway-05', GatewayName: 'Radiology Gateway' },
      { GatewayId: 'gateway-06', GatewayName: 'Cardiology Gateway' },
      { GatewayId: 'gateway-07', GatewayName: 'Neurology Gateway' },
      { GatewayId: 'gateway-08', GatewayName: 'Oncology Gateway' },
      { GatewayId: 'gateway-09', GatewayName: 'Pediatrics Gateway' },
      { GatewayId: 'gateway-10', GatewayName: 'Orthopedics Gateway' },
      { GatewayId: 'gateway-11', GatewayName: 'Dermatology Gateway' },
      { GatewayId: 'gateway-12', GatewayName: 'Ophthalmology Gateway' },
      { GatewayId: 'gateway-13', GatewayName: 'ENT Gateway' },
      { GatewayId: 'gateway-14', GatewayName: 'Urology Gateway' },
      { GatewayId: 'gateway-15', GatewayName: 'Nephrology Gateway' },
      { GatewayId: 'gateway-16', GatewayName: 'Pulmonology Gateway' },
      { GatewayId: 'gateway-17', GatewayName: 'Gastroenterology Gateway' },
      { GatewayId: 'gateway-18', GatewayName: 'Endocrinology Gateway' },
      { GatewayId: 'gateway-19', GatewayName: 'Rheumatology Gateway' },
      { GatewayId: 'gateway-20', GatewayName: 'Hematology Gateway' },
      { GatewayId: 'gateway-21', GatewayName: 'Psychiatry Gateway' },
      { GatewayId: 'gateway-22', GatewayName: 'Surgery Wing Gateway' },
      { GatewayId: 'gateway-23', GatewayName: 'Recovery Room Gateway' },
      { GatewayId: 'gateway-24', GatewayName: 'Pharmacy Gateway' },
      { GatewayId: 'gateway-25', GatewayName: 'Blood Bank Gateway' },
    ],
    'tenant-02': [
      { GatewayId: 'gateway-30', GatewayName: 'Emergency Room Gateway' },
      { GatewayId: 'gateway-31', GatewayName: 'Lab Gateway' },
      { GatewayId: 'gateway-32', GatewayName: 'Triage Gateway' },
      { GatewayId: 'gateway-33', GatewayName: 'Ambulance Bay Gateway' },
      { GatewayId: 'gateway-34', GatewayName: 'Waiting Area Gateway' },
    ],
    'tenant-03': [],
    'tenant-04': [{ GatewayId: 'gateway-40', GatewayName: 'Clinic A Gateway' }],
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
      GatewayId: `gateway-${Date.now()}`,
      GatewayName: 'New Gateway',
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
