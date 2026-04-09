export interface SensorBackend {
  sensor_id: string;
  gateway_id: string;
  sensor_name: string;
  status: string;
  profile: string;
  data_interval: number;
}
