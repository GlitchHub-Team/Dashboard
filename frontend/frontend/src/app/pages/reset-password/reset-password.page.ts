import { Component, computed, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import {
  FormBuilder,
  Validators,
  ReactiveFormsModule,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';

import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { PasswordChange } from '../../models/password-change.model';
import { ApiError } from '../../models/api-error.model';

@Component({
  selector: 'app-reset-password.page',
  imports: [
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './reset-password.page.html',
  styleUrl: './reset-password.page.css',
})
export class ResetPasswordPage {
  private readonly authActionsService = inject(AuthActionsService);
  private readonly router = inject(Router);
  private readonly formBuilder = inject(FormBuilder);

  protected resetPasswordForm = this.formBuilder.nonNullable.group(
    {
      newPassword: ['', [Validators.required]],
      confirmNewPassword: ['', [Validators.required]],
    },
    { validators: this.passwordsMatchValidator },
  );

  // TODO: refactor come fatto per forgot password dialog,
  // con service che espone loading/error e gestione errori di campo
  protected loading = signal(false);
  protected generalError = signal('');
  protected success = signal(false);

  protected passwordMismatch = computed(
    () =>
      this.resetPasswordForm.hasError('passwordMismatch') &&
      this.resetPasswordForm.get('confirmNewPassword')!.touched,
  );

  protected onSubmit(): void {
    if (!this.resetPasswordForm.valid) {
      this.resetPasswordForm.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.generalError.set('');

    // TODO: Da dove prendiamo il token?
    this.authActionsService
      .confirmPasswordChange({
        newPassword: this.resetPasswordForm.get('newPassword')?.value,
        token: 'TODO',
      } as PasswordChange)
      .subscribe({
        next: () => {
          this.loading.set(false);
          this.success.set(true);
          this.goToLogin();
        },
        error: (err: ApiError) => {
          this.loading.set(false);
          this.generalError.set(err.message);
        },
      });
  }

  protected goToLogin(): void {
    this.router.navigate(['/login']);
  }

  protected dismissError(): void {
    this.generalError.set('');
  }

  private passwordsMatchValidator(control: AbstractControl): ValidationErrors | null {
    const password = control.get('newPassword')?.value;
    const confirmPassword = control.get('confirmNewPassword')?.value;

    if (password && confirmPassword && password !== confirmPassword) {
      return { passwordMismatch: true };
    }
    return null;
  }
}
