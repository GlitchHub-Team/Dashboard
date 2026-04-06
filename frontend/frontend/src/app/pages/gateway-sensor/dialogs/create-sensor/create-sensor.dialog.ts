import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';
import { TitleCasePipe } from '@angular/common';

import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { SensorService } from '../../../../services/sensor/sensor.service';
import { SensorConfig } from '../../../../models/sensor/sensor-config.model';
import { ApiError } from '../../../../models/api-error.model';
import { sensorProfilesMapper } from '../../../../utils/sensor-profile.utils';

@Component({
  selector: 'app-create-sensor',
  imports: [
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatProgressSpinner,
    MatIcon,
    TitleCasePipe,
    MatDialogModule,
    ReactiveFormsModule,
  ],
  templateUrl: './create-sensor.dialog.html',
  styleUrl: './create-sensor.dialog.css',
})
export class CreateSensorDialog {
  private readonly dialogRef = inject(MatDialogRef<CreateSensorDialog>);
  private readonly sensorService = inject(SensorService);
  private readonly formBuilder = inject(FormBuilder);
  protected readonly data = inject<{ id: string; name: string }>(MAT_DIALOG_DATA);

  protected readonly profiles = Object.entries(SensorProfiles).map(([key, label]) => ({
    key,
    label,
  }));

  protected sensorForm = this.formBuilder.nonNullable.group({
    name: ['', Validators.required],
    profile: [null as SensorProfiles | null, Validators.required],
    interval: [1000, [Validators.required, Validators.min(100)]],
  });

  protected readonly generalError = signal('');
  protected readonly isSubmitting = signal(false);

  protected onSubmit(): void {
    if (!this.sensorForm.valid) {
      this.sensorForm.markAllAsTouched();
      return;
    }

    this.isSubmitting.set(true);
    this.generalError.set('');

    const sensorConfig: SensorConfig = {
      gatewayId: this.data.id,
      name: this.sensorForm.value.name!,
      profile: sensorProfilesMapper.toBackend(this.sensorForm.value.profile! as SensorProfiles),
      dataInterval: this.sensorForm.value.interval!,
    };

    this.sensorService.addNewSensor(sensorConfig).subscribe({
      next: () => {
        this.isSubmitting.set(false);
        this.dialogRef.close(true);
      },
      error: (err: ApiError) => {
        this.isSubmitting.set(false);
        this.generalError.set(err.message ?? 'Failed to create sensor. Please try again.');
      },
    });
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }
}
