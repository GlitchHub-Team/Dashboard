import { SensorProfiles } from './sensor-profiles.enum';

export interface SensorConfig {
  name: string;
  dataInterval: number;
  gatewayId: string;
  profile: SensorProfiles;
}
