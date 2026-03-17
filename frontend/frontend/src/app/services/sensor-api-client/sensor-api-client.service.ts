import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Sensor } from '../../models/sensor.model';
import { SensorConfig } from '../../models/sensor-config.model';

@Injectable({
  providedIn: 'root',
})
export class SensorApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/sensor`;

  public getSensorListByGateway(gatewayId: string): Observable<Sensor[]> {
    return this.http.get<Sensor[]>(`${this.apiUrl}/list`, { params: { gatewayId } });
  }

  public getSensorListByTenant(tenantId: string): Observable<Sensor[]> {
    return this.http.get<Sensor[]>(`${this.apiUrl}/list`, { params: { tenantId } });
  }

  public addNewSensor(config: SensorConfig): Observable<Sensor> {
    return this.http.post<Sensor>(`${this.apiUrl}/add`, config);
  }

  public deleteSensor(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/delete/${id}`);
  }
}
