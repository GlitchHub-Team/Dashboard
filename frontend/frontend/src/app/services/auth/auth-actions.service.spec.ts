import { TestBed } from '@angular/core/testing';
import { of, throwError, noop } from 'rxjs';

import { AuthActionsService } from './auth-actions.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { UserSessionService } from '../user-session/user-session.service';
import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/password-change.model';

describe('AuthActionsService', () => {
  let service: AuthActionsService;

  const authApiClientMock = {
    forgotPassword: vi.fn(),
    requestPasswordChange: vi.fn(),
    confirmPasswordChange: vi.fn(),
  };

  const userSessionMock = {
    currentUser: vi.fn(),
  };

  beforeEach(() => {
    // Reset all mocks between tests
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        AuthActionsService,
        { provide: AuthApiClientService, useValue: authApiClientMock },
        { provide: UserSessionService, useValue: userSessionMock },
      ],
    });

    service = TestBed.inject(AuthActionsService);
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

    it('should start with no password change result', () => {
      expect(service.passwordChangeResult()).toBeNull();
    });
  });

  describe('forgotPassword', () => {
    const email = 'user@example.com';

    it('should call authApiClient.forgotPassword with the email', () => {
      authApiClientMock.forgotPassword.mockReturnValue(of(undefined));

      service.forgotPassword(email).subscribe();

      expect(authApiClientMock.forgotPassword).toHaveBeenCalledWith(email);
    });

    it('should set loading false after success', () => {
      authApiClientMock.forgotPassword.mockReturnValue(of(undefined));

      service.forgotPassword(email).subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should clear previous error before request', () => {
      // First: trigger an error
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500, message: 'fail' }) as ApiError),
      );
      service.forgotPassword(email).subscribe({
        error: noop,
      });
      expect(service.error()).toBe('fail');

      // Second: new request should clear it
      authApiClientMock.forgotPassword.mockReturnValue(of(undefined));
      service.forgotPassword(email).subscribe();
      expect(service.error()).toBeNull();
    });

    it('should set error message on failure', () => {
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Email service down' }) as ApiError),
      );

      service.forgotPassword(email).subscribe({
        error: noop,
      });

      expect(service.error()).toBe('Email service down');
    });

    it('should set default error when message is empty', () => {
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500 }) as ApiError),
      );

      service.forgotPassword(email).subscribe({
        error: noop,
      });

      expect(service.error()).toBe('Failed to send reset email');
    });

    it('should rethrow the error', () => {
      const apiError: ApiError = { status: 500, message: 'fail' };
      authApiClientMock.forgotPassword.mockReturnValue(throwError(() => apiError));

      service.forgotPassword(email).subscribe({
        error: (err) => {
          expect(err).toBe(apiError);
        },
      });
    });

    it('should set loading false after error', () => {
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500, message: 'fail' }) as ApiError),
      );

      service.forgotPassword(email).subscribe({
        error: noop,
      });

      expect(service.loading()).toBe(false);
    });
  });

  describe('requestPasswordChange', () => {
    const mockUser = {
      id: 'user-42',
      email: 'user@example.com',
      role: 'admin',
      tenantId: 'tenant-1',
    };

    it('should call authApiClient with the user id', () => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.requestPasswordChange.mockReturnValue(of(undefined));

      service.requestPasswordChange().subscribe();

      expect(authApiClientMock.requestPasswordChange).toHaveBeenCalledWith('user-42');
    });

    it('should return 401 error if user is null', () => {
      userSessionMock.currentUser.mockReturnValue(null);

      service.requestPasswordChange().subscribe({
        error: (err: ApiError) => {
          expect(err.status).toBe(401);
          expect(err.message).toBe('User not authenticated');
        },
      });

      expect(authApiClientMock.requestPasswordChange).not.toHaveBeenCalled();
    });

    it('should return 401 error if user has no id', () => {
      userSessionMock.currentUser.mockReturnValue({ ...mockUser, id: '' });

      service.requestPasswordChange().subscribe({
        error: (err: ApiError) => {
          expect(err.status).toBe(401);
        },
      });

      expect(authApiClientMock.requestPasswordChange).not.toHaveBeenCalled();
    });

    it('should not set loading when user is missing', () => {
      userSessionMock.currentUser.mockReturnValue(null);

      service.requestPasswordChange().subscribe({
        error: noop,
      });

      expect(service.loading()).toBe(false);
    });

    it('should set error on API failure', () => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.requestPasswordChange.mockReturnValue(
        throwError(() => ({ status: 500, message: 'Server error' }) as ApiError),
      );

      service.requestPasswordChange().subscribe({
        error: noop,
      });

      expect(service.error()).toBe('Server error');
      expect(service.loading()).toBe(false);
    });

    it('should set default error when message is empty', () => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.requestPasswordChange.mockReturnValue(
        throwError(() => ({ status: 500 }) as ApiError),
      );

      service.requestPasswordChange().subscribe({
        error: noop,
      });

      expect(service.error()).toBe('Failed to request password change');
    });
  });

  describe('confirmPasswordChange', () => {
    const mockData: PasswordChange = {
      token: 'reset-token',
      newPassword: 'newSecret456',
    };

    it('should call authApiClient with the data', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));

      service.confirmPasswordChange(mockData).subscribe();

      expect(authApiClientMock.confirmPasswordChange).toHaveBeenCalledWith(mockData);
    });

    it('should set passwordChangeResult to true on success', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));

      service.confirmPasswordChange(mockData).subscribe();

      expect(service.passwordChangeResult()).toBe(true);
    });

    it('should set loading false after success', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));

      service.confirmPasswordChange(mockData).subscribe();

      expect(service.loading()).toBe(false);
    });

    it('should reset passwordChangeResult to null before request starts', () => {
      // Simulate a previous result
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));
      service.confirmPasswordChange(mockData).subscribe();
      expect(service.passwordChangeResult()).toBe(true);

      // Track intermediate state via a side effect
      // ESlint whines about explicit type any but it's the only way to capture the value during the request
      let resultDuringRequest: boolean | null = 'not-checked' as any;
      authApiClientMock.confirmPasswordChange.mockImplementation(() => {
        resultDuringRequest = service.passwordChangeResult();
        return of(undefined);
      });

      service.confirmPasswordChange(mockData).subscribe();

      expect(resultDuringRequest).toBeNull();
    });

    it('should set passwordChangeResult to false on error', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(
        throwError(() => ({ status: 400, message: 'Invalid token' }) as ApiError),
      );

      service.confirmPasswordChange(mockData).subscribe({
        error: noop,
      });

      expect(service.passwordChangeResult()).toBe(false);
    });

    it('should set error message on failure', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(
        throwError(() => ({ status: 400, message: 'Invalid token' }) as ApiError),
      );

      service.confirmPasswordChange(mockData).subscribe({
        error: noop,
      });

      expect(service.error()).toBe('Invalid token');
    });

    it('should set default error when message is empty', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(
        throwError(() => ({ status: 400 }) as ApiError),
      );

      service.confirmPasswordChange(mockData).subscribe({
        error: noop,
      });

      expect(service.error()).toBe('Failed to change password');
    });

    it('should set loading false after error', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(
        throwError(() => ({ status: 400, message: 'fail' }) as ApiError),
      );

      service.confirmPasswordChange(mockData).subscribe({
        error: noop,
      });

      expect(service.loading()).toBe(false);
    });
  });

  describe('clearMessages', () => {
    it('should clear error', () => {
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500, message: 'some error' }) as ApiError),
      );
      service.forgotPassword('x').subscribe({
        error: noop,
      });
      expect(service.error()).toBe('some error');

      service.clearMessages();

      expect(service.error()).toBeNull();
    });

    it('should clear passwordChangeResult', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));
      service.confirmPasswordChange({ token: 'x', newPassword: 'y' }).subscribe();
      expect(service.passwordChangeResult()).toBe(true);

      service.clearMessages();

      expect(service.passwordChangeResult()).toBeNull();
    });
  });
});
