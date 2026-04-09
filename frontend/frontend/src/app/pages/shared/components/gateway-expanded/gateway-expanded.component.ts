import { Component, input, output } from '@angular/core';
import { PageEvent } from '@angular/material/paginator';

import { SensorTableComponent } from '../sensor-table/sensor-table.component';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ActionMode } from '../../../../models/action-mode.model';

@Component({
  selector: 'app-gateway-expanded',
  imports: [SensorTableComponent],
  templateUrl: './gateway-expanded.component.html',
  styleUrl: './gateway-expanded.component.css',
})
export class GatewayExpandedComponent {
  public readonly sensors = input.required<Sensor[]>();
  public readonly gateway = input.required<Gateway>();
  public readonly loading = input<boolean>();
  public readonly actionMode = input<ActionMode>('dashboard');

  public readonly sensorTotal = input<number>(0);
  public readonly sensorPageIndex = input<number>(0);
  public readonly sensorLimit = input<number>(10);

  public readonly chartRequested = output<ChartRequest>();
  public readonly commandRequested = output<boolean>();
  // Emit del gateway associato per darlo al dialog di creazione sensore
  public readonly sensorCreateRequested = output<Gateway>();
  public readonly sensorDeleteRequested = output<Sensor>();
  public readonly sensorPageChange = output<PageEvent>();
}
