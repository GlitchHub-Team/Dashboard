import { Injectable } from '@angular/core';
import { Observable, of, delay, throwError, timer, switchMap } from 'rxjs';

import { LoginRequest } from '../models/auth/login-request.model';
import { AuthResponse } from '../models/auth/auth-response.model';
import { PasswordChange } from '../models/auth/password-change.model';
import { ForgotPasswordRequest } from '../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../models/auth/forgot-password.model';
import { ConfirmAccountResponse } from '../models/auth/confirm-account.model';
import { ApiError } from '../models/api-error.model';
import { userRoleMapper } from '../utils/user-role.utils';
import { userRoleMapperJWT } from '../utils/user-role-jwt.utils';

const MOCK_DELAY = 800;

interface MockUser {
  password: string;
  userId: string;
  role: string;
  tenantId?: string;
}

const MOCK_USERS: Record<string, MockUser> = {
  'super@test.com': {
    password: 'password',
    userId: '1',
    role: 'super_admin',
  },
  'admin@test.com': {
    password: 'password',
    userId: '2',
    role: 'tenant_admin',
    tenantId: 'tenant-1',
  },
  'user@test.com': {
    password: 'password',
    userId: '3',
    role: 'tenant_user',
    tenantId: 'tenant-1',
  },
};

@Injectable({ providedIn: 'root' })
export class AuthServiceMock {
  login(req: LoginRequest): Observable<AuthResponse> {
    const entry = MOCK_USERS[req.email];

    if (
      entry &&
      entry.password === req.password &&
      entry.role === req.userRole &&
      (entry.tenantId === req.tenantId || !entry.tenantId)
    ) {
      const token = this.buildJwt(entry.userId, entry.role, entry.tenantId);
      return of({ token }).pipe(delay(MOCK_DELAY));
    }

    return this.delayedError({ status: 401, message: 'Invalid email or password' });
  }

  logout(): Observable<void> {
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  forgotPassword(req: ForgotPasswordRequest): Observable<void> {
    if (!MOCK_USERS[req.email]) {
      return this.delayedError({ status: 404, message: 'Email not found' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmPasswordChange(data: PasswordChange): Observable<void> {
    if (!data.oldPassword) {
      return this.delayedError({ status: 400, message: 'Old password is required' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmPasswordReset(req: ForgotPasswordResponse): Observable<void> {
    if (!req.token || req.token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmAccount(req: ConfirmAccountResponse): Observable<void> {
    if (!req.token || req.token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  private buildJwt(userId: string, role: string, tenantId?: string): string {
    const header = btoa(JSON.stringify({ alg: 'none', typ: 'JWT' }));
    const payload = btoa(
      JSON.stringify({
        uid: userId,
        rol: userRoleMapperJWT.toBackend(userRoleMapper.fromBackend(role)), // Oscenità assurda ma dovrebbe funzionare
        tid: tenantId,
        exp: Math.floor(Date.now() / 1000) + 3600,
      }),
    );
    return `${header}.${payload}.mock`;
  }

  private delayedError(error: ApiError): Observable<never> {
    return timer(MOCK_DELAY).pipe(switchMap(() => throwError(() => error)));
  }
}
