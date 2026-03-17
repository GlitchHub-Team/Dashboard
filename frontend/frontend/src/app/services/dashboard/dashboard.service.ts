import { inject, Injectable, signal, computed } from '@angular/core';

import { GatewayService } from '../gateway/gateway.service';
import { SensorService } from '../sensor/sensor.service';
import { PermissionService } from '../permission/permission.service';
import { Permission } from '../../models/permission.enum';
import { Gateway } from '../../models/gateway.model';
import { ChartRequest } from '../../models/chart-request.model';

@Injectable({
  providedIn: 'root',
})
export class DashboardService {
  private readonly gatewayService = inject(GatewayService);
  private readonly sensorService = inject(SensorService);
  // Alert service
  private readonly permissionService = inject(PermissionService);

  private readonly _expandedGateway = signal<Gateway | null>(null);
  private readonly _selectedChart = signal<ChartRequest | null>(null);

  public readonly expandedGateway = this._expandedGateway.asReadonly();
  public readonly selectedChart = this._selectedChart.asReadonly();
  public readonly canSendCommands = computed(() => {
    // Se può mandare comandi, può vedere i gateway
    return this.permissionService.can(Permission.GATEWAY_COMMANDS);
  });

  public readonly gatewayList = this.gatewayService.gatewayList;
  public readonly sensorList = this.sensorService.sensorList;
  // Alert list
  public readonly gatewayLoading = this.gatewayService.loading;
  public readonly gatewayError = this.gatewayService.error;
  public readonly sensorLoading = this.sensorService.loading;
  public readonly sensorError = this.sensorService.error;
  // Alert loading
  // Alert error

  public loadDashboard(): void {
    // TODO: come passare il tenantId?
    // TODO: testing con mock services, uso tenant hard-coded
    if (this.canSendCommands()) {
      this.gatewayService.getGatewaysByTenant('tenant-01');
    } else {
      this.sensorService.getSensorsByTenant('tenant-01');
    }
  }

  public toggleExpandedGateway(gateway: Gateway): void {
    const current = this._expandedGateway();
    if (current?.id === gateway.id) {
      this._expandedGateway.set(null);
      this.sensorService.clearSensors();
    } else {
      this._expandedGateway.set(gateway);
      this.sensorService.getSensorsByGateway(gateway.id);
    }
  }

  public openChart(request: ChartRequest): void {
    this._selectedChart.set(request);
  }

  public closeChart(): void {
    this._selectedChart.set(null);
  }
}
