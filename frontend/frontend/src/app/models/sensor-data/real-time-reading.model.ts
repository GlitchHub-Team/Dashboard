export interface RealTimeReading {
  sensor_id: string;
  gateway_id: string;
  tenant_id: string;
  timestamp: string;
  profile: string;
  data: Record<string, any>;
}
