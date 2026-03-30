import { Component, inject, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { MatDialogRef, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { AuthActionsService } from '../../../../services/auth/auth-actions.service';
import { PasswordChange } from '../../../../models/auth/password-change.model';

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
  private readonly dialogRef = inject(MatDialogRef<ChangePasswordDialog>);
  private readonly authActionsService = inject(AuthActionsService);
  private readonly destroyRef = inject(DestroyRef);

  // Cambiare la password da loggato richiede solo la vecchia password e la nuova password
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

  constructor() {
    this.setupAutoClear();
  }

  protected onSubmit(): void {
    if (!this.form.valid) {
      this.form.markAllAsTouched();
      return;
    }

    const data: PasswordChange = {
      oldPassword: this.form.controls.oldPassword.value,
      newPassword: this.form.controls.newPassword.value,
    };

    // Il token passato è il JWT dell'utente loggato, che viene recuperato automaticamente dall'AuthInterceptor
    this.authActionsService
      .confirmPasswordChange(data)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.authActionsService.clearMessages();
          this.dialogRef.close(true);
        },
      });
  }

  protected onCancel(): void {
    this.authActionsService.clearMessages();
    this.dialogRef.close(false);
  }

  protected dismissError(): void {
    this.authActionsService.clearMessages();
  }

  protected getFieldError(field: string): string {
    const control = this.form.get(field);
    if (!control?.errors) return '';

    if (control.hasError('required')) {
      return `Campo obbligatorio.`;
    }

    return 'Invalid value';
  }

  private passwordsMatchValidator(control: AbstractControl): ValidationErrors | null {
    const newPassword = control.get('newPassword')?.value;
    const confirmNewPassword = control.get('confirmNewPassword')?.value;

    if (newPassword && confirmNewPassword && newPassword !== confirmNewPassword) {
      return { passwordMismatch: true };
    }
    return null;
  }

  private setupAutoClear(): void {
    for (const key of Object.keys(this.form.controls)) {
      this.form
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
