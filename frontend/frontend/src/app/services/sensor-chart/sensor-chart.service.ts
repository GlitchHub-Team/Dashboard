import { inject, Injectable, signal } from '@angular/core';
import { retry, Subscription, timer } from 'rxjs';

import { SensorLiveReadingsApiService } from '../sensor-live-api/sensor-live-readings-api.service';
import { SensorHistoricApiService } from '../sensor-historic-api/sensor-historic-api.service';
import { SensorHistoricAdapter } from '../../adapters/sensor-historic/sensor-historic.adapter';
import { SensorLiveReadingAdapter } from '../../adapters/sensor-live/sensor-live-reading.adapter';
import { SensorReading } from '../../models/sensor-data/sensor-reading.model';
import { ChartRequest } from '../../models/chart/chart-request.model';
import { ChartType } from '../../models/chart/chart-type.enum';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class SensorChartService {
  private readonly historicService = inject(SensorHistoricApiService);
  private readonly liveReadingsService = inject(SensorLiveReadingsApiService);
  private readonly historicAdapter = inject(SensorHistoricAdapter);
  private readonly liveReadingsAdapter = inject(SensorLiveReadingAdapter);

  private readonly MAX_LIVE_READINGS = 50;
  private subscription: Subscription | null = null;

  private readonly _historicReadings = signal<SensorReading[]>([]);
  private readonly _liveReadings = signal<SensorReading[]>([]);
  private readonly _loading = signal(false);
  // Status scelti in maniera arbitraria
  private readonly _connectionStatus = signal<
    'connected' | 'connecting' | 'disconnected' | 'reconnecting'
  >('disconnected');
  private readonly _error = signal<string | null>(null);

  public readonly historicReadings = this._historicReadings.asReadonly();
  public readonly liveReadings = this._liveReadings.asReadonly();
  public readonly loading = this._loading.asReadonly();
  public readonly connectionStatus = this._connectionStatus.asReadonly();
  public readonly error = this._error.asReadonly();

  public startChart(req: ChartRequest): void {
    this.reset();

    if (req.chartType === ChartType.HISTORIC) {
      this.startHistoricChart(req);
    } else {
      this.startLiveReadingsChart(req);
    }
  }

  // Cleanup di tutte le connessioni/grafici
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
        // Adapting della response al formato utilizzato dal grafico
        const historicData = this.historicAdapter.fromResponse(response);
        this._historicReadings.set(historicData.readings);
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
      .connect(req.sensor)
      .pipe(
        // Se cade la connessione, ritenta 3 volte con 3s di delay di riconnettersi
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
        // Ogni volta che arriva una nuova lettura, la adatta e la aggiunge alla lista di quelle visualizzate
        next: (raw) => {
          this._connectionStatus.set('connected');
          this._error.set(null);
          const reading = this.liveReadingsAdapter.fromDTO(raw);
          this._liveReadings.update((readings) => {
            // Mantiene solo le ultime MAX_LIVE_READINGS letture per evitare di sovraccaricare il grafico
            const updated = [...readings, reading];
            return updated.length > this.MAX_LIVE_READINGS
              ? updated.slice(updated.length - this.MAX_LIVE_READINGS)
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
    this._loading.set(false);
    this._connectionStatus.set('disconnected');
    this._error.set(null);
  }
}
