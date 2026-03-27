import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { PaginatedSensorResponse } from '../../models/sensor/paginated-sensor-response.model';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  public getSensorListByGateway(
    gatewayId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<SensorBackend>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http.get<PaginatedSensorResponse<SensorBackend>>(
      `${this.apiUrl}/gateway/${gatewayId}/sensors`,
      {
        params,
      },
    );
  }

  public getSensorListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedSensorResponse<SensorBackend>> {
    const params = new HttpParams().set('page', page).set('limit', limit);

    return this.http.get<PaginatedSensorResponse<SensorBackend>>(
      `${this.apiUrl}/tenant/${tenantId}/sensors`,
      {
        params,
      },
    );
  }

  public addNewSensor(config: SensorConfig): Observable<SensorBackend> {
    return this.http.post<SensorBackend>(`${this.apiUrl}/sensor`, config);
  }

  public deleteSensor(sensorId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/sensor/${sensorId}`);
  }
}
