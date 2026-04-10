import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs/internal/Observable';

import { environment } from '../../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class SensorCommandApiClientService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  public interruptSensor(sensorId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/sensor/${sensorId}/interrupt`, {});
  }

  public resumeSensor(sensorId: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/sensor/${sensorId}/resume`, {});
  }
}
