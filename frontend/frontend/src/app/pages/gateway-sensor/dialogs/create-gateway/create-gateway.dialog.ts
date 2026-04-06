import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinner } from '@angular/material/progress-spinner';
import { MatIcon } from '@angular/material/icon';

import { GatewayService } from '../../../../services/gateway/gateway.service';
import { ApiError } from '../../../../models/api-error.model';
import { GatewayConfig } from '../../../../models/gateway/gateway-config.model';

@Component({
  selector: 'app-create-gateway',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinner,
    MatIcon,
  ],
  templateUrl: './create-gateway.dialog.html',
  styleUrl: './create-gateway.dialog.css',
})
export class CreateGatewayDialog {
  private readonly gatewayService = inject(GatewayService);
  private readonly dialogRef = inject(MatDialogRef<CreateGatewayDialog>);

  protected readonly gatewayForm = inject(FormBuilder).group({
    name: ['', Validators.required],
    interval: [1000, [Validators.required, Validators.min(100)]],
  });

  protected readonly isSubmitting = signal(false);
  protected readonly generalError = signal('');

  protected onSubmit(): void {
    if (!this.gatewayForm.valid) {
      this.gatewayForm.markAllAsTouched();
      return;
    }

    this.isSubmitting.set(true);
    this.generalError.set('');

    const gatewayConfig: GatewayConfig = {
      name: this.gatewayForm.value.name!,
      interval: this.gatewayForm.value.interval!,
    };

    this.gatewayService.addNewGateway(gatewayConfig).subscribe({
      next: () => this.dialogRef.close(true),
      error: (err: ApiError) => {
        this.generalError.set(err.message || 'Failed to create gateway. Please try again.');
        this.isSubmitting.set(false);
      },
    });
  }

  protected onCancel(): void {
    this.dialogRef.close(false);
  }

  protected dismissError(): void {
    this.generalError.set('');
  }
}
