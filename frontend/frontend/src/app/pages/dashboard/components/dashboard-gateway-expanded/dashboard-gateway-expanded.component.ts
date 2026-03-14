import { Component, input, output } from '@angular/core';

import { DashboardSensorListComponent } from '../dashboard-sensor-list/dashboard-sensor-list.component';
import { Sensor } from '../../../../models/sensor.model';
import { Gateway } from '../../../../models/gateway.model';
import { ChartRequest } from '../../../../models/chart-request.model';

@Component({
  selector: 'app-dashboard-gateway-expanded',
  imports: [DashboardSensorListComponent],
  templateUrl: './dashboard-gateway-expanded.component.html',
  styleUrl: './dashboard-gateway-expanded.component.css',
})
export class DashboardGatewayExpandedComponent {
  public readonly sensors = input.required<Sensor[]>();
  public readonly gateway = input.required<Gateway>();
  public readonly loading = input<boolean>();

  public readonly chartRequested = output<ChartRequest>();
}
