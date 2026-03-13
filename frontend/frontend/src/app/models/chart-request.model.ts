import { ChartType } from './chart-type.enum';
import { Sensor } from './sensor.model';
import { TimeInterval } from './time-interval.model';

export interface ChartRequest {
  sensor: Sensor;
  chartType: ChartType;
  timeInterval?: TimeInterval;
}
