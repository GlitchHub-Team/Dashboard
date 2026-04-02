import { Component, inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatIcon } from '@angular/material/icon';

import { Sensor } from '../../../../models/sensor/sensor.model';
import { TimeInterval } from '../../../../models/chart/time-interval.model';
import { ValuesInterval } from '../../../../models/chart/values-interval.model';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';

@Component({
  selector: 'app-historic-chart-filters',
  imports: [
    MatDialogModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatIcon,
  ],
  templateUrl: './historic-chart-filters.dialog.html',
  styleUrl: './historic-chart-filters.dialog.css',
})
export class HistoricChartFiltersDialog {
  private readonly dialogRef = inject(MatDialogRef<HistoricChartFiltersDialog>);
  private readonly formBuilder = inject(FormBuilder);
  protected readonly data = inject<{ sensor: Sensor; chartType: ChartType }>(MAT_DIALOG_DATA);

  protected readonly filtersForm = this.formBuilder.nonNullable.group(
    {
      dataPointsCounter: [null as number | null, [Validators.required, Validators.min(1)]],
      from: [null as Date | null],
      fromTime: [null as string | null],
      to: [null as Date | null],
      toTime: [null as string | null],
      lowerBound: [null as number | null],
      upperBound: [null as number | null],
    },
    {
      validators: [this.dateRangeValidator, this.valueRangeValidator],
    },
  );

  protected onApply(): void {
    if (!this.filtersForm.valid) {
      this.filtersForm.markAllAsTouched();
      return;
    }

    const { from, fromTime, to, toTime, lowerBound, upperBound, dataPointsCounter } =
      this.filtersForm.value;

    const timeInterval: TimeInterval | undefined =
      from && to
        ? {
            from: this.combineDateAndTime(from, fromTime ?? '00:00'),
            to: this.combineDateAndTime(to, toTime ?? '23:59'),
          }
        : undefined;

    const valuesInterval: ValuesInterval | undefined =
      lowerBound != null && upperBound != null ? { lowerBound, upperBound } : undefined;

    const chartRequest: ChartRequest = {
      chartType: this.data.chartType,
      sensor: this.data.sensor,
      ...(timeInterval && { timeInterval }),
      ...(valuesInterval && { valuesInterval }),
      ...(dataPointsCounter != null && { dataPointsCounter }),
    };

    this.dialogRef.close(chartRequest);
  }

  protected onCancel(): void {
    this.dialogRef.close();
  }

  private combineDateAndTime(date: Date, time: string): Date {
    const [hours, minutes] = time.split(':').map(Number);
    const combined = new Date(date);
    combined.setHours(hours, minutes, 0, 0);
    return combined;
  }

  private dateRangeValidator(control: AbstractControl): ValidationErrors | null {
    const from = control.get('from')?.value;
    const fromTime = control.get('fromTime')?.value ?? '00:00';
    const to = control.get('to')?.value;
    const toTime = control.get('toTime')?.value ?? '23:59';

    if (!from || !to) return null;

    const [fh, fm] = fromTime.split(':').map(Number);
    const [th, tm] = toTime.split(':').map(Number);

    const fromDate = new Date(from);
    fromDate.setHours(fh, fm, 0, 0);

    const toDate = new Date(to);
    toDate.setHours(th, tm, 0, 0);

    return fromDate >= toDate ? { invalidDateRange: true } : null;
  }

  private valueRangeValidator(control: AbstractControl): ValidationErrors | null {
    const lower = control.get('lowerBound')?.value;
    const upper = control.get('upperBound')?.value;
    return lower != null && upper != null && lower >= upper ? { invalidValueRange: true } : null;
  }
}
