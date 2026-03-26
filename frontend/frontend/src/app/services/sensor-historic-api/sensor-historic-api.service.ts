import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { ChartRequest } from '../../models/chart/chart-request.model';

@Injectable({
  providedIn: 'root',
})
export class SensorHistoricApiService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = `${environment.apiUrl}`;

  // Setta tutti i parametri necessari per la richiesta di dati storici
  // TODO: Il time andrà settato come DATE TIME, da rivedere anche sulla parte di UI
  public getHistoricData(req: ChartRequest): Observable<HistoricResponse> {
    const params = new HttpParams()
      .set('from_time', req.timeInterval!.from.toISOString())
      .set('to_time', req.timeInterval!.to.toISOString())
      .set('lower_bound', req.valuesInterval!.lowerBound.toString())
      .set('upper_bound', req.valuesInterval!.upperBound.toString())
      .set('max_data_points', req.dataPointsCounter!.toString());
    return this.http.get<HistoricResponse>(
      `${this.apiUrl}/sensor/${req.sensor.id}/historical-data`,
      {
        params,
      },
    );
  }
}
