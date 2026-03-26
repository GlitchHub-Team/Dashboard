import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { AuthApiClientService } from './auth-api-client.service';
import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { UserRole } from '../../models/user/user-role.enum';
import { PasswordChange } from '../../models/auth/password-change.model';
import { ForgotPasswordRequest } from '../../models/auth/forgot-password-request.model';
import { ForgotPasswordResponse } from '../../models/auth/forgot-password.model';
import { ConfirmAccountResponse } from '../../models/auth/confirm-account.model';
import { environment } from '../../../environments/environment';

describe('AuthApiClientService', () => {
  let service: AuthApiClientService;
  let httpMock: HttpTestingController;
  const apiUrl = `${environment.apiUrl}/auth`;

  beforeEach(() => {
    vi.resetAllMocks();

    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AuthApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  // Fail se una richiesta viene fatta ma non gestita
  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('login', () => {
    it('should POST credentials to /auth/login', () => {
      const mockRequest: LoginRequest = {
        email: 'test@example.com',
        password: 'password',
        userRole: UserRole.TENANT_USER,
        tenantId: 'tenant-01',
      };
      const mockResponse: AuthResponse = { token: 'mock-token' };

      service.login(mockRequest).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/login`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockRequest);
      req.flush(mockResponse);
    });
  });

  describe('logout', () => {
    it('should POST userId to /auth/logout', () => {
      service.logout().subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });
  });

  describe('verifyForgotPasswordToken', () => {
    it('should GET to /auth/forgot_password/verify_token/:token and return true when result is true', () => {
      const token = 'reset-token';

      service.verifyForgotPasswordToken(token).subscribe((response) => {
        expect(response).toBe(true);
      });

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/verify_token/${token}`);
      expect(req.request.method).toBe('GET');
      req.flush({ result: true });
    });

    it('should return false when result is false', () => {
      const token = 'reset-token';

      service.verifyForgotPasswordToken(token).subscribe((response) => {
        expect(response).toBe(false);
      });

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/verify_token/${token}`);
      req.flush({ result: false });
    });
  });

  describe('forgotPasswordRequest', () => {
    it('should POST email and tenantId to /auth/forgot_password/request', () => {
      const mockRequest: ForgotPasswordRequest = {
        email: 'test@example.com',
        tenantId: 'tenant-01',
      };

      service.forgotPasswordRequest(mockRequest).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/request`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockRequest);
      req.flush(null);
    });

    it('should POST with only email when tenantId is omitted', () => {
      const mockRequest: ForgotPasswordRequest = { email: 'test@example.com' };

      service.forgotPasswordRequest(mockRequest).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/request`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockRequest);
      req.flush(null);
    });
  });

  describe('confirmPasswordReset', () => {
    it('should POST data to /auth/forgot_password', () => {
      const mockRequest: ForgotPasswordResponse = {
        token: 'reset-token',
        newPassword: 'new-password',
      };

      service.confirmPasswordReset(mockRequest).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/forgot_password`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockRequest);
      req.flush(null);
    });
  });

  describe('confirmPasswordChange', () => {
    it('should POST data to /auth/change_password', () => {
      const data: PasswordChange = { oldPassword: 'old-password', newPassword: 'new-password' };

      service.confirmPasswordChange(data).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/change_password`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(data);
      req.flush(null);
    });
  });

  describe('verifyAccountToken', () => {
    it('should GET to /auth/confirm_account/verify_token/:token and return true when result is true', () => {
      const token = 'account-token';

      service.verifyAccountToken(token).subscribe((response) => {
        expect(response).toBe(true);
      });

      const req = httpMock.expectOne(`${apiUrl}/confirm_account/verify_token/${token}`);
      expect(req.request.method).toBe('GET');
      req.flush({ result: true });
    });

    it('should return false when result is false', () => {
      const token = 'account-token';

      service.verifyAccountToken(token).subscribe((response) => {
        expect(response).toBe(false);
      });

      const req = httpMock.expectOne(`${apiUrl}/confirm_account/verify_token/${token}`);
      req.flush({ result: false });
    });
  });

  describe('confirmAccountCreation', () => {
    it('should POST data to /auth/confirm_account', () => {
      const mockRequest: ConfirmAccountResponse = {
        token: 'account-token',
        newPassword: 'new-password',
      };

      service.confirmAccountCreation(mockRequest).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/confirm_account`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockRequest);
      req.flush(null);
    });
  });

  describe('error handling', () => {
    it('should propagate HTTP errors on login', () => {
      const mockRequest: LoginRequest = {
        email: 'test@example.com',
        password: 'password',
        userRole: UserRole.TENANT_USER,
        tenantId: 'tenant-01',
      };
      const mockError = { status: 401, statusText: 'Unauthorized' };

      service.login(mockRequest).subscribe({
        next: () => expect.unreachable('expected an error'),
        error: (error) => {
          expect(error.status).toBe(401);
          expect(error.statusText).toBe('Unauthorized');
        },
      });

      const req = httpMock.expectOne(`${apiUrl}/login`);
      expect(req.request.method).toBe('POST');
      req.flush(null, mockError);
    });

    it('should propagate server errors on logout', () => {
      const mockError = { status: 500, statusText: 'Internal Server Error' };

      service.logout().subscribe({
        next: () => expect.unreachable('expected an error'),
        error: (error) => {
          expect(error.status).toBe(500);
          expect(error.statusText).toBe('Internal Server Error');
        },
      });

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      expect(req.request.method).toBe('POST');
      req.flush(null, mockError);
    });
  });
});
