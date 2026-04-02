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

  public getHistoricData(req: ChartRequest): Observable<HistoricResponse> {
    // DataPoints è richiesto quindi sono sicuro che sia valorizzato
    let params = new HttpParams().set('max_data_points', req.dataPointsCounter!);

    if (req.timeInterval) {
      params = params
        .set('from_time', req.timeInterval.from.toISOString())
        .set('to_time', req.timeInterval.to.toISOString());
    }

    // APIDOG richiede che siano number
    if (req.valuesInterval) {
      params = params
        .set('lower_bound', req.valuesInterval.lowerBound)
        .set('upper_bound', req.valuesInterval.upperBound);
    }

    return this.http.get<HistoricResponse>(
      `${this.apiUrl}/sensor/${req.sensor.id}/historical_data`,
      { params },
    );
  }
}
