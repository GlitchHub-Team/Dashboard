import { SensorProfiles } from './sensor-profiles.enum';

export interface Sensor {
  id: string;
  gatewayId: string;
  name: string;
  profile: SensorProfiles;
  dataInterval?: number;
}
