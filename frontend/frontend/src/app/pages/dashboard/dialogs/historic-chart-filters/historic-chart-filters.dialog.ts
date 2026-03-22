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
import { TimeInterval } from '../../../../models/time-interval.model';
import { ValuesInterval } from '../../../../models/values-interval.model';
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
      dataPointsCounter: [100, [Validators.required, Validators.min(1)]],
      from: [new Date(Date.now() - 24 * 60 * 60 * 1000), Validators.required],
      to: [new Date(), Validators.required],
      lowerBound: [0, Validators.required],
      upperBound: [100, Validators.required],
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

    const timeInterval: TimeInterval = {
      from: this.filtersForm.value.from!,
      to: this.filtersForm.value.to!,
    };

    const valuesInterval: ValuesInterval = {
      lowerBound: this.filtersForm.value.lowerBound!,
      upperBound: this.filtersForm.value.upperBound!,
    };

    const chartRequest: ChartRequest = {
      chartType: this.data.chartType,
      sensor: this.data.sensor,
      timeInterval,
      valuesInterval,
      dataPointsCounter: this.filtersForm.value.dataPointsCounter!,
    };

    this.dialogRef.close(chartRequest);
  }

  protected onCancel(): void {
    this.dialogRef.close();
  }

  private dateRangeValidator(control: AbstractControl): ValidationErrors | null {
    const from = control.get('from')?.value;
    const to = control.get('to')?.value;
    return from && to && from >= to ? { invalidDateRange: true } : null;
  }

  private valueRangeValidator(control: AbstractControl): ValidationErrors | null {
    const lower = control.get('lowerBound')?.value;
    const upper = control.get('upperBound')?.value;
    return lower != null && upper != null && lower >= upper ? { invalidValueRange: true } : null;
  }
}
