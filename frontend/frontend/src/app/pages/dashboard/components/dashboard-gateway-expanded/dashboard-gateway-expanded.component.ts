import { Component, input, output } from '@angular/core';
import { PageEvent } from '@angular/material/paginator';

import { DashboardSensorTableComponent } from '../dashboard-sensor-table/dashboard-sensor-table.component';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { ChartRequest } from '../../../../models/chart/chart-request.model';

@Component({
  selector: 'app-dashboard-gateway-expanded',
  imports: [DashboardSensorTableComponent],
  templateUrl: './dashboard-gateway-expanded.component.html',
  styleUrl: './dashboard-gateway-expanded.component.css',
})
export class DashboardGatewayExpandedComponent {
  public readonly sensors = input.required<Sensor[]>();
  public readonly gateway = input.required<Gateway>();
  public readonly loading = input<boolean>();

  public readonly sensorTotal = input<number>(0);
  public readonly sensorPageIndex = input<number>(0);
  public readonly sensorLimit = input<number>(10);

  public readonly chartRequested = output<ChartRequest>();

  public readonly sensorPageChange = output<PageEvent>();
}
