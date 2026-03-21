// pages/manager/manager.page.ts
import { Component, computed, inject, OnInit } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatIcon } from '@angular/material/icon';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';

import { GatewaySensorManagerService } from '../../services/gateway-sensor-manager/gateway-sensor-manager.service';
import { DashboardGatewayTableComponent } from '../dashboard/components/dashboard-gateway-table/dashboard-gateway-table.component';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { ConfirmDeleteDialog } from './dialogs/confirm-delete/confirm-delete.dialog';
import { CreateGatewayDialog } from './dialogs/create-gateway/create-gateway.dialog';
import { CreateSensorDialog } from './dialogs/create-sensor/create-sensor.dialog';

@Component({
  selector: 'app-gateway-sensor-manager',
  imports: [DashboardGatewayTableComponent, MatIcon],
  templateUrl: './gateway-sensor-manager.page.html',
  styleUrl: './gateway-sensor-manager.page.css',
})
export class GatewaySensorManagerPage implements OnInit {
  private readonly managerService = inject(GatewaySensorManagerService);
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);

  protected readonly gatewayList = this.managerService.gatewayList;
  protected readonly gatewayTotal = this.managerService.gatewayTotal;
  protected readonly gatewayPageIndex = this.managerService.gatewayPageIndex;
  protected readonly gatewayLimit = this.managerService.gatewayLimit;
  protected readonly gatewayLoading = this.managerService.gatewayLoading;

  protected readonly sensorList = this.managerService.sensorList;
  protected readonly sensorTotal = this.managerService.sensorTotal;
  protected readonly sensorPageIndex = this.managerService.sensorPageIndex;
  protected readonly sensorLimit = this.managerService.sensorLimit;
  protected readonly sensorLoading = this.managerService.sensorLoading;

  protected readonly expandedGateway = this.managerService.expandedGateway;
  protected readonly error = computed(
    () => this.managerService.gatewayError() ?? this.managerService.sensorError(),
  );

  public ngOnInit(): void {
    this.managerService.loadGateways();
  }

  protected onExpandedGatewayChange(gateway: Gateway): void {
    this.managerService.toggleExpandedGateway(gateway);
  }

  protected onGatewayPageChange(event: PageEvent): void {
    this.managerService.changeGatewayPage(event.pageIndex, event.pageSize);
  }

  protected onSensorPageChange(event: PageEvent): void {
    this.managerService.changeSensorPage(event.pageIndex, event.pageSize);
  }

  protected onCreateGateway(): void {
    const ref = this.dialog.open(CreateGatewayDialog);

    ref.afterClosed().subscribe((result) => {
      if (result) {
        this.managerService.createGateway(result);
        this.snackBar.open('Gateway created', 'Close', { duration: 3000 });
      }
    });
  }

  protected onDeleteGateway(gateway: Gateway): void {
    const ref = this.dialog.open(ConfirmDeleteDialog, {
      data: { entityName: gateway.name, entityType: 'gateway' },
    });

    ref.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.managerService.deleteGateway(gateway);
        this.snackBar.open('Gateway deleted', 'Close', { duration: 3000 });
      }
    });
  }

  protected onCreateSensor(gateway: Gateway): void {
    const ref = this.dialog.open(CreateSensorDialog, {
      data: { gatewayId: gateway.id },
    });

    ref.afterClosed().subscribe((result) => {
      if (result) {
        this.managerService.createSensor(gateway.id, result);
        this.snackBar.open('Sensor created', 'Close', { duration: 3000 });
      }
    });
  }

  protected onDeleteSensor(sensor: Sensor): void {
    const ref = this.dialog.open(ConfirmDeleteDialog, {
      data: { entityName: sensor.name, entityType: 'sensor' },
    });

    ref.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.managerService.deleteSensor(sensor);
        this.snackBar.open('Sensor deleted', 'Close', { duration: 3000 });
      }
    });
  }
}
