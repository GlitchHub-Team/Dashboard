import { GatewayStatus } from './gateway-status.enum';

export interface Gateway {
  id: string;
  name: string;
  status: GatewayStatus;
}
