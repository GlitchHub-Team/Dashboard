import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { Sensor } from '../models/sensor/sensor.model';
import { TimeInterval } from '../models/time-interval.model';
import { HistoricResponse } from '../models/sensor-data/historic-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';

@Injectable()
export class SensorHistoricMockService {
  private readonly DEFAULT_HOURS = 24;

  getHistoricData(sensor: Sensor, timeInterval: TimeInterval): Observable<HistoricResponse> {
    const interval = timeInterval ?? this.getDefaultInterval();
    const from = interval.from.getTime();
    const to = interval.to.getTime();
    const resolution = 60; // 60 seconds between points
    const count = Math.floor((to - from) / (resolution * 1000));

    const baseValue = this.getBaseValue(sensor.profile);
    const amplitude = this.getAmplitude(sensor.profile);

    const data = Array.from({ length: count }, (_, i) => {
      const timestamp = new Date(from + i * resolution * 1000).toISOString();
      const noise = (Math.random() - 0.5) * amplitude * 0.2;
      const value = baseValue + amplitude * Math.sin((2 * Math.PI * i) / count) + noise;

      return {
        sensorId: sensor.id,
        timestamp,
        profile: sensor.profile,
        value: Math.round(value * 100) / 100,
      };
    });

    const response: HistoricResponse = {
      count,
      resolution,
      data,
    };

    return of(response).pipe(delay(800));
  }

  private getDefaultInterval(): TimeInterval {
    const to = new Date();
    const from = new Date(to.getTime() - this.DEFAULT_HOURS * 60 * 60 * 1000); // 24 hours ago
    return { from, to };
  }

  private getBaseValue(profile: SensorProfiles): number {
    switch (profile) {
      case SensorProfiles.HEART_RATE_SERVICE:
        return 75;
      case SensorProfiles.PULSE_OXIMETER_SERVICE:
        return 97;
      case SensorProfiles.CUSTOM_ECG_SERVICE:
        return 0;
      case SensorProfiles.HEALTH_THERMOMETER_SERVICE:
        return 36.6;
      case SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE:
        return 22;
    }
  }

  private getAmplitude(profile: SensorProfiles): number {
    switch (profile) {
      case SensorProfiles.HEART_RATE_SERVICE:
        return 15;
      case SensorProfiles.PULSE_OXIMETER_SERVICE:
        return 3;
      case SensorProfiles.CUSTOM_ECG_SERVICE:
        return 1.5;
      case SensorProfiles.HEALTH_THERMOMETER_SERVICE:
        return 0.5;
      case SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE:
        return 5;
    }
  }
}
