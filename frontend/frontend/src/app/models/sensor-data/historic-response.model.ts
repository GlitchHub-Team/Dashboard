import { HistoricDataPoint } from './historic-data-point.model';

export interface HistoricResponse {
  count: number;
  resolution: number;
  data: HistoricDataPoint[];
}
