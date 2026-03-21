import { SensorProfiles } from '../sensor/sensor-profiles.enum';

export interface HistoricDataPoint {
  sensorId: string;
  timestamp: string;
  profile: SensorProfiles;
  value: number;
}
