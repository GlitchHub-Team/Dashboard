import { Component, computed, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatTooltip } from '@angular/material/tooltip';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { UpperCasePipe } from '@angular/common';

import { DashboardGatewayExpandedComponent } from '../dashboard-gateway-expanded/dashboard-gateway-expanded.component';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ActionMode } from '../../../../models/action-mode.model';

@Component({
  selector: 'app-dashboard-gateway-table',
  imports: [
    MatProgressSpinner,
    MatIcon,
    MatTableModule,
    MatTooltip,
    MatPaginatorModule,
    UpperCasePipe,
    DashboardGatewayExpandedComponent,
    MatButtonModule,
  ],
  templateUrl: './dashboard-gateway-table.component.html',
  styleUrl: './dashboard-gateway-table.component.css',
})
export class DashboardGatewayTableComponent {
  public readonly gateways = input.required<Gateway[]>();
  public readonly sensors = input.required<Sensor[]>();
  public readonly expandedGateway = input<Gateway | null>(null);
  public readonly actionMode = input<ActionMode>('dashboard');
  public readonly gatewayLoading = input<boolean>();
  public readonly sensorLoading = input<boolean>();

  public readonly gatewayTotal = input<number>(0);
  public readonly gatewayPageIndex = input<number>(0);
  public readonly gatewayLimit = input<number>(10);

  public readonly sensorTotal = input<number>(0);
  public readonly sensorPageIndex = input<number>(0);
  public readonly sensorLimit = input<number>(10);

  public readonly commandRequested = output<Gateway>();
  public readonly chartRequested = output<ChartRequest>();
  public readonly expandedGatewayChange = output<Gateway>();

  public readonly gatewayDeleteRequested = output<Gateway>();
  public readonly gatewayCreateRequested = output<void>();
  public readonly sensorDeleteRequested = output<Sensor>();
  public readonly sensorCreateRequested = output<Gateway>();

  public readonly gatewayPageChange = output<PageEvent>();
  public readonly sensorPageChange = output<PageEvent>();

  private readonly columns = ['id', 'tenantId', 'name', 'status'];
  protected readonly displayedColumns = computed(() => {
    switch (this.actionMode()) {
      case 'manage':
        return [...this.columns, 'delete'];
      default:
        return [...this.columns, 'commands'];
    }
  });

  protected isExpanded(gateway: Gateway): boolean {
    return this.expandedGateway()?.id === gateway.id;
  }

  protected onGatewayPageChange(event: PageEvent): void {
    this.gatewayPageChange.emit(event);
  }
}
