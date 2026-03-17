import { Component, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTooltip } from '@angular/material/tooltip';
import { TitleCasePipe } from '@angular/common';

import { Sensor } from '../../../../models/sensor.model';
import { ChartRequest } from '../../../../models/chart-request.model';
import { ChartType } from '../../../../models/chart-type.enum';

@Component({
  selector: 'app-dashboard-sensor-table',
  imports: [MatProgressSpinner, MatTableModule, MatTooltip, MatIcon, TitleCasePipe],
  templateUrl: './dashboard-sensor-table.component.html',
  styleUrl: './dashboard-sensor-table.component.css',
})
export class DashboardSensorTableComponent {
  public readonly sensors = input.required<Sensor[]>();
  public readonly loading = input<boolean>();

  protected readonly displayedColumns = ['id', 'gatewayId', 'name', 'profile', 'actions'];

  protected readonly ChartType = ChartType;
  public readonly chartRequested = output<ChartRequest>();

  protected onViewChart(sensor: Sensor, chartType: ChartType): void {
    this.chartRequested.emit({
      sensor,
      chartType,
      timeInterval: null!,
    });
  }
}
