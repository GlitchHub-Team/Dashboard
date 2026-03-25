import { Component, DestroyRef, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { AuthActionsService } from '../../services/auth/auth-actions.service';
import { Router } from '@angular/router';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';
import { ConfirmAccountFormComponent } from './components/confirm-account-form/confirm-account-form.component';

@Component({
  selector: 'app-confirm-account',
  imports: [ConfirmAccountFormComponent],
  templateUrl: './confirm-account.page.html',
  styleUrl: './confirm-account.page.css',
})
export class ConfirmAccountPage {
  private readonly authActionsService = inject(AuthActionsService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly loading = this.authActionsService.loading;
  protected readonly genenralError = this.authActionsService.error;

  protected onConfirmAccount(req: ConfirmAccountResponse): void {
    this.authActionsService
      .confirmAccount(req)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => {
        this.router.navigate(['/login']);
      });
  }

  protected onDismissError(): void {
    this.authActionsService.clearMessages();
  }
}
