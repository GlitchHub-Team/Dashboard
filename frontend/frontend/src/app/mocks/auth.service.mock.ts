import { Injectable, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, of, delay, throwError, tap, finalize, timer, switchMap } from 'rxjs';
import { inject } from '@angular/core';
import { TokenStorageService } from '../services/token-storage/token-storage.service';
import { UserSessionService } from '../services/user-session/user-session.service';
import { UserRole } from '../models/user-role.enum';
import { LoginRequest } from '../models/login-request.model';
import { AuthResponse } from '../models/auth-response.model';
import { PasswordChange } from '../models/password-change.model';
import { ApiError } from '../models/api-error.model';

@Injectable({ providedIn: 'root' })
export class AuthServiceMock {
  private readonly tokenStorage = inject(TokenStorageService);
  private readonly userSession = inject(UserSessionService);
  private readonly router = inject(Router);

  // ---- State ----
  private readonly _loading = signal(false);
  private readonly _error = signal<string | null>(null);
  private readonly _passwordChangeResult = signal<boolean | null>(null);

  // ---- Selectors ----
  public readonly loading = this._loading.asReadonly();
  public readonly error = this._error.asReadonly();
  public readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  public readonly isAuthenticated = computed<boolean>(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

  // ---- AuthSessionService methods ----

  public login(req: LoginRequest): Observable<AuthResponse> {
    this._loading.set(true);
    this._error.set(null);

    if (req.email === 'admin@test.com' && req.password === 'password') {
      const response: AuthResponse = {
        user: { id: 1, name: 'Admin', email: req.email, role: UserRole.SUPER_ADMIN, tenantId: 1 },
        token: this.fakeToken(req.email, UserRole.SUPER_ADMIN),
      };

      return of(response).pipe(
        delay(800),
        tap((res) => {
          this.tokenStorage.saveToken(res.token);
          this.userSession.initSession(res.user);
        }),
        finalize(() => this._loading.set(false)),
      );
    }

    const error: ApiError = {
      status: 401,
      message: 'Invalid email or password',
    };

    // timer() ensures the delay actually works for errors
    return timer(800).pipe(
      switchMap(() => {
        this._error.set(error.message);
        return throwError(() => error);
      }),
      finalize(() => this._loading.set(false)),
    );
  }

  public logout(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this._passwordChangeResult.set(null);
    this.router.navigate(['/login']);
  }

  public clearError(): void {
    this._error.set(null);
  }

  // ---- AuthActionsService methods ----

  public forgotPassword(email: string): Observable<void> {
    this._loading.set(true);
    this._error.set(null);

    if (email === 'notfound@test.com') {
      const error: ApiError = {
        status: 404,
        message: 'Email not found',
        errors: [{ field: 'email', message: 'No account with this email' }],
      };

      return timer(800).pipe(
        switchMap(() => throwError(() => error)),
        finalize(() => this._loading.set(false)),
      );
    }

    return of(undefined).pipe(
      delay(800),
      finalize(() => this._loading.set(false)),
    );
  }

  public requestPasswordChange(): Observable<void> {
    const user = this.userSession.currentUser();

    if (!user) {
      const error: ApiError = {
        status: 401,
        message: 'User not authenticated',
      };

      return timer(800).pipe(switchMap(() => throwError(() => error)));
    }

    this._loading.set(true);
    return of(undefined).pipe(
      delay(800),
      finalize(() => this._loading.set(false)),
    );
  }

  public confirmPasswordChange(data: PasswordChange): Observable<void> {
    this._loading.set(true);
    this._error.set(null);

    if (data.token === 'valid-token') {
      this._passwordChangeResult.set(true);
      return of(undefined).pipe(
        delay(800),
        finalize(() => this._loading.set(false)),
      );
    }

    const error: ApiError = {
      status: 400,
      message: 'Invalid or expired token',
    };
    this._passwordChangeResult.set(false);

    return timer(800).pipe(
      switchMap(() => throwError(() => error)),
      finalize(() => this._loading.set(false)),
    );
  }

  public clearMessages(): void {
    this._error.set(null);
  }

  // ---- Helper ----

  private fakeToken(email: string, role: UserRole): string {
    const header = btoa(JSON.stringify({ alg: 'none' }));
    const payload = btoa(
      JSON.stringify({
        email,
        role,
        exp: Math.floor(Date.now() / 1000) + 3600,
      }),
    );
    return `${header}.${payload}.fake`;
  }
}
