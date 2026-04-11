// services/sensor-chart/sensor-chart.service.ts
import { inject, Injectable, signal } from '@angular/core';
import { retry, Subscription, timer } from 'rxjs';

import { SensorLiveReadingsApiService } from '../sensor-live-api/sensor-live-readings-api.service';
import { SensorHistoricApiService } from '../sensor-historic-api/sensor-historic-api.service';
import { SensorAdapterFactory } from '../../adapters/sensor-adapter.factory';
import { SensorHistoricAdapter } from '../../adapters/sensor-historic/sensor-historic.adapter';
import { SensorLiveReadingAdapter } from '../../adapters/sensor-live/sensor-live-reading.adapter';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../models/sensor-data/field-descriptor.model';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { ApiError } from '../../models/api-error.model';
import { SENSOR_MAX_LIVE_READINGS } from '../../models/chart/sensor-visible-points.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

const ECG_LIVE_WINDOW_SECONDS = 2;

@Injectable({
  providedIn: 'root',
})
export class SensorChartService {
  private readonly historicService = inject(SensorHistoricApiService);
  private readonly liveReadingsService = inject(SensorLiveReadingsApiService);
  private readonly adapterFactory = inject(SensorAdapterFactory);

  private historicAdapter: SensorHistoricAdapter | null = null;
  private liveAdapter: SensorLiveReadingAdapter | null = null;
  private maxLiveReadings = 50;
  private isEcgLive = false;

  private subscription: Subscription | null = null;

  private readonly _historicReadings = signal<SensorReading[]>([]);
  private readonly _liveReadings = signal<SensorReading[]>([]);
  private readonly _fields = signal<FieldDescriptor[]>([]);
  private readonly _loading = signal(false);
  private readonly _connectionStatus = signal<
    'connected' | 'connecting' | 'disconnected' | 'reconnecting'
  >('disconnected');
  private readonly _error = signal<string | null>(null);
  private readonly _samplesPerPacket = signal<number | undefined>(undefined);

  public readonly historicReadings = this._historicReadings.asReadonly();
  public readonly liveReadings = this._liveReadings.asReadonly();
  public readonly fields = this._fields.asReadonly();
  public readonly loading = this._loading.asReadonly();
  public readonly connectionStatus = this._connectionStatus.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly samplesPerPacket = this._samplesPerPacket.asReadonly();

  public startChart(req: ChartRequest): void {
    this.reset();

    if (req.chartType === ChartType.HISTORIC) {
      this.historicAdapter = this.adapterFactory.createHistoricAdapter(req.sensor.profile);
      this._fields.set(this.historicAdapter.fields);
      this.startHistoricChart(req);
    } else {
      this.liveAdapter = this.adapterFactory.createLiveAdapter(req.sensor.profile);
      this._fields.set(this.liveAdapter.fields);
      this.isEcgLive = req.sensor.profile === SensorProfiles.CUSTOM_ECG_SERVICE;
      this.maxLiveReadings = SENSOR_MAX_LIVE_READINGS[req.sensor.profile] ?? 50;
      this.startLiveReadingsChart(req);
    }
  }

  public stopChart(): void {
    this.subscription?.unsubscribe();
    this.subscription = null;
    this.liveReadingsService.disconnect();
    this._connectionStatus.set('disconnected');
  }

  private startHistoricChart(req: ChartRequest): void {
    this._loading.set(true);

    this.subscription = this.historicService.getHistoricData(req).subscribe({
      next: (response) => {
        const historicData = this.historicAdapter!.fromResponse(response);
        this._historicReadings.set(historicData.readings);
        this._samplesPerPacket.set(historicData.samplesPerPacket);
      },
      error: (err: ApiError) => {
        this._error.set(err.message ?? 'Failed to load historic data');
        this._loading.set(false);
      },
      complete: () => this._loading.set(false),
    });
  }

  private startLiveReadingsChart(req: ChartRequest): void {
    this._connectionStatus.set('connecting');

    this.subscription = this.liveReadingsService
      .connect(req)
      .pipe(
        retry({
          count: 3,
          delay: (_, retryCount) => {
            this._connectionStatus.set('reconnecting');
            this._error.set(`Connection lost. Retry ${retryCount}/3...`);
            return timer(3000);
          },
        }),
      )
      .subscribe({
        next: (raw) => {
          this._connectionStatus.set('connected');
          this._error.set(null);
          const readings = this.liveAdapter!.fromDTO(raw);
          if (this.isEcgLive && this._samplesPerPacket() === undefined) {
            this._samplesPerPacket.set(readings.length);
            this.maxLiveReadings = Math.round(readings.length * ECG_LIVE_WINDOW_SECONDS);
          }
          this._liveReadings.update((current) => {
            const updated = [...current, ...readings];
            return updated.length > this.maxLiveReadings
              ? updated.slice(updated.length - this.maxLiveReadings)
              : updated;
          });
        },
        error: (err: ApiError) => {
          this._error.set(err.message ?? 'Failed to load live readings');
          this._connectionStatus.set('disconnected');
        },
        complete: () => {
          this._connectionStatus.set('disconnected');
        },
      });
  }

  private reset(): void {
    this.stopChart();
    this._historicReadings.set([]);
    this._liveReadings.set([]);
    this._fields.set([]);
    this._loading.set(false);
    this._connectionStatus.set('disconnected');
    this._error.set(null);
    this._samplesPerPacket.set(undefined);
    this.isEcgLive = false;
    this.historicAdapter = null;
    this.liveAdapter = null;
  }
}