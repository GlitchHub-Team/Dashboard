import { describe, it, expect } from 'vitest';
import { SensorAdapterFactory } from './sensor-adapter.factory';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { HeartRateHistoricAdapter } from './sensor-historic/heart-rate-historic.adapter';
import { PulseOximeterHistoricAdapter } from './sensor-historic/pulse-oximeter-historic.adapter';
import { EnvironmentalHistoricAdapter } from './sensor-historic/environmental-historic.adapter';
import { HealthThermometerHistoricAdapter } from './sensor-historic/health-thermometer-historic.adapter';
import { EcgHistoricAdapter } from './sensor-historic/ecg-historic.adapter';
import { HeartRateLiveAdapter } from './sensor-live/heart-rate-live.adapter';
import { PulseOximeterLiveAdapter } from './sensor-live/pulse-oximeter-live.adapter';
import { EnvironmentalLiveAdapter } from './sensor-live/environmental-live.adapter';
import { HealthThermometerLiveAdapter } from './sensor-live/health-thermometer-live.adapter';
import { EcgLiveAdapter } from './sensor-live/ecg-live.adapter';

describe('SensorAdapterFactory', () => {
  const factory = new SensorAdapterFactory();

  describe('createHistoricAdapter', () => {
    it.each([
      [SensorProfiles.HEART_RATE_SERVICE, HeartRateHistoricAdapter],
      [SensorProfiles.PULSE_OXIMETER_SERVICE, PulseOximeterHistoricAdapter],
      [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE, EnvironmentalHistoricAdapter],
      [SensorProfiles.HEALTH_THERMOMETER_SERVICE, HealthThermometerHistoricAdapter],
      [SensorProfiles.CUSTOM_ECG_SERVICE, EcgHistoricAdapter],
    ] as const)('should return %s instance for profile "%s"', (profile, AdapterClass) => {
      expect(factory.createHistoricAdapter(profile)).toBeInstanceOf(AdapterClass);
    });

    it('should return a new instance on each call', () => {
      const a = factory.createHistoricAdapter(SensorProfiles.HEART_RATE_SERVICE);
      const b = factory.createHistoricAdapter(SensorProfiles.HEART_RATE_SERVICE);
      expect(a).not.toBe(b);
    });

    it('should throw for an unknown profile', () => {
      expect(() => factory.createHistoricAdapter('unknown' as SensorProfiles)).toThrow(
        'No historic adapter registered for profile: unknown'
      );
    });
  });

  describe('createLiveAdapter', () => {
    it.each([
      [SensorProfiles.HEART_RATE_SERVICE, HeartRateLiveAdapter],
      [SensorProfiles.PULSE_OXIMETER_SERVICE, PulseOximeterLiveAdapter],
      [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE, EnvironmentalLiveAdapter],
      [SensorProfiles.HEALTH_THERMOMETER_SERVICE, HealthThermometerLiveAdapter],
      [SensorProfiles.CUSTOM_ECG_SERVICE, EcgLiveAdapter],
    ] as const)('should return %s instance for profile "%s"', (profile, AdapterClass) => {
      expect(factory.createLiveAdapter(profile)).toBeInstanceOf(AdapterClass);
    });

    it('should return a new instance on each call', () => {
      const a = factory.createLiveAdapter(SensorProfiles.HEART_RATE_SERVICE);
      const b = factory.createLiveAdapter(SensorProfiles.HEART_RATE_SERVICE);
      expect(a).not.toBe(b);
    });

    it('should throw for an unknown profile', () => {
      expect(() => factory.createLiveAdapter('unknown' as SensorProfiles)).toThrow(
        'No live adapter registered for profile: unknown'
      );
    });
  });
});
