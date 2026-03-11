import { computed, inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { UserSessionService } from '../user-session/user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { LoginRequest } from '../../models/login-request.model';
import { PasswordChange } from '../../models/password-change.model';
import { tap, Observable, throwError } from 'rxjs';
import { AuthResponse } from '../../models/auth-response.model';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private authApiClient = inject(AuthApiClientService);
  private tokenStorage = inject(TokenStorageService);
  private userSession = inject(UserSessionService);
  private router = inject(Router);

  // TODO: Da rivedere per come utilizziamo il token che viene (forse) inviato dal backend
  public readonly isAuthenticated = computed(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

  private _passwordChangeResult = signal<boolean | null>(null);
  readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  // Anche il service deve ritornare un Observable, così da poter gestire errori e successi in modo più flessibile nei componenti che lo utilizzano

  public login(req: LoginRequest): Observable<AuthResponse> {
    return this.authApiClient.login(req).pipe(
      tap((response) => {
        this.tokenStorage.saveToken(response.token);
        this.userSession.initSession(response.user);
      }),
    );
  }

  public logout(): void {
    const user = this.userSession.currentUser();

    if (user?.id) {
      this.authApiClient.logout(user.id).subscribe({
        next: () => this.clearAndRedirect(),
        error: () => this.clearAndRedirect(),
      });
    } else {
      this.clearAndRedirect();
    }
  }

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

  private clearAndRedirect(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this._passwordChangeResult.set(null);
    this.router.navigate(['/login']);
  }
}
