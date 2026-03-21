import { Component, computed, input, output } from '@angular/core';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatTooltip } from '@angular/material/tooltip';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { TitleCasePipe } from '@angular/common';
import { UpperCasePipe } from '@angular/common';

import { Sensor } from '../../../../models/sensor/sensor.model';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { ActionMode } from '../../../../models/action-mode.model';

@Component({
  selector: 'app-dashboard-sensor-table',
  imports: [
    MatProgressSpinner,
    MatTableModule,
    MatTooltip,
    MatIcon,
    MatPaginatorModule,
    TitleCasePipe,
    UpperCasePipe,
    MatButtonModule,
  ],
  templateUrl: './dashboard-sensor-table.component.html',
  styleUrl: './dashboard-sensor-table.component.css',
})
export class DashboardSensorTableComponent {
  public readonly sensors = input.required<Sensor[]>();
  public readonly loading = input<boolean>();
  public readonly actionMode = input<ActionMode>('dashboard');

  public readonly total = input<number>(0);
  public readonly pageIndex = input<number>(0);
  public readonly limit = input<number>(10);

  protected readonly displayedColumns = computed(() => {
    const base = ['id', 'name', 'profile', 'status'];
    switch (this.actionMode()) {
      case 'manage':
        return [...base, 'delete'];
      default:
        return [...base, 'actions'];
    }
  });

  protected readonly ChartType = ChartType;

  public readonly chartRequested = output<ChartRequest>();
  public readonly deleteRequested = output<Sensor>();
  public readonly createRequested = output<void>();
  public readonly pageChange = output<PageEvent>();

  protected onViewChart(sensor: Sensor, chartType: ChartType): void {
    this.chartRequested.emit({
      sensor,
      chartType,
      timeInterval: null!,
    });
  }

  protected onPageChange(event: PageEvent): void {
    this.pageChange.emit(event);
  }
}
