import { Component, computed, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTooltip } from '@angular/material/tooltip';

import { DashboardGatewayExpandedComponent } from '../dashboard-gateway-expanded/dashboard-gateway-expanded.component';
import { Gateway } from '../../../../models/gateway.model';
import { Sensor } from '../../../../models/sensor.model';
import { ChartRequest } from '../../../../models/chart-request.model';

@Component({
  selector: 'app-dashboard-gateway-table',
  imports: [
    MatProgressSpinner,
    MatIcon,
    MatTableModule,
    MatTooltip,
    DashboardGatewayExpandedComponent,
  ],
  templateUrl: './dashboard-gateway-table.component.html',
  styleUrl: './dashboard-gateway-table.component.css',
})
export class DashboardGatewayTableComponent {
  public readonly gateways = input.required<Gateway[]>();
  public readonly sensors = input.required<Sensor[]>();
  public readonly expandedGateway = input<Gateway | null>();
  public readonly canSendCommands = input<boolean>();
  public readonly gatewayLoading = input<boolean>();
  public readonly sensorLoading = input<boolean>();

  public readonly commandRequested = output<Gateway>();
  public readonly chartRequested = output<ChartRequest>();
  public readonly expandedGatewayChange = output<Gateway>();

  private readonly columns = ['id'];
  protected readonly displayedColumns = computed(() =>
    this.columns.concat(this.canSendCommands() ? ['commands'] : []),
  );

  protected isExpanded(gateway: Gateway): boolean {
    return this.expandedGateway()?.id === gateway.id;
  }
}
