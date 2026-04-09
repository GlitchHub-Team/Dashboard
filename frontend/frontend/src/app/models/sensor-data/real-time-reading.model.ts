export interface RealTimeReading {
  timestamp: string;
  profile: string;
  data: Record<string, any>;
}
