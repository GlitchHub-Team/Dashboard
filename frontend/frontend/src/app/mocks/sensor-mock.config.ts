// mocks/sensor-mock.config.ts
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

export interface MockFieldConfig {
  key: string;
  baseValue: number;
  step: number;
  amplitude: number;
  min: number;
  max: number;
}

export function getMockFieldConfigs(profile: SensorProfiles): MockFieldConfig[] {
  switch (profile) {
    case SensorProfiles.HEART_RATE_SERVICE:
      return [
        { key: 'BpmValue', baseValue: 75, step: 3, amplitude: 15, min: 40, max: 180 },
      ];

    case SensorProfiles.PULSE_OXIMETER_SERVICE:
      return [
        { key: 'Spo2Value', baseValue: 97, step: 1, amplitude: 3, min: 85, max: 100 },
        { key: 'PulseRateValue', baseValue: 72, step: 2, amplitude: 10, min: 40, max: 160 },
      ];

    case SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE:
      return [
        { key: 'TemperatureValue', baseValue: 22, step: 0.5, amplitude: 5, min: -10, max: 50 },
        { key: 'HumidityValue', baseValue: 45, step: 2, amplitude: 15, min: 0, max: 100 },
      ];

    case SensorProfiles.HEALTH_THERMOMETER_SERVICE:
      return [
        { key: 'TemperatureValue', baseValue: 36.6, step: 0.2, amplitude: 0.5, min: 35, max: 42 },
      ];

    case SensorProfiles.CUSTOM_ECG_SERVICE:
      // Not used directly — ECG has special generation
      return [];
  }
}