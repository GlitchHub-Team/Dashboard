export interface GatewayBackend {
  gateway_id: string;
  tenant_id?: string;
  name: string;
  status: string;
  interval: number;
}
