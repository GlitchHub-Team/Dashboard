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

  public readonly _gatewayList = this.gatewayService.gatewayList;
  public readonly _sensorList = this.sensorService.sensorList;
  // Alert list
  public readonly _gatewayLoading = this.gatewayService.loading;
  public readonly _gatewayError = this.gatewayService.error;
  public readonly _sensorLoading = this.sensorService.loading;
  public readonly _sensorError = this.sensorService.error;
  // Alert loading
  // Alert error

  public loadDashboard(): void {
    // TODO: come passare il tenantId?
    if (this.canSendCommands()) {
      this.gatewayService.getGatewaysByTenant('current');
    } else {
      this.sensorService.getSensorsByGateway('current');
    }
  }

  public toggleExpandedGateway(gateway: Gateway): void {
    this._expandedGateway.update((current) => (current?.id === gateway.id ? null : gateway));
    this.sensorService.getSensorsByGateway(gateway.id);
  }

  public openChart(request: ChartRequest): void {
    this._selectedChart.set(request);
  }

  public closeChart(): void {
    this._selectedChart.set(null);
  }
}
