import { Component, computed, inject, OnInit } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatIcon } from '@angular/material/icon';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';

import { DashboardService } from '../../services/dashboard/dashboard.service';
import { DashboardGatewayTableComponent } from './components/dashboard-gateway-table/dashboard-gateway-table.component';
import { DashboardSensorTableComponent } from './components/dashboard-sensor-table/dashboard-sensor-table.component';
import { Gateway } from '../../models/gateway/gateway.model';
import { ChartRequest } from '../../models/chart-request.model';

@Component({
  selector: 'app-dashboard',
  imports: [DashboardGatewayTableComponent, DashboardSensorTableComponent, MatIcon],
  templateUrl: './dashboard.page.html',
  styleUrl: './dashboard.page.css',
})
export class DashboardPage implements OnInit {
  private readonly dashboardService = inject(DashboardService);
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);

  protected readonly gatewayList = this.dashboardService.gatewayList;
  protected readonly gatewayTotal = this.dashboardService.gatewayTotal;
  protected readonly gatewayPageIndex = this.dashboardService.gatewayPageIndex;
  protected readonly gatewayLimit = this.dashboardService.gatewayLimit;
  protected readonly gatewayLoading = this.dashboardService.gatewayLoading;

  protected readonly sensorList = this.dashboardService.sensorList;
  protected readonly sensorTotal = this.dashboardService.sensorTotal;
  protected readonly sensorPageIndex = this.dashboardService.sensorPageIndex;
  protected readonly sensorLimit = this.dashboardService.sensorLimit;
  protected readonly sensorLoading = this.dashboardService.sensorLoading;

  protected readonly expandedGateway = this.dashboardService.expandedGateway;
  protected readonly selectedChart = this.dashboardService.selectedChart;
  protected readonly canSendCommands = this.dashboardService.canSendCommands;
  protected readonly error = computed(
    () => this.dashboardService.gatewayError() ?? this.dashboardService.sensorError(),
  );

  public ngOnInit(): void {
    this.dashboardService.loadDashboard();
  }

  protected onExpandedGatewayChange(gateway: Gateway): void {
    this.dashboardService.toggleExpandedGateway(gateway);
  }

  protected onGatewayPageChange(event: PageEvent): void {
    this.dashboardService.changeGatewayPage(event.pageIndex, event.pageSize);
  }

  protected onSensorPageChange(event: PageEvent): void {
    this.dashboardService.changeSensorPage(event.pageIndex, event.pageSize);
  }

  protected onCommandRequested(gateway: Gateway): void {
    this.snackBar.open(gateway.id, 'Close', { duration: 2000 });
  }

  protected onChartOpen(request: ChartRequest): void {
    this.dashboardService.openChart(request);
  }

  protected onChartClosed(): void {
    this.dashboardService.closeChart();
  }
}
