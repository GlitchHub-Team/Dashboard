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

  const authApiClientMock = {
    login: vi.fn(),
    logout: vi.fn(),
  };

  const tokenStorageMock = {
    saveToken: vi.fn(),
    clearToken: vi.fn(),
    isValid: vi.fn(),
  };

  const userSessionMock = {
    initSession: vi.fn(),
    clearSession: vi.fn(),
    currentUser: vi.fn(),
  };

  const routerMock = {
    navigate: vi.fn(),
  };

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
    it('should be created', () => {
      expect(service).toBeTruthy();
    });

    it('should start with loading false', () => {
      expect(service.loading()).toBe(false);
    });

    it('should start with no error', () => {
      expect(service.error()).toBeNull();
    });
  });

  describe('isAuthenticated', () => {
    it('should be true when token is valid and user exists', () => {
      tokenStorageMock.isValid.mockReturnValue(true);
      userSessionMock.currentUser.mockReturnValue({
        id: '1',
        email: 'test@test.com',
        role: UserRole.SUPER_ADMIN,
        tenantId: 'tenant-1',
      });

      expect(service.isAuthenticated()).toBe(true);
    });

    it('should be false when token is invalid', () => {
      tokenStorageMock.isValid.mockReturnValue(false);
      userSessionMock.currentUser.mockReturnValue({ id: '1' });

      expect(service.isAuthenticated()).toBe(false);
    });

    it('should be false when user is null', () => {
      tokenStorageMock.isValid.mockReturnValue(true);
      userSessionMock.currentUser.mockReturnValue(null);

      expect(service.isAuthenticated()).toBe(false);
    });

    it('should be false when both are missing', () => {
      tokenStorageMock.isValid.mockReturnValue(false);
      userSessionMock.currentUser.mockReturnValue(null);

      expect(service.isAuthenticated()).toBe(false);
    });
  });

  describe('login', () => {
    const mockRequest: LoginRequest = {
      email: 'user@example.com',
      password: 'secret123',
    };

    const mockResponse: AuthResponse = {
      token: 'jwt-token-abc',
      user: {
        id: '1',
        email: 'user@example.com',
        role: UserRole.SUPER_ADMIN,
        tenantId: 'tenant-1',
      },
    };

    it('should call authApiClient.login with the request', () => {
      authApiClientMock.login.mockReturnValue(of(mockResponse));

      service.login(mockRequest).subscribe();

      expect(authApiClientMock.login).toHaveBeenCalledWith(mockRequest);
    });

    it('should save token on success', () => {
      authApiClientMock.login.mockReturnValue(of(mockResponse));

      service.login(mockRequest).subscribe();

      expect(tokenStorageMock.saveToken).toHaveBeenCalledWith('jwt-token-abc');
    });

    it('should init user session on success', () => {
      authApiClientMock.login.mockReturnValue(of(mockResponse));

      service.login(mockRequest).subscribe();

      expect(userSessionMock.initSession).toHaveBeenCalledWith(mockResponse.user);
    });

    it('should return the auth response', () => {
      authApiClientMock.login.mockReturnValue(of(mockResponse));

      service.login(mockRequest).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });
    });

    it('should set loading false after success', () => {
      authApiClientMock.login.mockReturnValue(of(mockResponse));

      service.login(mockRequest).subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should clear previous error before request', () => {
      // Trigger an error first
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 401, message: 'bad' }) as ApiError),
      );
      service.login(mockRequest).subscribe({ error: noop });
      expect(service.error()).toBe('bad');

      // New request should clear it
      authApiClientMock.login.mockReturnValue(of(mockResponse));
      service.login(mockRequest).subscribe();
      expect(service.error()).toBeNull();
    });

    it('should set error message on failure', () => {
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 401, message: 'Invalid credentials' }) as ApiError),
      );

      service.login(mockRequest).subscribe({ error: noop });

      expect(service.error()).toBe('Invalid credentials');
    });

    it('should set default error when message is empty', () => {
      authApiClientMock.login.mockReturnValue(throwError(() => ({ status: 500 }) as ApiError));

      service.login(mockRequest).subscribe({ error: noop });

      expect(service.error()).toBe('Login failed');
    });

    it('should rethrow the error', () => {
      const apiError: ApiError = { status: 401, message: 'fail' };
      authApiClientMock.login.mockReturnValue(throwError(() => apiError));

      service.login(mockRequest).subscribe({
        error: (err) => {
          expect(err).toBe(apiError);
        },
      });
    });

    it('should not save token on failure', () => {
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 401, message: 'fail' }) as ApiError),
      );

      service.login(mockRequest).subscribe({ error: noop });

      expect(tokenStorageMock.saveToken).not.toHaveBeenCalled();
    });

    it('should not init session on failure', () => {
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 401, message: 'fail' }) as ApiError),
      );

      service.login(mockRequest).subscribe({ error: noop });

      expect(userSessionMock.initSession).not.toHaveBeenCalled();
    });

    it('should set loading false after error', () => {
      authApiClientMock.login.mockReturnValue(
        throwError(() => ({ status: 500, message: 'fail' }) as ApiError),
      );

      service.login(mockRequest).subscribe({ error: noop });

      expect(service.loading()).toBe(false);
    });
  });

  describe('logout', () => {
    const mockUser = {
      id: 'user-42',
      email: 'user@example.com',
      role: UserRole.SUPER_ADMIN,
      tenantId: 'tenant-1',
    };

    describe('when user exists', () => {
      beforeEach(() => {
        userSessionMock.currentUser.mockReturnValue(mockUser);
      });

      it('should call authApiClient.logout with user id', () => {
        authApiClientMock.logout.mockReturnValue(of(undefined));

        service.logout();

        expect(authApiClientMock.logout).toHaveBeenCalledWith('user-42');
      });

      it('should clear token after successful API call', () => {
        authApiClientMock.logout.mockReturnValue(of(undefined));

        service.logout();

        expect(tokenStorageMock.clearToken).toHaveBeenCalled();
      });

      it('should clear session after successful API call', () => {
        authApiClientMock.logout.mockReturnValue(of(undefined));

        service.logout();

        expect(userSessionMock.clearSession).toHaveBeenCalled();
      });

      it('should navigate to /login after successful API call', () => {
        authApiClientMock.logout.mockReturnValue(of(undefined));

        service.logout();

        expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
      });

      it('should still clear and redirect on API error', () => {
        authApiClientMock.logout.mockReturnValue(
          throwError(() => ({ status: 500, message: 'Server error' })),
        );

        service.logout();

        expect(tokenStorageMock.clearToken).toHaveBeenCalled();
        expect(userSessionMock.clearSession).toHaveBeenCalled();
        expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
      });
    });

    describe('when user is null', () => {
      beforeEach(() => {
        userSessionMock.currentUser.mockReturnValue(null);
      });

      it('should not call authApiClient.logout', () => {
        service.logout();

        expect(authApiClientMock.logout).not.toHaveBeenCalled();
      });

      it('should still clear token', () => {
        service.logout();

        expect(tokenStorageMock.clearToken).toHaveBeenCalled();
      });

      it('should still clear session', () => {
        service.logout();

        expect(userSessionMock.clearSession).toHaveBeenCalled();
      });

      it('should still navigate to /login', () => {
        service.logout();

        expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
      });
    });

    describe('when user has no id', () => {
      beforeEach(() => {
        userSessionMock.currentUser.mockReturnValue({ ...mockUser, id: '' });
      });

      it('should not call authApiClient.logout', () => {
        service.logout();

        expect(authApiClientMock.logout).not.toHaveBeenCalled();
      });

      it('should still clear and redirect', () => {
        service.logout();

        expect(tokenStorageMock.clearToken).toHaveBeenCalled();
        expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
      });
    });
  });

  describe('clearError', () => {
    it('should reset error to null', () => {
      // Trigger an error first
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
