import { Component, inject, DestroyRef } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { ResetPasswordFormComponent } from './components/reset-password-form/reset-password-form.component';
import { PasswordChange } from '../../models/auth/password-change.model';

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

  // TODO: Da dove lo prendiamo il token? Dall'URL?
  //private readonly token = this.route.snapshot.queryParamMap.get('token') ?? '';

  protected onSubmitReset(newPassword: string): void {
    const data: PasswordChange = {
      newPassword,
      token: 'TODO',
    };

    this.authActionsService
      .confirmPasswordChange(data)
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
