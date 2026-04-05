import { Injectable } from '@angular/core';

import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { SensorHistoricAdapter } from './sensor-historic/sensor-historic.adapter';
import { SensorLiveReadingAdapter } from './sensor-live/sensor-live-reading.adapter';
import { HeartRateHistoricAdapter } from './sensor-historic/heart-rate-historic.adapter';
import { HeartRateLiveAdapter } from './sensor-live/heart-rate-live.adapter';
import { PulseOximeterHistoricAdapter } from './sensor-historic/pulse-oximeter-historic.adapter';
import { PulseOximeterLiveAdapter } from './sensor-live/pulse-oximeter-live.adapter';
import { EnvironmentalHistoricAdapter } from './sensor-historic/environmental-historic.adapter';
import { EnvironmentalLiveAdapter } from './sensor-live/environmental-live.adapter';
import { HealthThermometerHistoricAdapter } from './sensor-historic/health-thermometer-historic.adapter';
import { HealthThermometerLiveAdapter } from './sensor-live/health-thermometer-live.adapter';
import { EcgHistoricAdapter } from './sensor-historic/ecg-historic.adapter';
import { EcgLiveAdapter } from './sensor-live/ecg-live.adapter';

@Injectable({ providedIn: 'root' })
export class SensorAdapterFactory {
  private readonly historicAdapters: Record<string, () => SensorHistoricAdapter> = {
    [SensorProfiles.HEART_RATE_SERVICE]: () => new HeartRateHistoricAdapter(),
    [SensorProfiles.PULSE_OXIMETER_SERVICE]: () => new PulseOximeterHistoricAdapter(),
    [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: () => new EnvironmentalHistoricAdapter(),
    [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: () => new HealthThermometerHistoricAdapter(),
    [SensorProfiles.CUSTOM_ECG_SERVICE]: () => new EcgHistoricAdapter()
  };

  private readonly liveAdapters: Record<string, () => SensorLiveReadingAdapter> = {
    [SensorProfiles.HEART_RATE_SERVICE]: () => new HeartRateLiveAdapter(),
    [SensorProfiles.PULSE_OXIMETER_SERVICE]: () => new PulseOximeterLiveAdapter(),
    [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE]: () => new EnvironmentalLiveAdapter(),
    [SensorProfiles.HEALTH_THERMOMETER_SERVICE]: () => new HealthThermometerLiveAdapter(),
    [SensorProfiles.CUSTOM_ECG_SERVICE]: () => new EcgLiveAdapter(),
  };

  createHistoricAdapter(profile: SensorProfiles): SensorHistoricAdapter {
    const factory = this.historicAdapters[profile];
    if (!factory) {
      throw new Error(`No historic adapter registered for profile: ${profile}`);
    }
    return factory();
  }

  createLiveAdapter(profile: SensorProfiles): SensorLiveReadingAdapter {
    const factory = this.liveAdapters[profile];
    if (!factory) {
      throw new Error(`No live adapter registered for profile: ${profile}`);
    }
    return factory();
  }
}