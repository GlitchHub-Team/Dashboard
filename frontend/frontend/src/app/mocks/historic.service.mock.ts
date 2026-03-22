import { Injectable } from '@angular/core';
import { Observable, of, delay } from 'rxjs';
import { ChartRequest } from '../models/chart/chart-request.model';
import { HistoricResponse } from '../models/sensor-data/historic-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { TimeInterval } from '../models/time-interval.model';

@Injectable()
export class SensorHistoricMockService {
  private readonly DEFAULT_HOURS = 24;

  getHistoricData(req: ChartRequest): Observable<HistoricResponse> {
    const interval = req.timeInterval ?? this.getDefaultInterval();
    const from = interval.from.getTime();
    const to = interval.to.getTime();
    const resolution = 60;
    const totalPoints = Math.floor((to - from) / (resolution * 1000));
    const count = req.dataPointsCounter
      ? Math.min(req.dataPointsCounter, totalPoints)
      : totalPoints;
    const step = totalPoints / count;

    const baseValue = this.getBaseValue(req.sensor.profile);
    const amplitude = this.getAmplitude(req.sensor.profile);
    const lowerBound = req.valuesInterval?.lowerBound ?? -Infinity;
    const upperBound = req.valuesInterval?.upperBound ?? Infinity;

    const data = Array.from({ length: count }, (_, i) => {
      const timestamp = new Date(from + Math.floor(i * step) * resolution * 1000).toISOString();
      const noise = (Math.random() - 0.5) * amplitude * 0.2;
      const raw = baseValue + amplitude * Math.sin((2 * Math.PI * i) / count) + noise;
      const value = Math.min(upperBound, Math.max(lowerBound, Math.round(raw * 100) / 100));

      return {
        sensorId: req.sensor.id,
        timestamp,
        profile: req.sensor.profile,
        value,
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
