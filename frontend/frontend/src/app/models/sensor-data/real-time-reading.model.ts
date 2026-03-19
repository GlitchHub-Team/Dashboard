import { SensorProfiles } from '../sensor/sensor-profiles.enum';

export interface RealTimeReading {
  sensorId: string;
  profile: SensorProfiles;
  value: number;
}
