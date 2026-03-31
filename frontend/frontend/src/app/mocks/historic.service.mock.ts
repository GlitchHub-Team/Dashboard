import { Injectable } from '@angular/core';
import { Observable, of, delay, throwError } from 'rxjs';
import { ChartRequest } from '../models/chart/chart-request.model';
import { HistoricResponse } from '../models/sensor-data/historic-response.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { TimeInterval } from '../models/chart/time-interval.model';
import { HttpErrorResponse } from '@angular/common/http';

@Injectable()
export class SensorHistoricMockService {
  private readonly DEFAULT_HOURS = 24;

  getHistoricData(req: ChartRequest): Observable<HistoricResponse> {
    const shouldFail = false;

    if (shouldFail) {
      return throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            statusText: 'Bad Request',
            error: { error: 'tenant already exists' },
          }),
      ).pipe(delay(500));
    }
    const defaultInterval = this.getDefaultInterval();
    const from = (req.timeInterval?.from ?? defaultInterval.from).getTime();
    const to = (req.timeInterval?.to ?? defaultInterval.to).getTime();
    const durationMs = to - from;
    const resolution = 60_000; // 60 seconds in ms
    const totalPoints = Math.floor(durationMs / resolution);
    const count = req.dataPointsCounter
      ? Math.min(req.dataPointsCounter, totalPoints)
      : totalPoints;
    const step = totalPoints / count;

    const baseValue = this.getBaseValue(req.sensor.profile);
    const amplitude = this.getAmplitude(req.sensor.profile);
    const lowerBound = req.valuesInterval?.lowerBound ?? -Infinity;
    const upperBound = req.valuesInterval?.upperBound ?? Infinity;
    const unit = this.getUnit(req.sensor.profile);

    const timestamps: number[] = [];
    const values: number[] = [];

    for (let i = 0; i < count; i++) {
      timestamps.push(from + Math.floor(i * step) * resolution);
      const noise = (Math.random() - 0.5) * amplitude * 0.2;
      const raw = baseValue + amplitude * Math.sin((2 * Math.PI * i) / count) + noise;
      values.push(Math.min(upperBound, Math.max(lowerBound, Math.round(raw * 100) / 100)));
    }

    const response: HistoricResponse = {
      count: { current: count, real: count, total: totalPoints },
      duration: durationMs,
      dataset: { timestamps, values },
      unit,
    };

    return of(response).pipe(delay(800));
  }

  private getDefaultInterval(): TimeInterval {
    const to = new Date();
    const from = new Date(to.getTime() - this.DEFAULT_HOURS * 60 * 60 * 1000);
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

  private getUnit(profile: SensorProfiles): string {
    switch (profile) {
      case SensorProfiles.HEART_RATE_SERVICE:
        return 'bpm';
      case SensorProfiles.PULSE_OXIMETER_SERVICE:
        return '%';
      case SensorProfiles.CUSTOM_ECG_SERVICE:
        return 'mV';
      case SensorProfiles.HEALTH_THERMOMETER_SERVICE:
        return '°C';
      case SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE:
        return '°C';
    }
  }
}
