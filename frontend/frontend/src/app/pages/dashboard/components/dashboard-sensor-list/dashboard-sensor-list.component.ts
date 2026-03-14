import { Component, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';

import { DashboardSensorRowComponent } from '../dashboard-sensor-row.component/dashboard-sensor-row.component';
import { Sensor } from '../../../../models/sensor.model';
import { ChartRequest } from '../../../../models/chart-request.model';

@Component({
  selector: 'app-dashboard-sensor-list',
  imports: [DashboardSensorRowComponent, MatProgressSpinner, MatIcon],
  templateUrl: './dashboard-sensor-list.component.html',
  styleUrl: './dashboard-sensor-list.component.css',
})
export class DashboardSensorListComponent {
  public readonly sensors = input.required<Sensor[]>();
  public readonly loading = input<boolean>();

  public readonly chartRequested = output<ChartRequest>();
}
