import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, EMPTY, finalize, map } from 'rxjs';

import { SensorApiClientService } from '../sensor-api-client/sensor-api-client.service';
import { SensorAdapter } from '../../adapters/sensor/sensor.adapter';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class SensorService {
  private readonly sensorApi = inject(SensorApiClientService);
  private readonly adapter = inject(SensorAdapter);

  private readonly _sensorList = signal<Sensor[]>([]);
  private readonly _total = signal(0);
  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _pageIndex = signal(0);
  private readonly _limit = signal(10);

  private readonly _currentGatewayId = signal<string | null>(null);
  private readonly _currentTenantId = signal<string | null>(null);

  public readonly sensorList = this._sensorList.asReadonly();
  public readonly total = this._total.asReadonly();
  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly pageIndex = this._pageIndex.asReadonly();
  public readonly limit = this._limit.asReadonly();

  public getSensorsByGateway(gatewayId: string, page: number, limit: number): void {
    this._currentGatewayId.set(gatewayId);
    this._currentTenantId.set(null);
    this._pageIndex.set(page);
    this._limit.set(limit);
    this.setGettingSensorsState();

    this.sensorApi
      .getSensorListByGateway(gatewayId, page + 1, limit)
      .pipe(
        // Adapting della response al formato usato dal frontend (quindi da SensorBackend a Sensor)
        map((response) => this.adapter.fromPaginatedDTO(response)),
        tap((result) => {
          this._sensorList.set(result.sensors);
          this._total.set(result.total);
        }),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load sensors');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public getSensorsByTenant(tenantId: string, page: number, limit: number): void {
    this._currentTenantId.set(tenantId);
    this._currentGatewayId.set(null);
    this._pageIndex.set(page);
    this._limit.set(limit);
    this.setGettingSensorsState();

    this.sensorApi
      .getSensorListByTenant(tenantId, page + 1, limit)
      .pipe(
        // Adapting della response al formato usato dal frontend (quindi da SensorBackend a Sensor)
        map((response) => this.adapter.fromPaginatedDTO(response)),
        tap((result) => {
          this._sensorList.set(result.sensors);
          this._total.set(result.total);
        }),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load sensors');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public addNewSensor(config: SensorConfig): Observable<Sensor> {
    return this.sensorApi.addNewSensor(config).pipe(map((dto) => this.adapter.fromDTO(dto)));
  }

  public deleteSensor(id: string): Observable<void> {
    this.setLoadingState();

    return this.sensorApi.deleteSensor(id).pipe(
      tap(() => this.refetchCurrentPage()),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to delete sensor');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public changePage(page: number, limit: number): void {
    const gatewayId = this._currentGatewayId();
    const tenantId = this._currentTenantId();

    if (gatewayId) {
      this.getSensorsByGateway(gatewayId, page, limit);
    } else if (tenantId) {
      this.getSensorsByTenant(tenantId, page, limit);
    }
  }

  public clearSensors(): void {
    this._sensorList.set([]);
    this._total.set(0);
    this._currentGatewayId.set(null);
    this._currentTenantId.set(null);
    this._pageIndex.set(0);
    this._error.set(null);
  }

  private refetchCurrentPage(): void {
    this.changePage(this._pageIndex(), this._limit());
  }

  private setGettingSensorsState(): void {
    this._sensorList.set([]);
    this._loading.set(true);
    this._error.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }
}
