import { Observable } from 'rxjs';

import { HistoricResponse } from '../../models/sensor-data/historic-response.model';
import { ChartRequest } from '../../models/chart/chart-request.model';

export abstract class SensorHistoricApiAdapter {
  abstract getHistoricData(req: ChartRequest): Observable<HistoricResponse>;
}