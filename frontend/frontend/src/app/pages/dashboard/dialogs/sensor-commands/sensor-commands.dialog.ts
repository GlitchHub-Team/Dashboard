import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIcon } from '@angular/material/icon';
import { MatProgressSpinner } from '@angular/material/progress-spinner';

import { SensorService } from '../../../../services/sensor/sensor.service';
import { ApiError } from '../../../../models/api-error.model';
import { ActionMode } from '../../../../models/action-mode.model';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorStatus } from '../../../../models/sensor-status.enum';
import { MatInputModule } from '@angular/material/input';

@Component({
  selector: 'app-sensor-commands',
  imports: [
    MatDialogModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIcon,
    MatProgressSpinner,
  ],
  templateUrl: './sensor-commands.dialog.html',
  styleUrl: './sensor-commands.dialog.css',
})
export class SensorCommandsDialog {
  private readonly dialogRef = inject(MatDialogRef<SensorCommandsDialog>);
  private readonly formBuilder = inject(FormBuilder);
  private readonly sensorService = inject(SensorService);
  protected readonly data = inject<{ sensor: Sensor; mode: ActionMode }>(MAT_DIALOG_DATA);

  protected generalError = signal('');
  protected isSubmitting = signal(false);

  protected readonly commands: { value: string; label: string }[] =
    this.data.sensor.status === SensorStatus.ACTIVE
      ? [{ value: 'interrupt', label: 'Interrompi' }]
      : [{ value: 'resume', label: 'Riprendi' }];

  protected readonly commandForm = this.formBuilder.nonNullable.group({
    command: ['', Validators.required],
  });

  protected onConfirm(): void {
    if (!this.commandForm.valid) {
      this.commandForm.markAllAsTouched();
      return;
    }

    const command = this.commandForm.controls.command.value;

    this.isSubmitting.set(true);
    this.generalError.set('');

    switch (command) {
      case 'interrupt':
        this.sensorService.interruptSensor(this.data.sensor.id).subscribe({
          next: () => {
            this.dialogRef.close(true);
          },
          error: (err: ApiError) => {
            this.generalError.set(err.message ?? 'Invio comando fallito');
            this.isSubmitting.set(false);
          },
        });
        break;
      case 'resume':
        this.sensorService.resumeSensor(this.data.sensor.id).subscribe({
          next: () => {
            this.dialogRef.close(true);
          },
          error: (err: ApiError) => {
            this.generalError.set(err.message ?? 'Invio comando fallito');
            this.isSubmitting.set(false);
          },
        });
        break;
      default:
        this.generalError.set('Comando sconosciuto');
        this.isSubmitting.set(false);
        this.dialogRef.close(false);
    }
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }

  protected dismissError(): void {
    this.generalError.set('');
  }
}
