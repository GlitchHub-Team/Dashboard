// mocks/sensor-real-time-mock.service.ts
import { Injectable } from '@angular/core';
import { Observable, interval, map, Subject, takeUntil, switchMap, throwError, timer } from 'rxjs';
import { ChartRequest } from '../models/chart/chart-request.model';
import { RealTimeReading } from '../models/sensor-data/real-time-reading.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { getMockFieldConfigs, MockFieldConfig } from './sensor-mock.config';
import { ApiError } from '../models/api-error.model';

@Injectable()
export class SensorRealTimeMockService {
  private readonly stop$ = new Subject<void>();
  private currentValues = new Map<string, number>();
  private readingCount = 0;

  private readonly shouldFailConnection = false;
  private readonly shouldDisconnectAfter = 0;
  public samplesPerSecond = 250;

  connect(req: ChartRequest): Observable<RealTimeReading> {
    this.readingCount = 0;
    this.currentValues.clear();

    const sensor = req.sensor;

    if (this.shouldFailConnection) {
      return this.delayedError(0, 'WebSocket connection failed');
    }

    const isEcg = sensor.profile === SensorProfiles.CUSTOM_ECG_SERVICE;

    if (!isEcg) {
      const fieldConfigs = getMockFieldConfigs(sensor.profile);
      fieldConfigs.forEach((fc) => this.currentValues.set(fc.key, fc.baseValue));
    }

    return interval(1000).pipe(
      takeUntil(this.stop$),
      map(() => {
        this.readingCount++;

        if (this.shouldDisconnectAfter > 0 && this.readingCount >= this.shouldDisconnectAfter) {
          throw { status: 0, message: 'WebSocket connection lost' } as ApiError;
        }

        const data = isEcg
          ? { Waveform: this.generateEcgWaveform() }
          : this.generateScalarData(getMockFieldConfigs(sensor.profile));

        return {
          timestamp: new Date().toISOString(),
          profile: sensor.profile,
          data,
        };
      }),
    );
  }

  disconnect(): void {
    this.stop$.next();
  }

  private generateScalarData(fieldConfigs: MockFieldConfig[]): Record<string, number> {
    const data: Record<string, number> = {};

    fieldConfigs.forEach((fc) => {
      const current = this.currentValues.get(fc.key)!;
      const delta = (Math.random() - 0.5) * fc.step;
      const next = Math.min(fc.max, Math.max(fc.min, current + delta));
      const rounded = Math.round(next * 100) / 100;

      this.currentValues.set(fc.key, rounded);
      data[fc.key] = rounded;
    });

    return data;
  }

  private generateEcgWaveform(): number[] {
    const waveform: number[] = [];
    const samplesPerSecond = this.samplesPerSecond;

    for (let i = 0; i < samplesPerSecond; i++) {
      const t = i / samplesPerSecond;
      let value = 0;

      value += (Math.random() - 0.5) * 20;
      value += 80 * Math.exp(-Math.pow((t - 0.1) * 20, 2));
      value -= 60 * Math.exp(-Math.pow((t - 0.28) * 40, 2));
      value += 900 * Math.exp(-Math.pow((t - 0.32) * 50, 2));
      value -= 120 * Math.exp(-Math.pow((t - 0.36) * 40, 2));
      value += 150 * Math.exp(-Math.pow((t - 0.55) * 12, 2));

      waveform.push(Math.round(value));
    }

    return waveform;
  }

  private delayedError(status: number, message: string): Observable<never> {
    return timer(500).pipe(switchMap(() => throwError(() => ({ status, message }) as ApiError)));
  }
}
