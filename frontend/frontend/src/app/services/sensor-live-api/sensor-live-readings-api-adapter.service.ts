import { Observable } from 'rxjs';

import { RealTimeReading } from '../../models/sensor-data/real-time-reading.model';
import { ChartRequest } from '../../models/chart/chart-request.model';

export abstract class SensorLiveReadingsApiAdapter {
  abstract connect(req: ChartRequest): Observable<RealTimeReading>;
  abstract disconnect(): void;
}