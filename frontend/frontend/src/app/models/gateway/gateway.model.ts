import { GatewayStatus } from '../gateway-status.enum';

export interface Gateway {
  id: string;
  tenantId?: string;
  name: string;
  status: GatewayStatus;
  interval: number;
  publicIdentifier?: string;
}
