import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { SensorCommandApiClientAdapter } from './sensor-command-api-client-adapter.service';

@Injectable({
  providedIn: 'root',
})
export class SensorCommandApiClientService extends SensorCommandApiClientAdapter {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  public interruptSensor(sensorId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/sensor/${sensorId}/interrupt`, {});
  }

  public resumeSensor(sensorId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/sensor/${sensorId}/resume`, {});
  }
}