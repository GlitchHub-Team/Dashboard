import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { FormBuilder, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatDialog } from '@angular/material/dialog';

import { AuthService } from '../../services/auth/auth.service';
import { ApiError } from '../../models/api-error.model';
import { ForgotPasswordDialog } from './dialogs/forgot-password/forgot-password.dialog';

@Component({
  selector: 'app-login.page',
  imports: [
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressBarModule,
  ],
  templateUrl: './login.page.html',
  styleUrl: './login.page.css',
})
export class LoginPage {
  private formBuilder = inject(FormBuilder);
  private authService = inject(AuthService);
  private router = inject(Router);
  private dialog = inject(MatDialog);

  protected loginForm = this.formBuilder.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required]],
  });

  protected loading = signal(false);
  protected generalError = signal('');

  protected onSubmit(): void {
    if (!this.loginForm.valid) {
      this.loginForm.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.generalError.set('');

    this.authService.login(this.loginForm.getRawValue()).subscribe({
      next: () => {
        this.loading.set(false);
        this.router.navigate(['/dashboard']);
      },
      error: (err: ApiError) => {
        this.loading.set(false);
        this.generalError.set(err.message);
      },
    });
  }

  protected onForgotPassword(): void {
    const dialogRef = this.dialog.open(ForgotPasswordDialog, {
      width: '800px',
      disableClose: true,
    });

    const instance = dialogRef.componentInstance as ForgotPasswordDialog;

    instance.save$.subscribe((email: string) => {
      this.authService.requestPasswordReset(email).subscribe({
        next: () => {
          instance.setLoading(false);
          dialogRef.close(true);
        },
        error: (err: ApiError) => {
          instance.setLoading(false);
          instance.setServerError(err);
        },
      });
    });
  }

  protected dismissError(): void {
    this.generalError.set('');
  }
}
