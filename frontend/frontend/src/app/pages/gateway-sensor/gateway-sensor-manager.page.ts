import { Component, computed, DestroyRef, inject, OnInit } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatDialog } from '@angular/material/dialog';
import { MatIcon } from '@angular/material/icon';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';
import { filter, switchMap } from 'rxjs';

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
  private readonly destroyRef = inject(DestroyRef);

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
    this.dialog
      .open(CreateGatewayDialog)
      .afterClosed()
      .pipe(
        filter((created) => !!created),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.managerService.refreshGateways();
        this.snackBar.open('Gateway creato con successo', 'Close', { duration: 3000 });
      });
  }

  protected onDeleteGateway(gateway: Gateway): void {
    this.dialog
      .open(ConfirmDeleteDialog, {
        data: {
          title: 'Elimina Gateway',
          message: `Sei sicuro di voler eliminare il gateway "${gateway.name}"?`,
        },
      })
      .afterClosed()
      .pipe(
        filter((confirmed) => !!confirmed),
        switchMap(() => this.managerService.deleteGateway(gateway)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.snackBar.open('Gateway eliminato con successo', 'Close', { duration: 3000 });
      });
  }

  protected onCreateSensor(gateway: Gateway): void {
    this.dialog
      .open(CreateSensorDialog, {
        data: {
          id: gateway.id,
          name: gateway.name,
        },
      })
      .afterClosed()
      .pipe(
        filter((created) => !!created),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.managerService.refreshSensors(gateway.id);
        this.snackBar.open('Sensore creato con successo', 'Close', { duration: 3000 });
      });
  }

  protected onDeleteSensor(sensor: Sensor): void {
    this.dialog
      .open(ConfirmDeleteDialog, {
        data: {
          title: 'Elimina Sensore',
          message: `Sei sicuro di voler eliminare il sensore "${sensor.name}"?`,
        },
      })
      .afterClosed()
      .pipe(
        filter((confirmed) => !!confirmed),
        switchMap(() => this.managerService.deleteSensor(sensor)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => {
        this.snackBar.open('Sensore eliminato con successo', 'Close', { duration: 3000 });
      });
  }
}
