import { SensorReading } from './sensor-reading.model';

export interface HistoricReadings {
  dataCount: number;
  readings: SensorReading[];
}
