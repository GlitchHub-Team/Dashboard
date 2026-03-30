import { Component, inject, input, output } from '@angular/core';
import { FormBuilder, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { AbstractControl, ValidationErrors } from '@angular/forms';

import { ConfirmAccountResponse } from '../../../../models/auth/confirm-account.model';

@Component({
  selector: 'app-confirm-account-form',
  imports: [
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './confirm-account-form.component.html',
  styleUrl: './confirm-account-form.component.css',
})
export class ConfirmAccountFormComponent {
  private readonly formBuilder = inject(FormBuilder);

  public loading = input(false);
  public generalError = input<string | null>(null);

  public submitConfirmAccount = output<ConfirmAccountResponse>();
  public dismissError = output<void>();

  protected confirmAccountForm = this.formBuilder.nonNullable.group(
    {
      newPassword: ['', Validators.required],
      confirmNewPassword: ['', Validators.required],
    },
    { validators: this.passwordsMatchValidator },
  );

  protected onSubmit(): void {
    if (!this.confirmAccountForm.valid) {
      this.confirmAccountForm.markAllAsTouched();
      return;
    }

    // Il token viene recuperato dalla page
    this.submitConfirmAccount.emit({
      token: '',
      newPassword: this.confirmAccountForm.value.newPassword!,
    });
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
