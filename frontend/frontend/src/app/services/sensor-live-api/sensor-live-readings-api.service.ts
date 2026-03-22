import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/internal/Observable';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';

import { environment } from '../../../environments/environment';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { Sensor } from '../../models/sensor/sensor.model';

@Injectable({
  providedIn: 'root',
})
export class SensorLiveReadingsApiService {
  private readonly apiUrl = `${environment.wsUrl}`;
  private socket$: WebSocketSubject<RealTimeReading> | null = null;

  public connect(sensor: Sensor): Observable<RealTimeReading> {
    const url = `${this.apiUrl}/sensor/${sensor.id}/real_time_data`;

    this.socket$ = webSocket<RealTimeReading>(url);
    return this.socket$.asObservable();
  }

  public disconnect(): void {
    if (this.socket$) {
      this.socket$.complete();
      this.socket$ = null;
    }
  }
}
