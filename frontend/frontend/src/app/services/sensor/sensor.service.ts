import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, EMPTY, finalize } from 'rxjs';

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
    this.setGettingSensorsState();

    this.sensorApi
      .getSensorListByGateway(gatewayId)
      .pipe(
        tap((list) => this._sensorList.set(list)),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load sensors');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public getSensorsByTenant(tenantId: string): void {
    this.setGettingSensorsState();

    this.sensorApi
      .getSensorListByTenant(tenantId)
      .pipe(
        tap((list) => this._sensorList.set(list)),
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load sensors');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      )
      .subscribe();
  }

  public addNewSensor(config: SensorConfig): Observable<Sensor> {
    this.setLoadingState();

    return this.sensorApi.addNewSensor(config).pipe(
      tap((newSensor) => {
        this._sensorList.update((list) => [...list, newSensor]);
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to add sensor');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public deleteSensor(id: string): Observable<void> {
    this.setLoadingState();

    return this.sensorApi.deleteSensor(id).pipe(
      tap(() => {
        this._sensorList.update((list) => list.filter((s) => s.id !== id));
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to delete sensor');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public clearSensors(): void {
    this._sensorList.set([]);
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
