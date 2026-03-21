import { inject, Injectable, signal } from '@angular/core';

import { GatewayService } from '../gateway/gateway.service';
import { SensorService } from '../sensor/sensor.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { GatewayConfig } from '../../models/gateway/gateway-config.model';
import { SensorConfig } from '../../models/sensor/sensor-config.model';

@Injectable({ providedIn: 'root' })
export class GatewaySensorManagerService {
  private readonly gatewayService = inject(GatewayService);
  private readonly sensorService = inject(SensorService);

  private readonly _expandedGateway = signal<Gateway | null>(null);

  public readonly expandedGateway = this._expandedGateway.asReadonly();

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

  public loadGateways(): void {
    this.gatewayService.getGateways(this.gatewayPageIndex(), this.gatewayLimit());
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

  public changeGatewayPage(pageIndex: number, limit: number): void {
    this.collapseGateway();
    this.gatewayService.changePage(pageIndex, limit);
  }

  public changeSensorPage(pageIndex: number, limit: number): void {
    this.sensorService.changePage(pageIndex, limit);
  }

  public createGateway(data: GatewayConfig): void {
    this.gatewayService.addNewGateway(data).subscribe(() => {
      this.refreshGateways();
    });
  }

  public deleteGateway(gateway: Gateway): void {
    const wasExpanded = this._expandedGateway()?.id === gateway.id;
    this.gatewayService.deleteGateway(gateway.id).subscribe(() => {
      if (wasExpanded) this.collapseGateway();
      this.refreshGateways();
    });
  }

  public createSensor(gatewayId: string, data: SensorConfig): void {
    const sensorData = { ...data, gatewayId };
    this.sensorService.addNewSensor(sensorData).subscribe(() => {
      this.refreshSensors(gatewayId);
    });
  }

  public deleteSensor(sensor: Sensor): void {
    this.sensorService.deleteSensor(sensor.id).subscribe(() => {
      this.refreshSensors(sensor.gatewayId);
    });
  }

  private refreshGateways(): void {
    this.gatewayService.changePage(this.gatewayPageIndex(), this.gatewayLimit());
  }

  private refreshSensors(gatewayId: string): void {
    this.sensorService.getSensorsByGateway(gatewayId, this.sensorPageIndex(), this.sensorLimit());
  }

  private collapseGateway(): void {
    this._expandedGateway.set(null);
    this.sensorService.clearSensors();
  }
}
