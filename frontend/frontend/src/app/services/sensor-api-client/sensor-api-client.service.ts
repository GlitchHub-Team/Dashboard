import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { SensorBackend } from '../../models/sensor/sensor-backend.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';
import { PaginatedResponse } from '../../models/paginated-response.model';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/sensor`;

  public getSensorListByGateway(
    gatewayId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<SensorBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedResponse<SensorBackend>>(`${this.apiUrl}/${gatewayId}/list`, {
      params,
    });
  }

  public getSensorListByTenant(
    tenantId: string,
    page: number,
    limit: number,
  ): Observable<PaginatedResponse<SensorBackend>> {
    const params = new HttpParams().set('page', page.toString()).set('limit', limit.toString());

    return this.http.get<PaginatedResponse<SensorBackend>>(
      `${this.apiUrl}/tenant/${tenantId}/list`,
      {
        params,
      },
    );
  }

  public addNewSensor(config: SensorConfig): Observable<SensorBackend> {
    return this.http.post<SensorBackend>(`${this.apiUrl}/add`, config);
  }

  public deleteSensor(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/delete/${id}`);
  }
}
