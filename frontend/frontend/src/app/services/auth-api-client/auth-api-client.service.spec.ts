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
    it.each([
      [undefined, { token: 'reset-token', tenant_id: undefined }],
      ['tenant-01', { token: 'reset-token', tenant_id: 'tenant-01' }],
    ])('should POST to /auth/forgot_password/verify_token (tenantId=%s)', (tenantId, expectedBody) => {
      service.verifyForgotPasswordToken('reset-token', tenantId as string | undefined).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/verify_token`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(expectedBody);
      req.flush(null);
    });
  });

  describe('forgotPasswordRequest', () => {
    it.each<[ForgotPasswordRequest, object]>([
      [{ email: 'test@example.com', tenantId: 'tenant-01' }, { email: 'test@example.com', tenant_id: 'tenant-01' }],
      [{ email: 'test@example.com' }, { email: 'test@example.com', tenant_id: undefined }],
    ])('should POST to /auth/forgot_password/request (tenantId=%s)', (request, expectedBody) => {
      service.forgotPasswordRequest(request).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password/request`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(expectedBody);
      req.flush(null);
    });
  });

  describe('confirmPasswordReset', () => {
    it.each<[ForgotPasswordResponse, object]>([
      [{ token: 'reset-token', tenantId: 'tenant-01', newPassword: 'new-password' }, { token: 'reset-token', tenant_id: 'tenant-01', new_password: 'new-password' }],
      [{ token: 'reset-token', newPassword: 'new-password' }, { token: 'reset-token', tenant_id: undefined, new_password: 'new-password' }],
    ])('should POST to /auth/forgot_password (tenantId=%s)', (request, expectedBody) => {
      service.confirmPasswordReset(request).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/forgot_password`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(expectedBody);
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
    it.each([
      [undefined, { token: 'account-token', tenant_id: undefined }],
      ['tenant-01', { token: 'account-token', tenant_id: 'tenant-01' }],
    ])('should POST to /auth/confirm_account/verify_token/ (tenantId=%s)', (tenantId, expectedBody) => {
      service.verifyAccountToken('account-token', tenantId as string | undefined).subscribe();

      const req = httpMock.expectOne(`${apiUrl}/confirm_account/verify_token/`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(expectedBody);
      req.flush(null);
    });
  });

  describe('confirmAccountCreation', () => {
    const mockResponse: AuthResponse = { jwt: 'mock-token' };

    it.each<[ConfirmAccountResponse, object]>([
      [{ token: 'account-token', tenantId: 'tenant-01', newPassword: 'new-password' }, { token: 'account-token', tenant_id: 'tenant-01', new_password: 'new-password' }],
      [{ token: 'account-token', newPassword: 'new-password' }, { token: 'account-token', tenant_id: undefined, new_password: 'new-password' }],
    ])('should POST to /auth/confirm_account and return AuthResponse (tenantId=%s)', (request, expectedBody) => {
      service.confirmAccountCreation(request).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      const req = httpMock.expectOne(`${apiUrl}/confirm_account`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(expectedBody);
      req.flush(mockResponse);
    });
  });

  describe('error handling', () => {
    it.each([
      ['login', () => service.login({ email: 'test@example.com', password: 'password', tenantId: 'tenant-01' }), `${apiUrl}/login`, 401, 'Unauthorized'],
      ['logout', () => service.logout(), `${apiUrl}/logout`, 500, 'Internal Server Error'],
    ])('should propagate HTTP errors from %s', (_name, call, url, status, statusText) => {
      (call() as ReturnType<typeof service.logout>).subscribe({
        next: () => expect.unreachable('expected an error'),
        error: (error) => {
          expect(error.status).toBe(status);
          expect(error.statusText).toBe(statusText);
        },
      });

      const req = httpMock.expectOne(url);
      req.flush(null, { status, statusText });
    });
  });
});
