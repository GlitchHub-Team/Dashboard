export interface HistoricResponse {
  count: {
    current: number;
    real: number;
    total: number;
  };
  duration: number;
  // Timestamp in milliseconds UNIX
  dataset: {
    timestamps: number[];
    values: number[];
  };
  unit: string;
}
