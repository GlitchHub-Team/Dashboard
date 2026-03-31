import { Component, inject, input, output } from '@angular/core';
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
import { ForgotPasswordResponse } from '../../../../models/auth/forgot-password.model';

@Component({
  selector: 'app-reset-password-form',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './reset-password-form.component.html',
  styleUrl: './reset-password-form.component.css',
})
export class ResetPasswordFormComponent {
  private readonly formBuilder = inject(FormBuilder);

  public loading = input(false);
  public generalError = input<string | null>(null);
  public success = input(false);

  public submitReset = output<ForgotPasswordResponse>();
  public goToLogin = output<void>();
  public dismissError = output<void>();

  protected resetPasswordForm = this.formBuilder.nonNullable.group(
    {
      newPassword: ['', [Validators.required]],
      confirmNewPassword: ['', [Validators.required]],
    },
    { validators: this.passwordsMatchValidator },
  );

  protected onSubmit(): void {
    if (!this.resetPasswordForm.valid) {
      this.resetPasswordForm.markAllAsTouched();
      return;
    }

    // Come per confirm-account il token viene recuperato dalla page
    const forgotPasswordResponse: ForgotPasswordResponse = {
      token: '', // Il token viene aggiunto nella page, quindi qui può essere lasciato vuoto
      tenantId: undefined, // Il tenant_id viene gestito a livello di page, quindi può essere lasciato undefined qui
      newPassword: this.resetPasswordForm.controls.newPassword.value!,
    };

    this.submitReset.emit(forgotPasswordResponse);
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
