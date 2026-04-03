import { Status } from '../gateway-sensor-status.enum';

export interface Gateway {
  id: string;
  tenantId?: string;
  name: string;
  status: Status;
  interval: number;
}
