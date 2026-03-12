import { Component, inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import {
  FormBuilder,
  Validators,
  ReactiveFormsModule,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { AuthActionsService } from '../../../../services/auth/auth-actions.service';
import { ApiError } from '../../../../models/api-error.model';
import { PasswordChange } from '../../../../models/password-change.model';
import { TokenStorageService } from '../../../../services/token-storage/token-storage.service';

@Component({
  selector: 'app-change-password-dialog',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './change-password.dialog.html',
  styleUrl: './change-password.dialog.css',
})
export class ChangePasswordDialog {
  private readonly formBuilder = inject(FormBuilder);
  private readonly tokenStorageService = inject(TokenStorageService);
  private readonly dialogRef = inject(MatDialogRef<ChangePasswordDialog>);
  private readonly authActionsService = inject(AuthActionsService);
  private readonly destroyRef = inject(DestroyRef);

  protected form = this.formBuilder.nonNullable.group(
    {
      oldPassword: ['', [Validators.required]],
      newPassword: ['', [Validators.required]],
      confirmNewPassword: ['', [Validators.required]],
    },
    { validators: this.passwordsMatchValidator },
  );

  protected readonly loading = this.authActionsService.loading;
  protected readonly generalError = this.authActionsService.error;

  private serverErrors = new Map<string, string>();

  constructor() {
    this.setupAutoClear();
  }

  protected onSubmit(): void {
    if (!this.form.valid) {
      this.form.markAllAsTouched();
      return;
    }

    this.serverErrors.clear();

    const data: PasswordChange = {
      token: this.tokenStorageService.getToken() ?? '',
      newPassword: this.form.controls.newPassword.value,
    };

    this.authActionsService
      .confirmPasswordChange(data)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.dialogRef.close(true),
        error: (err: ApiError) => this.handleServerErrors(err),
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
    const control = this.form.get(field);
    if (!control?.errors) return '';

    if (control.hasError('serverError')) {
      return this.serverErrors.get(field) ?? '';
    }

    if (control.hasError('required')) {
      return `${label} is required.`;
    }

    return 'Invalid value';
  }

  private handleServerErrors(error: ApiError): void {
    if (!error.errors?.length) return;

    for (const fieldError of error.errors) {
      const control = this.form.get(fieldError.field);
      if (control) {
        control.setErrors({ serverError: true });
        control.markAsTouched();
        this.serverErrors.set(fieldError.field, fieldError.message);
      }
    }
  }

  private setupAutoClear(): void {
    for (const key of Object.keys(this.form.controls)) {
      this.form
        .get(key)!
        .valueChanges.pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe(() => {
          if (this.serverErrors.has(key)) {
            this.serverErrors.delete(key);
          }
          if (this.generalError()) {
            this.authActionsService.clearMessages();
          }
        });
    }
  }

  private passwordsMatchValidator(control: AbstractControl): ValidationErrors | null {
    const password = control.get('newPassword')?.value;
    const confirm = control.get('confirmNewPassword')?.value;

    if (password && confirm && password !== confirm) {
      return { passwordMismatch: true };
    }
    return null;
  }
}
