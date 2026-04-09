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
  private readonly shouldFailLogin = false;
  private readonly shouldFailLogout = false;
  private readonly shouldFailForgotPasswordRequest = false;
  private readonly shouldFailConfirmPasswordChange = false;
  private readonly shouldFailConfirmPasswordReset = false;
  private readonly shouldFailConfirmAccountCreation = false;
  private readonly shouldFailVerifyForgotPasswordToken = false;
  private readonly shouldFailVerifyAccountToken = false;

  login(req: LoginRequest): Observable<AuthResponse> {
    if (this.shouldFailLogin) {
      return this.delayedError({ status: 500, message: 'Server error during login' });
    }

    const entry = MOCK_USERS[req.email];

    if (
      entry &&
      entry.password === req.password &&
      (entry.tenantId === req.tenantId || !entry.tenantId)
    ) {
      const jwt = this.buildJwt(entry.userId, entry.role, entry.tenantId);
      return of({ jwt }).pipe(delay(MOCK_DELAY));
    }

    return this.delayedError({ status: 401, message: 'Invalid email or password' });
  }

  logout(): Observable<void> {
    if (this.shouldFailLogout) {
      return this.delayedError({ status: 500, message: 'Server error during logout' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  forgotPasswordRequest(req: ForgotPasswordRequest): Observable<void> {
    if (this.shouldFailForgotPasswordRequest) {
      return this.delayedError({
        status: 500,
        message: 'Server error during forgot password request',
      });
    }

    if (!MOCK_USERS[req.email]) {
      return this.delayedError({ status: 404, message: 'Email not found' });
    } else if (req.tenantId && MOCK_USERS[req.email].tenantId !== req.tenantId) {
      return this.delayedError({ status: 404, message: 'Wrong tenant' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmPasswordChange(data: PasswordChange): Observable<void> {
    if (this.shouldFailConfirmPasswordChange) {
      return this.delayedError({ status: 500, message: 'Server error during password change' });
    }

    if (!data.oldPassword) {
      return this.delayedError({ status: 400, message: 'Old password is required' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  verifyForgotPasswordToken(token: string): Observable<void> {
    if (this.shouldFailVerifyForgotPasswordToken) {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }

    if (!token || token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  verifyAccountToken(token: string): Observable<void> {
    if (this.shouldFailVerifyAccountToken) {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }

    if (!token || token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmPasswordReset(req: ForgotPasswordResponse): Observable<void> {
    if (this.shouldFailConfirmPasswordReset) {
      return this.delayedError({ status: 500, message: 'Server error during password reset' });
    }

    if (!req.token || req.token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }
    return of(undefined).pipe(delay(MOCK_DELAY));
  }

  confirmAccountCreation(req: ConfirmAccountResponse): Observable<AuthResponse> {
    if (this.shouldFailConfirmAccountCreation) {
      return this.delayedError({
        status: 500,
        message: 'Server error during account confirmation',
      });
    }

    if (!req.token || req.token === 'expired-token') {
      return this.delayedError({ status: 400, message: 'Invalid or expired token' });
    }

    const entry = Object.values(MOCK_USERS)[0];
    const jwt = this.buildJwt(entry.userId, entry.role, entry.tenantId);
    return of({ jwt }).pipe(delay(MOCK_DELAY));
  }

  private buildJwt(userId: string, role: string, tenantId?: string): string {
    const header = btoa(JSON.stringify({ alg: 'none', typ: 'JWT' }));
    const payload = btoa(
      JSON.stringify({
        uid: userId,
        rol: userRoleMapperJWT.toBackend(userRoleMapper.fromBackend(role)),
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
