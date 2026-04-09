import { Component, computed, inject, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';
import { MatDialog } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatTooltip } from '@angular/material/tooltip';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { UpperCasePipe } from '@angular/common';

import { GatewayExpandedComponent } from '../gateway-expanded/gateway-expanded.component';
import { GatewayCommandsDialog } from '../../../dashboard/dialogs/gateway-commands/gateway-commands.dialog';
import { Gateway } from '../../../../models/gateway/gateway.model';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ActionMode } from '../../../../models/action-mode.model';
import { MatSnackBar } from '@angular/material/snack-bar';

@Component({
  selector: 'app-gateway-table',
  imports: [
    MatProgressSpinner,
    MatIcon,
    MatTableModule,
    MatTooltip,
    MatPaginatorModule,
    UpperCasePipe,
    GatewayExpandedComponent,
    MatButtonModule,
  ],
  templateUrl: './gateway-table.component.html',
  styleUrl: './gateway-table.component.css',
})
export class GatewayTableComponent {
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);

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

  public readonly commandRequested = output<boolean>();
  public readonly chartRequested = output<ChartRequest>();
  public readonly expandedGatewayChange = output<Gateway>();

  public readonly gatewayDeleteRequested = output<Gateway>();
  public readonly gatewayCreateRequested = output<void>();
  public readonly sensorDeleteRequested = output<Sensor>();
  public readonly sensorCreateRequested = output<Gateway>();

  public readonly gatewayPageChange = output<PageEvent>();
  public readonly sensorPageChange = output<PageEvent>();

  private readonly columns = ['id', 'tenantId', 'name', 'status', 'commands'];
  protected readonly displayedColumns = computed(() => {
    if (this.actionMode() === 'manage') {
      return ['id', 'tenantId', 'name', 'status', 'publicKey', 'commands', 'delete'];
    }
    return this.columns;
  });

  protected isExpanded(gateway: Gateway): boolean {
    return this.expandedGateway()?.id === gateway.id;
  }

  protected onGatewayPageChange(event: PageEvent): void {
    this.gatewayPageChange.emit(event);
  }

  protected copyToClipboard(value: string): void {
    navigator.clipboard.writeText(value);
    this.snackBar.open('Public key copiata negli appunti', 'Chiudi', { duration: 2000 });
  }

  protected onSendCommand(gateway: Gateway): void {
    const ref = this.dialog.open(GatewayCommandsDialog, {
      data: { gateway: gateway, mode: this.actionMode() },
    });

    ref.afterClosed().subscribe((result: boolean | undefined) => {
      if (result) {
        this.commandRequested.emit(true);
      }
    });
  }
}
