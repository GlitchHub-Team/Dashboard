import { Injectable, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, of, delay, throwError } from 'rxjs';
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
  private tokenStorage = inject(TokenStorageService);
  private userSession = inject(UserSessionService);
  private router = inject(Router);

  readonly isAuthenticated = computed<boolean>(
    () => this.tokenStorage.isValid() && this.userSession.currentUser() !== null,
  );

  private _passwordChangeResult = signal<boolean | null>(null);
  readonly passwordChangeResult = this._passwordChangeResult.asReadonly();

  login(req: LoginRequest): Observable<AuthResponse> {
    // Simulate success
    if (req.email === 'admin@test.com' && req.password === 'password') {
      const response: AuthResponse = {
        user: { id: 1, name: 'Admin', email: req.email, role: UserRole.SUPER_ADMIN, tenantId: 1 },
        token: this.fakeToken(req.email, UserRole.SUPER_ADMIN),
      };
      this.tokenStorage.saveToken(response.token);
      this.userSession.initSession(response.user);
      return of(response).pipe(delay(800));
    }

    // Simulate error
    const error: ApiError = {
      status: 401,
      message: 'Invalid email or password',
    };
    return throwError(() => error).pipe(delay(800));
  }

  forgotPassword(email: string): Observable<void> {
    if (email === 'notfound@test.com') {
      const error: ApiError = {
        status: 404,
        message: 'Email not found',
        errors: [{ field: 'email', message: 'No account with this email' }],
      };
      return throwError(() => error).pipe(delay(800));
    }
    return of(undefined).pipe(delay(800));
  }

  requestPasswordChange(): Observable<void> {
    const user = this.userSession.currentUser();
    if (!user) {
      const error: ApiError = {
        status: 401,
        message: 'User not authenticated',
      };
      return throwError(() => error).pipe(delay(800));
    }
    return of(undefined).pipe(delay(800));
  }

  confirmPasswordChange(data: PasswordChange): Observable<void> {
    if (data.token === 'valid-token') {
      this._passwordChangeResult.set(true);
      return of(undefined).pipe(delay(800));
    } else {
      const error: ApiError = {
        status: 400,
        message: 'Invalid or expired token',
      };
      this._passwordChangeResult.set(false);
      return throwError(() => error).pipe(delay(800));
    }
  }

  logout(): void {
    this.tokenStorage.clearToken();
    this.userSession.clearSession();
    this._passwordChangeResult.set(null);
    this.router.navigate(['/login']);
  }

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
