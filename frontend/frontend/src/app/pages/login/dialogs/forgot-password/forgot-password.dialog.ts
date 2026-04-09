import { Component, inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSelectModule } from '@angular/material/select';

import { AuthActionsService } from '../../../../services/auth/auth-actions.service';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { ForgotPasswordRequest } from '../../../../models/auth/forgot-password-request.model';

@Component({
  selector: 'app-forgot-password.dialog',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
    MatSelectModule,
  ],
  templateUrl: './forgot-password.dialog.html',
  styleUrl: './forgot-password.dialog.css',
})
export class ForgotPasswordDialog {
  private readonly formBuilder = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<ForgotPasswordDialog>);
  private readonly authActionsService = inject(AuthActionsService);
  private readonly tenantService = inject(TenantService);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly displayedTenants = this.tenantService.tenantList;

  protected readonly forgotPasswordForm = this.formBuilder.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    tenantId: [''],
  });

  protected readonly loading = this.authActionsService.loading;
  protected readonly generalError = this.authActionsService.error;

  constructor() {
    this.tenantService.retrieveTenants(false);
    this.setupAutoClear();
  }

  protected onSubmit(): void {
    if (!this.forgotPasswordForm.valid) {
      this.forgotPasswordForm.markAllAsTouched();
      return;
    }

    // Confeziona e invia la richiesta di reset password, chiude il dialog alla risposta positiva
    const tenantId = this.forgotPasswordForm.controls.tenantId.value;
    const forgotPasswordRequest: ForgotPasswordRequest = {
      email: this.forgotPasswordForm.controls.email.value!,
      ...(tenantId ? { tenantId } : {}),
    };

    this.authActionsService
      .forgotPassword(forgotPasswordRequest)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.dialogRef.close(true),
      });
  }

  protected onCancel(): void {
    this.authActionsService.clearMessages();
    this.dialogRef.close(false);
  }

  protected dismissError(): void {
    this.authActionsService.clearMessages();
  }

  protected getFieldError(field: string, label: string): string {
    const control = this.forgotPasswordForm.get(field);
    if (!control?.errors) return '';

    if (control.hasError('required')) {
      return `${label} is required.`;
    }

    if (control.hasError('email')) {
      return 'Please enter a valid email address.';
    }

    return 'Invalid value';
  }

  // Pulisce errori quando l'utente digita nei campi
  private setupAutoClear(): void {
    for (const key of Object.keys(this.forgotPasswordForm.controls)) {
      this.forgotPasswordForm
        .get(key)!
        .valueChanges.pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe(() => {
          if (this.generalError()) {
            this.authActionsService.clearMessages();
          }
        });
    }
  }
}
