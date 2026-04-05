// mocks/sensor-historic-mock.service.ts
import { Injectable } from '@angular/core';
import { Observable, of, delay, throwError } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { ChartRequest } from '../models/chart/chart-request.model';
import { HistoricResponse, HistoricSample } from '../models/sensor-data/historic-response.model';
import { TimeInterval } from '../models/chart/time-interval.model';
import { SensorProfiles } from '../models/sensor/sensor-profiles.enum';
import { getMockFieldConfigs } from './sensor-mock.config';

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
            error: { error: 'request failed' },
          }),
      ).pipe(delay(500));
    }

    const isEcg = req.sensor.profile === SensorProfiles.CUSTOM_ECG_SERVICE;
    const samples = isEcg
      ? this.generateEcgSamples(req)
      : this.generateScalarSamples(req);

    const response: HistoricResponse = {
      count: samples.length,
      samples,
    };

    return of(response).pipe(delay(800));
  }

  private generateScalarSamples(req: ChartRequest): HistoricSample[] {
    const defaultInterval = this.getDefaultInterval();
    const from = (req.timeInterval?.from ?? defaultInterval.from).getTime();
    const to = (req.timeInterval?.to ?? defaultInterval.to).getTime();
    const durationMs = to - from;
    const resolution = 60_000;
    const totalPoints = Math.floor(durationMs / resolution);
    const count = req.dataPointsCounter
      ? Math.min(req.dataPointsCounter, totalPoints)
      : totalPoints;
    const step = totalPoints / count;

    const fieldConfigs = getMockFieldConfigs(req.sensor.profile);
    const samples: HistoricSample[] = [];

    for (let i = 0; i < count; i++) {
      const timestamp = from + Math.floor(i * step) * resolution;
      const data: Record<string, number> = {};

      fieldConfigs.forEach((fc) => {
        const noise = (Math.random() - 0.5) * fc.amplitude * 0.2;
        const raw =
          fc.baseValue +
          fc.amplitude * Math.sin((2 * Math.PI * i) / count) +
          noise;
        data[fc.key] = Math.min(
          fc.max,
          Math.max(fc.min, Math.round(raw * 100) / 100),
        );
      });

      samples.push({
        sensor_id: req.sensor.id,
        gateway_id: req.sensor.gatewayId,
        tenant_id: 'mock-tenant',
        timestamp: new Date(timestamp).toISOString(),
        profile: req.sensor.profile,
        data: data,
      });
    }

    return samples;
  }

  private generateEcgSamples(req: ChartRequest): HistoricSample[] {
    const defaultInterval = this.getDefaultInterval();
    const from = (req.timeInterval?.from ?? defaultInterval.from).getTime();
    const to = (req.timeInterval?.to ?? defaultInterval.to).getTime();

    // Each sample = 1 second = 250 waveform values
    const totalSeconds = Math.floor((to - from) / 1000);
    const count = req.dataPointsCounter
      ? Math.min(req.dataPointsCounter, totalSeconds)
      : Math.min(totalSeconds, 60);  // cap at 60 seconds for mock

    const samples: HistoricSample[] = [];

    for (let s = 0; s < count; s++) {
      const timestamp = from + s * 1000;
      const waveform = this.generateEcgWaveform(s);

      samples.push({
        sensor_id: req.sensor.id,
        gateway_id: req.sensor.gatewayId,
        tenant_id: 'mock-tenant',
        timestamp: new Date(timestamp).toISOString(),
        profile: req.sensor.profile,
        data: { Waveform: waveform },
      });
    }

    return samples;
  }

  /**
   * Generates a somewhat realistic ECG waveform for 1 second (250 samples).
   * Simulates the PQRST pattern repeating ~once per second.
   */
  private generateEcgWaveform(secondIndex: number): number[] {
    const waveform: number[] = [];
    const samplesPerSecond = 250;

    for (let i = 0; i < samplesPerSecond; i++) {
      const t = i / samplesPerSecond;  // 0.0 → 1.0 within this second
      let value = 0;

      // Baseline noise
      value += (Math.random() - 0.5) * 20;

      // P wave (small bump around t=0.1)
      value += 80 * Math.exp(-Math.pow((t - 0.1) * 20, 2));

      // QRS complex (sharp spike around t=0.3)
      value -= 60 * Math.exp(-Math.pow((t - 0.28) * 40, 2));  // Q dip
      value += 900 * Math.exp(-Math.pow((t - 0.32) * 50, 2));  // R peak
      value -= 120 * Math.exp(-Math.pow((t - 0.36) * 40, 2));  // S dip

      // T wave (broad bump around t=0.55)
      value += 150 * Math.exp(-Math.pow((t - 0.55) * 12, 2));

      waveform.push(Math.round(value));
    }

    return waveform;
  }

  private getDefaultInterval(): TimeInterval {
    const to = new Date();
    const from = new Date(to.getTime() - this.DEFAULT_HOURS * 60 * 60 * 1000);
    return { from, to };
  }
}