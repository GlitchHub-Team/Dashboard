export interface SensorBackend {
  SensorId: string;
  GatewayId: string;
  Name: string;
  Profile: string;
  DataInterval?: number;
}
