import { SensorReading } from './sensor-reading.model';

export interface HistoricReadings {
  resolution: number;
  readings: SensorReading[];
}
