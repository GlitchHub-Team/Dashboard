import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/internal/Observable';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';

import { environment } from '../../../environments/environment';
import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { Subject, takeUntil } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class SensorLiveReadingsApiService {
  private readonly apiUrl = `${environment.wsUrl}`;
  private socket$: WebSocketSubject<RealTimeReading> | null = null;
  private readonly disconnect$ = new Subject<void>();

  // Connette al WS e ritorna l'Observable per recuperare le letture
  public connect(sensor: Sensor): Observable<RealTimeReading> {
    this.disconnect();

    const url = `${this.apiUrl}/sensor/${sensor.id}/real_time_data`;

    this.socket$ = webSocket<RealTimeReading>(url);
    return this.socket$.pipe(takeUntil(this.disconnect$));
  }

  public disconnect(): void {
    if (this.socket$) {
      this.disconnect$.next();
      this.socket$.complete();
      this.socket$ = null;
    }
  }
}
