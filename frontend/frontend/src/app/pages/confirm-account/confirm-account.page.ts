import { Component, DestroyRef, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router } from '@angular/router';
import { ActivatedRoute } from '@angular/router';

import { AuthActionsService } from '../../services/auth/auth-actions.service';
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
  private readonly activatedRoute = inject(ActivatedRoute);
  private readonly destroyRef = inject(DestroyRef);
  private readonly token = this.activatedRoute.snapshot.queryParamMap.get('token') ?? '';

  protected readonly loading = this.authActionsService.loading;
  protected readonly genenralError = this.authActionsService.error;

  protected onConfirmAccount(req: ConfirmAccountResponse): void {
    const requestWithToken: ConfirmAccountResponse = {
      ...req,
      token: this.token,
    };
    this.authActionsService
      .confirmAccount(requestWithToken)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => {
        this.router.navigate(['/login']);
      });
  }

  protected onDismissError(): void {
    this.authActionsService.clearMessages();
  }
}
