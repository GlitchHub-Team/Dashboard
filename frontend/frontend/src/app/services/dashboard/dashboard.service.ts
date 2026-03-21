import { inject, Injectable, signal, computed } from '@angular/core';

import { GatewayService } from '../gateway/gateway.service';
import { SensorService } from '../sensor/sensor.service';
import { PermissionService } from '../permission/permission.service';
import { Permission } from '../../models/permission.enum';
import { Gateway } from '../../models/gateway/gateway.model';
import { ChartRequest } from '../../models/chart/chart-request.model';

@Injectable({
  providedIn: 'root',
})
export class DashboardService {
  private readonly gatewayService = inject(GatewayService);
  private readonly sensorService = inject(SensorService);
  private readonly permissionService = inject(PermissionService);

  private readonly _expandedGateway = signal<Gateway | null>(null);
  private readonly _selectedChart = signal<ChartRequest | null>(null);

  public readonly expandedGateway = this._expandedGateway.asReadonly();
  public readonly selectedChart = this._selectedChart.asReadonly();
  public readonly canSendCommands = computed(() =>
    this.permissionService.can(Permission.GATEWAY_COMMANDS),
  );

  public readonly gatewayList = this.gatewayService.gatewayList;
  public readonly gatewayTotal = this.gatewayService.total;
  public readonly gatewayPageIndex = this.gatewayService.pageIndex;
  public readonly gatewayLimit = this.gatewayService.limit;
  public readonly gatewayLoading = this.gatewayService.loading;
  public readonly gatewayError = this.gatewayService.error;

  public readonly sensorList = this.sensorService.sensorList;
  public readonly sensorTotal = this.sensorService.total;
  public readonly sensorPageIndex = this.sensorService.pageIndex;
  public readonly sensorLimit = this.sensorService.limit;
  public readonly sensorLoading = this.sensorService.loading;
  public readonly sensorError = this.sensorService.error;

  public loadDashboard(): void {
    if (this.canSendCommands()) {
      this.gatewayService.getGatewaysByTenant(
        'tenant-01',
        this.gatewayPageIndex(),
        this.gatewayLimit(),
      );
    } else {
      this.sensorService.getSensorsByTenant(
        'tenant-01',
        this.sensorPageIndex(),
        this.sensorLimit(),
      );
    }
  }

  public changeGatewayPage(pageIndex: number, limit: number): void {
    this.collapseGateway();
    this.gatewayService.changePage(pageIndex, limit);
  }

  public changeSensorPage(pageIndex: number, limit: number): void {
    this.sensorService.changePage(pageIndex, limit);
  }

  public toggleExpandedGateway(gateway: Gateway): void {
    const current = this._expandedGateway();
    if (current?.id === gateway.id) {
      this.collapseGateway();
    } else {
      this._expandedGateway.set(gateway);
      this.sensorService.getSensorsByGateway(
        gateway.id,
        this.sensorPageIndex(),
        this.sensorLimit(),
      );
    }
  }

  public openChart(request: ChartRequest): void {
    this._selectedChart.set(request);
  }

  public closeChart(): void {
    this._selectedChart.set(null);
  }

  private collapseGateway(): void {
    this._expandedGateway.set(null);
    this.sensorService.clearSensors();
  }
}
