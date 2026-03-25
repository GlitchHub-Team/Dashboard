import { Component, inject, DestroyRef } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { ResetPasswordFormComponent } from './components/reset-password-form/reset-password-form.component';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';

@Component({
  selector: 'app-reset-password-page',
  standalone: true,
  imports: [ResetPasswordFormComponent],
  templateUrl: './reset-password.page.html',
  styleUrl: './reset-password.page.css',
})
export class ResetPasswordPage {
  protected readonly authActionsService = inject(AuthActionsService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly destroyRef = inject(DestroyRef);
  private readonly token = this.route.snapshot.queryParamMap.get('token') ?? '';

  protected onSubmitReset(forgotPasswordResponse: ForgotPasswordResponse): void {
    const requestWithToken: ForgotPasswordResponse = {
      ...forgotPasswordResponse,
      token: this.token,
    };
    this.authActionsService
      .confirmPasswordReset(requestWithToken)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe();
  }

  protected onGoToLogin(): void {
    this.router.navigate(['/login']);
  }

  protected onDismissError(): void {
    this.authActionsService.clearMessages();
  }
}
