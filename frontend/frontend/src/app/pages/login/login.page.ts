import { Component, inject, DestroyRef } from '@angular/core';
import { Router } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatDialog } from '@angular/material/dialog';

import { AuthSessionService } from '../../services/auth/auth-session.service';
import { LoginFormComponent } from './components/login-form/login-form.component';
import { ForgotPasswordDialog } from './dialogs/forgot-password/forgot-password.dialog';
import { LoginRequest } from '../../models/login-request.model';

@Component({
  selector: 'app-login.page',
  standalone: true,
  imports: [LoginFormComponent],
  templateUrl: './login.page.html',
  styleUrl: './login.page.css',
})
export class LoginPage {
  protected readonly authSessionService = inject(AuthSessionService);
  private readonly router = inject(Router);
  private readonly dialog = inject(MatDialog);
  private readonly destroyRef = inject(DestroyRef);

  protected onLogin(req: LoginRequest): void {
    this.authSessionService
      .login(req)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.router.navigate(['/dashboard']),
      });
  }

  protected onForgotPassword(): void {
    this.dialog.open(ForgotPasswordDialog, {
      width: '800px',
      disableClose: true,
    });
  }

  protected onDismissError(): void {
    this.authSessionService.clearError();
  }
}
