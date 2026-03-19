import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { TimeInterval } from '../../models/time-interval.model';
import { Sensor } from '../../models/sensor/sensor.model';

@Injectable({
  providedIn: 'root',
})
export class SensorHistoricApiService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}/sensor-historic`;

  public getHistoricData(sensor: Sensor, timeInterval: TimeInterval): Observable<HistoricResponse> {
    const params = new HttpParams()
      .set('from', timeInterval.from.toISOString())
      .set('to', timeInterval.to.toISOString());

    return this.http.get<HistoricResponse>(`${this.apiUrl}/data?sensorId=${sensor.id}`, { params });
  }
}
