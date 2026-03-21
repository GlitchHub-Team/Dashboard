import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of, throwError, noop } from 'rxjs';

import { AuthSessionService } from './auth-session.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { TokenStorageService } from '../token-storage/token-storage.service';
import { UserSessionService } from '../user-session/user-session.service';
import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { ApiError } from '../../models/api-error.model';
import { UserRole } from '../../models/user-role.enum';

describe('AuthSessionService', () => {
  let service: AuthSessionService;

  const authApiClientMock = { login: vi.fn(), logout: vi.fn() };
  const tokenStorageMock = { saveToken: vi.fn(), clearToken: vi.fn(), isValid: vi.fn() };
  const userSessionMock = { initSession: vi.fn(), clearSession: vi.fn(), currentUser: vi.fn() };
  const routerMock = { navigate: vi.fn() };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        AuthSessionService,
        { provide: AuthApiClientService, useValue: authApiClientMock },
        { provide: TokenStorageService, useValue: tokenStorageMock },
        { provide: UserSessionService, useValue: userSessionMock },
        { provide: Router, useValue: routerMock },
      ],
    });

    service = TestBed.inject(AuthSessionService);
  });

  describe('initial state', () => {
    it('should be created with loading=false and error=null', () => {
      expect(service).toBeTruthy();
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });
  });

  describe('isAuthenticated', () => {
    it.each([
      [true, { id: '1', email: 'test@test.com', role: UserRole.SUPER_ADMIN, tenantId: 't' }, true],
      [false, { id: '1' }, false],
      [true, null, false],
      [false, null, false],
    ] as const)('isValid=%s, user=%o to %s', (isValid, user, expected) => {
      tokenStorageMock.isValid.mockReturnValue(isValid);
      userSessionMock.currentUser.mockReturnValue(user);
      expect(service.isAuthenticated()).toBe(expected);
    });
  });

  describe('login', () => {
    const mockRequest: LoginRequest = { email: 'user@example.com', password: 'secret123' };
    const mockResponse: AuthResponse = {
      token: 'jwt-token-abc',
      user: {
        id: '1',
        email: 'user@example.com',
        role: UserRole.SUPER_ADMIN,
        tenantId: 'tenant-1',
      },
    };

    it('should call API, save token, init session, return response, and clear loading on success', () => {
      authApiClientMock.login.mockReturnValue(of(mockResponse));

      service.login(mockRequest).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      expect(authApiClientMock.login).toHaveBeenCalledWith(mockRequest);
      expect(tokenStorageMock.saveToken).toHaveBeenCalledWith('jwt-token-abc');
      expect(userSessionMock.initSession).toHaveBeenCalledWith(mockResponse.user);
      expect(service.loading()).toBe(false);
    });

    it('should clear previous error before a new request', () => {
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 401, message: 'bad' }) as ApiError),
      );
      service.login(mockRequest).subscribe({ error: noop });
      expect(service.error()).toBe('bad');

      authApiClientMock.login.mockReturnValue(of(mockResponse));
      service.login(mockRequest).subscribe();
      expect(service.error()).toBeNull();
    });

    it.each([
      [{ status: 401, message: 'Invalid credentials' } as ApiError, 'Invalid credentials'],
      [{ status: 500 } as ApiError, 'Login failed'],
    ])('should set error on failure and rethrow (apiError=%o)', (apiError, expectedMsg) => {
      authApiClientMock.login.mockReturnValue(throwError(() => apiError));

      service.login(mockRequest).subscribe({
        error: (err) => expect(err).toBe(apiError),
      });

      expect(service.error()).toBe(expectedMsg);
      expect(service.loading()).toBe(false);
      expect(tokenStorageMock.saveToken).not.toHaveBeenCalled();
      expect(userSessionMock.initSession).not.toHaveBeenCalled();
    });
  });

  describe('logout', () => {
    const mockUser = {
      id: 'user-42',
      email: 'user@example.com',
      role: UserRole.SUPER_ADMIN,
      tenantId: 'tenant-1',
    };

    it('should call logout API, clear token/session, and navigate on success', () => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.logout.mockReturnValue(of(undefined));

      service.logout();

      expect(authApiClientMock.logout).toHaveBeenCalledWith('user-42');
      expect(tokenStorageMock.clearToken).toHaveBeenCalled();
      expect(userSessionMock.clearSession).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });

    it('should still clear and redirect even on API error', () => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.logout.mockReturnValue(throwError(() => ({ status: 500 })));

      service.logout();

      expect(tokenStorageMock.clearToken).toHaveBeenCalled();
      expect(userSessionMock.clearSession).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });

    it.each([
      ['null user', null],
      ['user with no id', { ...mockUser, id: '' }],
    ])('should skip API call but still clear and redirect when %s', (_, user) => {
      userSessionMock.currentUser.mockReturnValue(user);

      service.logout();

      expect(authApiClientMock.logout).not.toHaveBeenCalled();
      expect(tokenStorageMock.clearToken).toHaveBeenCalled();
      expect(userSessionMock.clearSession).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('clearError', () => {
    it('should reset error to null', () => {
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 401, message: 'some error' }) as ApiError),
      );
      service.login({ email: 'x', password: 'y' }).subscribe({ error: noop });
      expect(service.error()).toBe('some error');

      service.clearError();

      expect(service.error()).toBeNull();
    });
  });
});
