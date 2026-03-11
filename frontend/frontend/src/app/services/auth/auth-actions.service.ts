import { inject, Injectable, signal } from '@angular/core';
import { Observable, throwError, tap, catchError, finalize } from 'rxjs';
import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/password-change.model';
import { UserSessionService } from '../user-session/user-session.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';

@Injectable({
  providedIn: 'root',
})
export class AuthActionsService {
  private readonly authApiClient = inject(AuthApiClientService);
  private readonly userSession = inject(UserSessionService);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _passwordChangeResult = signal<boolean | null>(null);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  public forgotPassword(email: string): Observable<void> {
    this._loading.set(true);
    this._error.set(null);

    return this.authApiClient.forgotPassword(email).pipe(
      catchError((err: ApiError) => {
        if (!err.errors?.length) {
          this._error.set(err.message ?? 'Failed to send reset email');
        }
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public requestPasswordChange(): Observable<void> {
    const user = this.userSession.currentUser();

    if (!user?.id) {
      return throwError(
        () =>
          ({
            status: 401,
            message: 'User not authenticated',
          }) as ApiError,
      );
    }

    this._loading.set(true);
    this._error.set(null);

    return this.authApiClient.requestPasswordChange(user.id).pipe(
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Failed to request password change');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    this._loading.set(true);
    this._error.set(null);
    this._passwordChangeResult.set(null);

    return this.authApiClient.confirmPasswordChange(data).pipe(
      tap(() => this._passwordChangeResult.set(true)),
      catchError((err: ApiError) => {
        this._passwordChangeResult.set(false);
        this._error.set(err.message ?? 'Failed to change password');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public clearMessages(): void {
    this._error.set(null);
    this._passwordChangeResult.set(null);
  }
}
