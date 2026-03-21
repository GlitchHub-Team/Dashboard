import { Status } from '../gateway-sensor-status.enum';
import { SensorProfiles } from './sensor-profiles.enum';

export interface Sensor {
  id: string;
  gatewayId: string;
  name: string;
  profile: SensorProfiles;
  status: Status;
  dataInterval?: number;
}
