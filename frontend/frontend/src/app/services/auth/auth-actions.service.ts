import { inject, Injectable, signal } from '@angular/core';
import { Observable, tap, catchError, finalize, EMPTY } from 'rxjs';

import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/auth/password-change.model';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';
import { ForgotPasswordRequest } from '../../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';

@Injectable({
  providedIn: 'root',
})
export class AuthActionsService {
  private readonly authApiClient = inject(AuthApiClientService);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _passwordChangeResult = signal<boolean | null>(null);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  public forgotPassword(forgotPasswordRequest: ForgotPasswordRequest): Observable<void> {
    this.setLoadingState();

    return this.authApiClient.forgotPasswordRequest(forgotPasswordRequest).pipe(
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to send reset email');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    this.setLoadingState();
    this._passwordChangeResult.set(null);

    return this.authApiClient.confirmPasswordChange(data).pipe(
      tap(() => this._passwordChangeResult.set(true)),
      catchError((err: ApiError) => {
        this._passwordChangeResult.set(false);
        this._error.set(err.message ?? 'Failed to change password');
        return EMPTY;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public confirmPasswordReset(req: ForgotPasswordResponse): Observable<void> {
    this.setLoadingState();

    if (this.authApiClient.verifyForgotPasswordToken(req.token)) {
      return this.authApiClient.confirmPasswordReset(req).pipe(
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to reset password');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      );
    } else {
      this._error.set('Invalid or expired token');
      this._loading.set(false);
      return EMPTY;
    }
  }

  public confirmAccount(req: ConfirmAccountResponse): Observable<void> {
    this.setLoadingState();

    if (this.authApiClient.verifyAccountToken(req.token)) {
      return this.authApiClient.confirmAccountCreation(req).pipe(
        catchError((err: ApiError) => {
          this._error.set(err.message ?? 'Failed to confirm account');
          return EMPTY;
        }),
        finalize(() => this._loading.set(false)),
      );
    } else {
      this._error.set('Invalid or expired token');
      this._loading.set(false);
      return EMPTY;
    }
  }

  public clearMessages(): void {
    this._error.set(null);
    this._passwordChangeResult.set(null);
  }

  private setLoadingState(): void {
    this._loading.set(true);
    this._error.set(null);
  }
}
