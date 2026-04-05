import { ChartType } from './chart-type.enum';
import { Sensor } from '../sensor/sensor.model';
import { TimeInterval } from './time-interval.model';
import { ValuesInterval } from './values-interval.model';

export interface ChartRequest {
  sensor: Sensor;
  chartType: ChartType;
  dataPointsCounter?: number;
  timeInterval?: TimeInterval;
  valuesInterval?: ValuesInterval;
}
