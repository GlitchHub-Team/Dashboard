import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable, map } from 'rxjs';

import { environment } from '../../../environments/environment';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { PaginatedSensorResponse } from '../../models/sensor/paginated-sensor-response.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorApiAdapter } from '../../adapters/sensor/sensor-api.adapter';
import { SensorApiClientAdapter } from './sensor-api-client-adapter.service';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientService extends SensorApiClientAdapter {
  private readonly http = inject(HttpClient);
  private readonly mapper = inject(SensorApiAdapter);
  private readonly apiUrl = `${environment.apiUrl}`;

  public getSensorListByGateway(
    gatewayId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<Sensor>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http
      .get<PaginatedSensorResponse<SensorBackend>>(
        `${this.apiUrl}/gateway/${gatewayId}/sensors`,
        { params },
      )
      .pipe(map((response) => this.mapper.fromPaginatedDTO(response)));
  }

  public getSensorListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<Sensor>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http
      .get<PaginatedSensorResponse<SensorBackend>>(
        `${this.apiUrl}/tenant/${tenantId}/sensors`,
        { params },
      )
      .pipe(map((response) => this.mapper.fromPaginatedDTO(response)));
  }

  public addNewSensor(config: SensorConfig): Observable<Sensor> {
    return this.http
      .post<SensorBackend>(`${this.apiUrl}/sensor`, {
        gateway_id: config.gatewayId,
        sensor_name: config.name,
        profile: config.profile,
        data_interval: config.dataInterval,
      })
      .pipe(map((dto) => this.mapper.fromDTO(dto)));
  }

  public deleteSensor(sensorId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/sensor/${sensorId}`);
  }
}