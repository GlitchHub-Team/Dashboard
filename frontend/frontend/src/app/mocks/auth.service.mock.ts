import { Injectable } from '@angular/core';
import { Observable, of, delay, throwError, timer, switchMap } from 'rxjs';

import { LoginRequest } from '../models/auth/login-request.model';
import { AuthResponse } from '../models/auth/auth-response.model';
import { PasswordChange } from '../models/auth/password-change.model';
import { ApiError } from '../models/api-error.model';
import { UserRole } from '../models/user-role.enum';

const MOCK_DELAY = 800;

const MOCK_USERS: Record<string, { password: string; user: AuthResponse['user'] }> = {
  'super@test.com': {
    password: 'password',
    user: { id: '1', email: 'super@test.com', role: UserRole.SUPER_ADMIN },
  },
  'admin@test.com': {
    password: 'password',
    user: { id: '1', email: 'admin@test.com', role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' },
  },
  'user@test.com': {
    password: 'password',
    user: { id: '2', email: 'user@test.com', role: UserRole.TENANT_USER, tenantId: 'tenant-1' },
  },
};

@Injectable({ providedIn: 'root' })
export class AuthServiceMock {
  login(req: LoginRequest): Observable<AuthResponse> {
    const entry = MOCK_USERS[req.email];

    if (entry && entry.password === req.password) {
      const response: AuthResponse = {
        user: entry.user,
        token: this.fakeToken(entry.user.email, entry.user.role),
      };
      return of(response).pipe(delay(MOCK_DELAY));
    }

    return this.delayedError({ status: 401, message: 'Invalid email or password' });
  }

  logout(_userId: string): Observable<void> {
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  forgotPassword(email: string): Observable<void> {
    if (!MOCK_USERS[email]) {
      return this.delayedError({ status: 404, message: 'Email not found' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  requestPasswordChange(_userId: string): Observable<void> {
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmPasswordChange(data: PasswordChange): Observable<void> {
    if (!data.token || data.token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  private delayedError(error: ApiError): Observable<never> {
    return timer(MOCK_DELAY).pipe(switchMap(() => throwError(() => error)));
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
