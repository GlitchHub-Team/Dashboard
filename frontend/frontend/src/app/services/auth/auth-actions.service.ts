import { inject, Injectable, signal } from '@angular/core';
import { Observable } from 'rxjs/internal/Observable';
import { throwError } from 'rxjs/internal/observable/throwError';
import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/password-change.model';
import { tap } from 'rxjs/internal/operators/tap';
import { UserSessionService } from '../user-session/user-session.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';

@Injectable({
  providedIn: 'root',
})
export class AuthActionsService {
  private authApiClient = inject(AuthApiClientService);
  private userSession = inject(UserSessionService);

  private _passwordChangeResult = signal<boolean | null>(null);
  public readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  public forgotPassword(email: string): Observable<void> {
    return this.authApiClient.forgotPassword(email);
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
    return this.authApiClient.requestPasswordChange(user.id);
  }

  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    return this.authApiClient.confirmPasswordChange(data).pipe(
      tap({
        next: () => this._passwordChangeResult.set(true),
        error: () => this._passwordChangeResult.set(false),
      }),
    );
  }
}
