import { Component, computed, effect, inject, input, OnDestroy, output } from '@angular/core';
import { MatIcon } from '@angular/material/icon';
import { MatIconButton } from '@angular/material/button';
import { MatCard, MatCardContent, MatCardHeader, MatCardTitle } from '@angular/material/card';
import { MatProgressSpinner } from '@angular/material/progress-spinner';

import { HistoricChartComponent } from './components/historic-chart/historic-chart.component';
import { SensorChartService } from '../../../../services/sensor-chart/sensor-chart.service';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { RealTimeChartComponent } from './components/real-time-chart/real-time-chart.component';
import { getSensorProfileDisplay } from '../../../../models/chart/sensor-profile-display.model';

@Component({
  selector: 'app-chart-container',
  imports: [
    MatCard,
    MatCardContent,
    MatCardHeader,
    MatCardTitle,
    MatIconButton,
    MatIcon,
    MatProgressSpinner,
    HistoricChartComponent,
    RealTimeChartComponent,
  ],
  templateUrl: './chart-container.component.html',
  styleUrl: './chart-container.component.css',
})
export class ChartContainerComponent implements OnDestroy {
  private readonly chartService = inject(SensorChartService);

  public chartRequest = input<ChartRequest | null>(null);
  public chartClosed = output<void>();

  protected readonly historicReadings = this.chartService.historicReadings;
  protected readonly liveReadings = this.chartService.liveReadings;
  protected readonly loading = this.chartService.loading;
  protected readonly connectionStatus = this.chartService.connectionStatus;
  protected readonly error = this.chartService.error;

  protected readonly isHistoricChart = computed(() => {
    return this.chartRequest()?.chartType === ChartType.HISTORIC;
  });

  protected readonly isLiveChart = computed(() => {
    return this.chartRequest()?.chartType === ChartType.REALTIME;
  });

  protected readonly statusLabel = computed(() => {
    switch (this.connectionStatus()) {
      case 'connected':
        return 'Connected';
      case 'connecting':
        return 'Connecting...';
      case 'disconnected':
        return 'Disconnected';
      case 'reconnecting':
        return 'Reconnecting...';
    }
  });

  protected readonly statusClass = computed(() => {
    switch (this.connectionStatus()) {
      case 'connected':
        return 'status-connected';
      case 'connecting':
        return 'status-connecting';
      case 'disconnected':
        return 'status-disconnected';
      case 'reconnecting':
        return 'status-reconnecting';
    }
  });

  protected readonly profileDisplay = computed(() => {
    const request = this.chartRequest();
    return request ? getSensorProfileDisplay(request.sensor.profile) : null;
  });

  constructor() {
    effect(() => {
      const req = this.chartRequest();
      if (req) {
        this.chartService.startChart(req);
      }
    });
  }

  protected onClose(): void {
    this.chartService.stopChart();
    this.chartClosed.emit();
  }

  ngOnDestroy(): void {
    this.chartService.stopChart();
  }
}
