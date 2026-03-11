import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';

import { AuthSessionService } from '../../services/auth/auth-session.service';
import { LoginFormComponent } from './components/login-form/login-form.component';
import { ForgotPasswordDialog } from './dialogs/forgot-password/forgot-password.dialog';
import { LoginRequest } from '../../models/login-request.model';
import { ApiError } from '../../models/api-error.model';
import { AuthActionsService } from '../../services/auth/auth-actions.service';

@Component({
  selector: 'app-login.page',
  standalone: true,
  imports: [LoginFormComponent],
  templateUrl: './login.page.html',
  styleUrl: './login.page.css',
})
export class LoginPage {
  private readonly authSessionService = inject(AuthSessionService);
  private readonly authActionsService = inject(AuthActionsService);
  private readonly router = inject(Router);
  private readonly dialog = inject(MatDialog);

  protected loading = signal(false);
  protected generalError = signal('');

  protected onLogin(req: LoginRequest): void {
    this.loading.set(true);
    this.generalError.set('');

    this.authSessionService.login(req).subscribe({
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
      this.authActionsService.forgotPassword(email).subscribe({
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

  protected onDismissError(): void {
    this.generalError.set('');
  }
}