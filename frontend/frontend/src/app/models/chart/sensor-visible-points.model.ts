import { SensorProfiles } from "../sensor/sensor-profiles.enum";

export const SENSOR_VISIBLE_POINTS: Record<string, number> = {
  [SensorProfiles.HEART_RATE_SERVICE]: 50,
  [SensorProfiles.PULSE_OXIMETER_SERVICE]: 50,
  [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: 50,
  [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: 50,
  [SensorProfiles.CUSTOM_ECG_SERVICE]: 1250,
};

export const SENSOR_MAX_LIVE_READINGS: Record<string, number> = {
  [SensorProfiles.HEART_RATE_SERVICE]: 50,
  [SensorProfiles.PULSE_OXIMETER_SERVICE]: 50,
  [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: 50,
  [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: 50,
  [SensorProfiles.CUSTOM_ECG_SERVICE]: 1250,
};