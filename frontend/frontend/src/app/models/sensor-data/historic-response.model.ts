export interface HistoricResponse {
  count: number;
  samples: HistoricSample[];
}

export interface HistoricSample {
  sensor_id: string;
  gateway_id: string;
  tenant_id: string;
  timestamp: string;
  profile: string;
  data: Record<string, any>;
}
