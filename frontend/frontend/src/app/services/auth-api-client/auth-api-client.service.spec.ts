import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { AuthApiClientService } from './auth-api-client.service';
import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
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
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AuthApiClientService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('login', () => {
    it('should POST credentials to /auth/login and return AuthResponse', () => {
      const mockRequest: LoginRequest = {
        email: 'test@example.com',
        password: 'password',
        tenantId: 'tenant-01',
      };
      const mockResponse: AuthResponse = { jwt: 'mock-token' };

      service.login(mockRequest).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/login`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        email: 'test@example.com',
        password: 'password',
        tenant_id: 'tenant-01',
      });
      req.flush(mockResponse);
    });
  });

  describe('logout', () => {
    it('should POST empty body to /auth/logout', () => {
      service.logout().subscribe();

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({});
      req.flush(null);
    });
  });

  describe('verifyForgotPasswordToken', () => {
    it('should POST token in body to /auth/forgot_password/verify_token', () => {
      service.verifyForgotPasswordToken('reset-token').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/verify_token`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ token: 'reset-token', tenant_id: undefined });
      req.flush(null);
    });

    it('should include tenant_id in body when tenantId is provided', () => {
      service.verifyForgotPasswordToken('reset-token', 'tenant-01').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/verify_token`);
      expect(req.request.body).toEqual({ token: 'reset-token', tenant_id: 'tenant-01' });
      req.flush(null);
    });
  });

  describe('forgotPasswordRequest', () => {
    it('should POST email and tenantId to /auth/forgot_password/request', () => {
      const mockRequest: ForgotPasswordRequest = {
        email: 'test@example.com',
        tenantId: 'tenant-01',
      };

      service.forgotPasswordRequest(mockRequest).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/request`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ email: 'test@example.com', tenant_id: 'tenant-01' });
      req.flush(null);
    });

    it('should POST with tenant_id as undefined when tenantId is omitted', () => {
      const mockRequest: ForgotPasswordRequest = { email: 'test@example.com' };

      service.forgotPasswordRequest(mockRequest).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/request`);
      expect(req.request.body).toEqual({ email: 'test@example.com', tenant_id: undefined });
      req.flush(null);
    });
  });

  describe('confirmPasswordReset', () => {
    it('should POST token, tenant_id, and new_password to /auth/forgot_password', () => {
      const mockRequest: ForgotPasswordResponse = {
        token: 'reset-token',
        tenantId: 'tenant-01',
        newPassword: 'new-password',
      };

      service.confirmPasswordReset(mockRequest).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        token: 'reset-token',
        tenant_id: 'tenant-01',
        new_password: 'new-password',
      });
      req.flush(null);
    });

    it('should POST with tenant_id as undefined when omitted', () => {
      const mockRequest: ForgotPasswordResponse = {
        token: 'reset-token',
        newPassword: 'new-password',
      };

      service.confirmPasswordReset(mockRequest).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password`);
      expect(req.request.body).toEqual({
        token: 'reset-token',
        tenant_id: undefined,
        new_password: 'new-password',
      });
      req.flush(null);
    });
  });

  describe('confirmPasswordChange', () => {
    it('should POST old_password and new_password to /auth/change_password', () => {
      const data: PasswordChange = { oldPassword: 'old-password', newPassword: 'new-password' };

      service.confirmPasswordChange(data).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/change_password`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        old_password: 'old-password',
        new_password: 'new-password',
      });
      req.flush(null);
    });
  });

  describe('verifyAccountToken', () => {
    it('should POST token in body to /auth/confirm_account/verify_token/', () => {
      service.verifyAccountToken('account-token').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/confirm_account/verify_token/`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ token: 'account-token', tenant_id: undefined });
      req.flush(null);
    });

    it('should include tenant_id in body when tenantId is provided', () => {
      service.verifyAccountToken('account-token', 'tenant-01').subscribe();

      const req = httpMock.expectOne(`${apiUrl}/confirm_account/verify_token/`);
      expect(req.request.body).toEqual({ token: 'account-token', tenant_id: 'tenant-01' });
      req.flush(null);
    });
  });

  describe('confirmAccountCreation', () => {
    it('should POST token, tenant_id, and new_password to /auth/confirm_account and return AuthResponse', () => {
      const mockRequest: ConfirmAccountResponse = {
        token: 'account-token',
        tenantId: 'tenant-01',
        newPassword: 'new-password',
      };
      const mockResponse: AuthResponse = { jwt: 'mock-token' };

      service.confirmAccountCreation(mockRequest).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/confirm_account`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        token: 'account-token',
        tenant_id: 'tenant-01',
        new_password: 'new-password',
      });
      req.flush(mockResponse);
    });

    it('should POST with tenant_id as undefined when omitted', () => {
      const mockRequest: ConfirmAccountResponse = {
        token: 'account-token',
        newPassword: 'new-password',
      };

      service.confirmAccountCreation(mockRequest).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/confirm_account`);
      expect(req.request.body).toEqual({
        token: 'account-token',
        tenant_id: undefined,
        new_password: 'new-password',
      });
      req.flush({ jwt: 'mock-token' });
    });
  });

  describe('error handling', () => {
    it('should propagate HTTP errors on login', () => {
      const mockRequest: LoginRequest = {
        email: 'test@example.com',
        password: 'password',
        tenantId: 'tenant-01',
      };

      service.login(mockRequest).subscribe({
        next: () => expect.unreachable('expected an error'),
        error: (error) => {
          expect(error.status).toBe(401);
          expect(error.statusText).toBe('Unauthorized');
        },
      });

      const req = httpMock.expectOne(`${apiUrl}/login`);
      req.flush(null, { status: 401, statusText: 'Unauthorized' });
    });

    it('should propagate server errors on logout', () => {
      service.logout().subscribe({
        next: () => expect.unreachable('expected an error'),
        error: (error) => {
          expect(error.status).toBe(500);
          expect(error.statusText).toBe('Internal Server Error');
        },
      });

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      req.flush(null, { status: 500, statusText: 'Internal Server Error' });
    });
  });
});
