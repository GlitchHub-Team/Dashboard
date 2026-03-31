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
  private readonly authActionsService = inject(AuthActionsService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly destroyRef = inject(DestroyRef);
  // Token e tenantId recuperati dall'URL
  private readonly token = this.route.snapshot.queryParamMap.get('token') ?? '';
  private readonly tenantId = this.route.snapshot.queryParamMap.get('tenant_id') ?? undefined;

  protected readonly loading = this.authActionsService.loading;
  protected readonly generalError = this.authActionsService.error;
  protected readonly passwordChangeResult = this.authActionsService.passwordChangeResult;

  protected onSubmitReset(forgotPasswordResponse: ForgotPasswordResponse): void {
    const requestWithToken: ForgotPasswordResponse = {
      ...forgotPasswordResponse,
      token: this.token,
      tenantId: this.tenantId,
    };
    // La conferma del reset password non logga l'utente, quindi rimane sulla pagina di reset password
    // anche dopo la conferma, mostrando un messaggio di successo e un pulsante per tornare al login
    this.authActionsService
      .confirmPasswordReset(requestWithToken)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe();
  }

  protected onGoToLogin(): void {
    this.authActionsService.clearMessages();
    this.router.navigate(['/login']);
  }

  protected onDismissError(): void {
    this.authActionsService.clearMessages();
  }
}
