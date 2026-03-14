import { Component, input, output } from '@angular/core';
import { MatIcon } from '@angular/material/icon';

import { Sensor } from '../../../../models/sensor.model';
import { ChartRequest } from '../../../../models/chart-request.model';
import { ChartType } from '../../../../models/chart-type.enum';

@Component({
  selector: 'app-dashboard-sensor-row',
  imports: [MatIcon],
  templateUrl: './dashboard-sensor-row.component.html',
  styleUrl: './dashboard-sensor-row.component.css',
})
export class DashboardSensorRowComponent {
  public readonly sensor = input.required<Sensor>();

  protected readonly ChartType = ChartType;
  public readonly chartRequested = output<ChartRequest>();

  protected onViewChart(chartType: ChartType): void {
    this.chartRequested.emit({
      sensor: this.sensor(),
      chartType,
      timeInterval: null!,
    });
  }
}
