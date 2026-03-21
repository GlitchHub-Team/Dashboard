import { SensorProfiles } from './sensor-profiles.enum';

export interface SensorConfig {
  gatewayId: string;
  name: string;
  profile: SensorProfiles;
}
