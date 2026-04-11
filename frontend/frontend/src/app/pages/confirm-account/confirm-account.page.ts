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
  // Token e tenantId recuperati dall'URL
  private readonly token = this.activatedRoute.snapshot.paramMap.get('token') ?? '';
  private readonly tenantId = this.activatedRoute.snapshot.queryParamMap.get('tid') ?? undefined;

  protected readonly loading = this.authActionsService.loading;
  protected readonly generalError = this.authActionsService.error;

  protected onConfirmAccount(req: ConfirmAccountResponse): void {
    const requestWithToken: ConfirmAccountResponse = {
      ...req,
      token: this.token,
      tenantId: this.tenantId,
    };
    // La conferma di un account ritorna il JWT dell'utente appena confermato,
    // quindi logga automaticamente l'utente dopo la conferma
    this.authActionsService
      .confirmAccount(requestWithToken)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => this.router.navigate(['/dashboard']),
      });
  }

  protected onDismissError(): void {
    this.authActionsService.clearMessages();
  }
}
