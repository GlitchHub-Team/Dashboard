import { TestBed } from '@angular/core/testing';
import { of, throwError, noop } from 'rxjs';

import { AuthActionsService } from './auth-actions.service';
import { AuthApiClientService } from '../auth-api-client/auth-api-client.service';
import { ApiError } from '../../models/api-error.model';
import { PasswordChange } from '../../models/auth/password-change.model';
import { ForgotPasswordRequest } from '../../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';

describe('AuthActionsService', () => {
  let service: AuthActionsService;

  const authApiClientMock = {
    forgotPasswordRequest: vi.fn(),
    confirmPasswordChange: vi.fn(),
    verifyForgotPasswordToken: vi.fn(),
    confirmPasswordReset: vi.fn(),
    verifyAccountToken: vi.fn(),
    confirmAccountCreation: vi.fn(),
  };

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [
        AuthActionsService,
        { provide: AuthApiClientService, useValue: authApiClientMock },
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
    const mockRequest: ForgotPasswordRequest = { email: 'user@example.com', tenantId: 'tenant-01' };

    it('should call forgotPasswordRequest with the request object, clear loading, and leave error null on success', () => {
      authApiClientMock.forgotPasswordRequest.mockReturnValue(of(undefined));
      service.forgotPassword(mockRequest).subscribe();

      expect(authApiClientMock.forgotPasswordRequest).toHaveBeenCalledWith(mockRequest);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous error before a new request', () => {
      authApiClientMock.forgotPasswordRequest.mockReturnValue(
        throwError(() => ({ status: 500, message: 'fail' }) as ApiError),
      );
      service.forgotPassword(mockRequest).subscribe({ error: noop });
      expect(service.error()).toBe('fail');

      authApiClientMock.forgotPasswordRequest.mockReturnValue(of(undefined));
      service.forgotPassword(mockRequest).subscribe();
      expect(service.error()).toBeNull();
    });

    it.each([
      [{ status: 500, message: 'Email service down' } as ApiError, 'Email service down'],
      [{ status: 500 } as ApiError, 'Failed to send reset email'],
    ])('should set error on failure (message=%s)', (apiError, expected) => {
      authApiClientMock.forgotPasswordRequest.mockReturnValue(throwError(() => apiError));
      service.forgotPassword(mockRequest).subscribe({ error: noop });

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });
  });

  describe('confirmPasswordChange', () => {
    const mockData: PasswordChange = { oldPassword: 'oldSecret123', newPassword: 'newSecret456' };

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

  describe('confirmPasswordReset', () => {
    const mockReq: ForgotPasswordResponse = { token: 'reset-token', newPassword: 'newSecret456' };

    it('should call confirmPasswordReset and clear loading on success', () => {
      authApiClientMock.verifyForgotPasswordToken.mockReturnValue(of(true));
      authApiClientMock.confirmPasswordReset.mockReturnValue(of(undefined));

      service.confirmPasswordReset(mockReq).subscribe();

      expect(authApiClientMock.confirmPasswordReset).toHaveBeenCalledWith(mockReq);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it('should clear previous error before a new request', () => {
      authApiClientMock.verifyForgotPasswordToken.mockReturnValue(of(true));
      authApiClientMock.confirmPasswordReset.mockReturnValue(
        throwError(() => ({ status: 500, message: 'fail' }) as ApiError),
      );
      service.confirmPasswordReset(mockReq).subscribe({ error: noop });
      expect(service.error()).toBe('fail');

      authApiClientMock.confirmPasswordReset.mockReturnValue(of(undefined));
      service.confirmPasswordReset(mockReq).subscribe();
      expect(service.error()).toBeNull();
    });

    it.each([
      [{ status: 400, message: 'Token expired' } as ApiError, 'Token expired'],
      [{ status: 400 } as ApiError, 'Failed to reset password'],
    ])('should set error on API failure (message=%s)', (apiError, expected) => {
      authApiClientMock.verifyForgotPasswordToken.mockReturnValue(of(true));
      authApiClientMock.confirmPasswordReset.mockReturnValue(throwError(() => apiError));

      service.confirmPasswordReset(mockReq).subscribe({ error: noop });

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });
  });

  describe('confirmAccount', () => {
    const mockReq: ConfirmAccountResponse = { token: 'account-token', newPassword: 'newSecret456' };

    it('should call confirmAccountCreation and clear loading on success', () => {
      authApiClientMock.verifyAccountToken.mockReturnValue(of(true));
      authApiClientMock.confirmAccountCreation.mockReturnValue(of(undefined));

      service.confirmAccount(mockReq).subscribe();

      expect(authApiClientMock.confirmAccountCreation).toHaveBeenCalledWith(mockReq);
      expect(service.loading()).toBe(false);
      expect(service.error()).toBeNull();
    });

    it.each([
      [
        { status: 400, message: 'Account already confirmed' } as ApiError,
        'Account already confirmed',
      ],
      [{ status: 400 } as ApiError, 'Failed to confirm account'],
    ])('should set error on API failure (message=%s)', (apiError, expected) => {
      authApiClientMock.verifyAccountToken.mockReturnValue(of(true));
      authApiClientMock.confirmAccountCreation.mockReturnValue(throwError(() => apiError));

      service.confirmAccount(mockReq).subscribe({ error: noop });

      expect(service.error()).toBe(expected);
      expect(service.loading()).toBe(false);
    });
  });

  describe('clearMessages', () => {
    it('should clear both error and passwordChangeResult', () => {
      authApiClientMock.forgotPasswordRequest.mockReturnValue(
        throwError(() => ({ status: 500, message: 'some error' }) as ApiError),
      );
      service.forgotPassword({ email: 'x@example.com' }).subscribe({ error: noop });
      expect(service.error()).toBe('some error');

      authApiClientMock.confirmPasswordChange.mockReturnValue(of(undefined));
      service.confirmPasswordChange({ oldPassword: 'old', newPassword: 'new' }).subscribe();
      expect(service.passwordChangeResult()).toBe(true);

      service.clearMessages();

      expect(service.error()).toBeNull();
      expect(service.passwordChangeResult()).toBeNull();
    });
  });
});
