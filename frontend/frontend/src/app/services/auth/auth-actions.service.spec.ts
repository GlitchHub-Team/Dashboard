import { TestBed } from '@angular/core/testing';
import { of, throwError, noop } from 'rxjs';

import { AuthActionsService } from './auth-actions.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { UserSessionService } from '../user-session/user-session.service';
import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/auth/password-change.model';

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
    it('should be created with loading=false, error=null, passwordChangeResult=null', () => {
      expect(service).toBeTruthy();
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
      expect(service.passwordChangeResult()).toBeNull();
    });
  });

  describe('forgotPassword', () => {
    const email = 'user@example.com';

    it('should call forgotPassword with email, clear loading, and leave error null on success', () => {
      authApiClientMock.forgotPassword.mockReturnValue(of(undefined));
      service.forgotPassword(email).subscribe();

      expect(authApiClientMock.forgotPassword).toHaveBeenCalledWith(email);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous error before a new request', () => {
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500, message: 'fail' }) as ApiError),
      );
      service.forgotPassword(email).subscribe({ error: noop });
      expect(service.error()).toBe('fail');

      authApiClientMock.forgotPassword.mockReturnValue(of(undefined));
      service.forgotPassword(email).subscribe();
      expect(service.error()).toBeNull();
    });

    it.each([
      [{ status: 500, message: 'Email service down' } as ApiError, 'Email service down'],
      [{ status: 500 } as ApiError, 'Failed to send reset email'],
    ])('should set error on failure (message=%s)', (apiError, expected) => {
      authApiClientMock.forgotPassword.mockReturnValue(throwError(() => apiError));
      service.forgotPassword(email).subscribe({ error: noop });

      expect(service.error()).toBe(expected);
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

    it('should call requestPasswordChange with user id on success', () => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.requestPasswordChange.mockReturnValue(of(undefined));

      service.requestPasswordChange().subscribe();

      expect(authApiClientMock.requestPasswordChange).toHaveBeenCalledWith('user-42');
      expect(service.loading()).toBe(false);
    });

    it.each([
      ['null user', null],
      ['user with no id', { ...mockUser, id: '' }],
    ])('should emit 401 and not call API when %s', (_, user) => {
      userSessionMock.currentUser.mockReturnValue(user);

      service.requestPasswordChange().subscribe({
        error: (err: ApiError) => {
          expect(err.status).toBe(401);
        },
      });

      expect(authApiClientMock.requestPasswordChange).not.toHaveBeenCalled();
      expect(service.loading()).toBe(false);
    });

    it.each([
      [{ status: 500, message: 'Server error' } as ApiError, 'Server error'],
      [{ status: 500 } as ApiError, 'Failed to request password change'],
    ])('should set error on API failure (message=%s)', (apiError, expected) => {
      userSessionMock.currentUser.mockReturnValue(mockUser);
      authApiClientMock.requestPasswordChange.mockReturnValue(throwError(() => apiError));

      service.requestPasswordChange().subscribe({ error: noop });

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });
  });

  describe('confirmPasswordChange', () => {
    const mockData: PasswordChange = { token: 'reset-token', newPassword: 'newSecret456' };

    it('should call confirmPasswordChange, set result=true, and clear loading on success', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));
      service.confirmPasswordChange(mockData).subscribe();

      expect(authApiClientMock.confirmPasswordChange).toHaveBeenCalledWith(mockData);
      expect(service.passwordChangeResult()).toBe(true);
      expect(service.loading()).toBe(false);
    });

    it('should reset passwordChangeResult to null before request starts', () => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));
      service.confirmPasswordChange(mockData).subscribe();
      expect(service.passwordChangeResult()).toBe(true);

      // ESlint whines about explicit type any but it's the only way to capture the value during the request
      let resultDuringRequest: boolean | null = 'not-checked' as any;
      authApiClientMock.confirmPasswordChange.mockImplementation(() => {
        resultDuringRequest = service.passwordChangeResult();
        return of(undefined);
      });

      service.confirmPasswordChange(mockData).subscribe();
      expect(resultDuringRequest).toBeNull();
    });

    it.each([
      [{ status: 400, message: 'Invalid token' } as ApiError, 'Invalid token'],
      [{ status: 400 } as ApiError, 'Failed to change password'],
    ])('should set result=false and error on failure (message=%s)', (apiError, expected) => {
      authApiClientMock.confirmPasswordChange.mockReturnValue(throwError(() => apiError));
      service.confirmPasswordChange(mockData).subscribe({ error: noop });

      expect(service.passwordChangeResult()).toBe(false);
      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });
  });

  describe('clearMessages', () => {
    it('should clear both error and passwordChangeResult', () => {
      authApiClientMock.forgotPassword.mockReturnValue(
        throwError(() => ({ status: 500, message: 'some error' }) as ApiError),
      );
      service.forgotPassword('x').subscribe({ error: noop });
      expect(service.error()).toBe('some error');

      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));
      service.confirmPasswordChange({ token: 'x', newPassword: 'y' }).subscribe();
      expect(service.passwordChangeResult()).toBe(true);

      service.clearMessages();

      expect(service.error()).toBeNull();
      expect(service.passwordChangeResult()).toBeNull();
    });
  });
});
