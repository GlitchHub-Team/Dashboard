import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, finalize } from 'rxjs';

import { SensorApiClientService } from '../sensor-api-client/sensor-api-client.service';
import { Sensor } from '../../models/sensor.model';
import { SensorConfig } from '../../models/sensor-config.model';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class SensorService {
  private readonly sensorApi = inject(SensorApiClientService);

  private readonly _sensorList = signal<Sensor[]>([]);
  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);

  public readonly sensorList = this._sensorList.asReadonly();
  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();

  public getSensorsByGateway(gatewayId: string): void {
    this._loading.set(true);
    this._error.set(null);

    this.sensorApi
      .getSensorListByGateway(gatewayId)
      .pipe(
        tap((list) => this._sensorList.set(list)),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load sensors');
          throw err;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public getTenantSensorsByTenant(tenantId: string): void {
    this._loading.set(true);
    this._error.set(null);

    this.sensorApi
      .getSensorListByTenant(tenantId)
      .pipe(
        tap((list) => this._sensorList.set(list)),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load sensors');
          throw err;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public addNewSensor(config: SensorConfig): Observable<Sensor> {
    this._loading.set(true);
    this._error.set(null);

    return this.sensorApi.addNewSensor(config).pipe(
      tap((newSensor) => {
        this._sensorList.update((list) => [...list, newSensor]);
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to add sensor');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public deleteSensor(id: string): Observable<void> {
    this._loading.set(true);
    this._error.set(null);

    return this.sensorApi.deleteSensor(id).pipe(
      tap(() => {
        this._sensorList.update((list) => list.filter((s) => s.id !== id));
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to delete sensor');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }
}
