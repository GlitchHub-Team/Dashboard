import { computed, inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, tap, catchError, finalize } from 'rxjs';

import { UserSessionService } from '../user-session/user-session.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { LoginRequest } from '../../models/login-request.model';
import { AuthResponse } from '../../models/auth-response.model';
import { ApiError } from '../../models/api-error.model';

@Injectable({
  providedIn: 'root',
})
export class AuthSessionService {
  private readonly authApiClient = inject(AuthApiClientService);
  private readonly tokenStorage = inject(TokenStorageService);
  private readonly userSession = inject(UserSessionService);
  private readonly router = inject(Router);

  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);

  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  // TODO: Da rivedere per come utilizziamo il token che viene (forse) inviato dal backend
  public readonly isAuthenticated = computed(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

  public login(req: LoginRequest): Observable<AuthResponse> {
    this._loading.set(true);
    this._error.set(null);

    return this.authApiClient.login(req).pipe(
      tap((response) => {
        this.tokenStorage.saveToken(response.token);
        this.userSession.initSession(response.user);
      }),
      catchError((err: ApiError) => {
        this._error.set(err.message ?? 'Login failed');
        throw err;
      }),
      finalize(() => this._loading.set(false)),
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

  public clearError(): void {
    this._error.set(null);
  }

  private clearAndRedirect(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this.router.navigate(['/login']);
  }
}
