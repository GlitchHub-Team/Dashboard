import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';

import { AuthApiClientService } from './auth-api-client.service';
import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { User } from '../../models/user/user.model';
import { UserRole } from '../../models/user/user-role.enum';
import { PasswordChange } from '../../models/auth/password-change.model';
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
      const mockRequest: LoginRequest = { email: 'test@example.com', password: 'password' };
      const mockUser: User = {
        id: '1',
        email: 'test@example.com',
        username: 'testuser',
        role: UserRole.TENANT_USER,
        tenantId: 'tenant-01',
      };
      const mockResponse: AuthResponse = { user: mockUser, token: 'mock-token' };

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
      const userId = '1';

      service.logout(userId).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/logout`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ userId });
      req.flush(null);
    });
  });

  describe('forgotPassword', () => {
    it('should POST email to /auth/forgot-password', () => {
      const email = 'test@example.com';

      service.forgotPassword(email).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/forgot-password`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ email });
      req.flush(null);
    });
  });

  describe('requestPasswordChange', () => {
    it('should POST userId to /auth/request-password-change', () => {
      const userId = '1';

      service.requestPasswordChange(userId).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/request-password-change`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ userId });
      req.flush(null);
    });
  });

  describe('confirmPasswordChange', () => {
    it('should POST data to /auth/confirm-password-change', () => {
      const data: PasswordChange = {
        token: 'reset-token',
        newPassword: 'new-password',
      };

      service.confirmPasswordChange(data).subscribe((response) => {
        expect(response).toBeNull();
      });

      const req = httpMock.expectOne(`${apiUrl}/confirm-password-change`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(data);
      req.flush(null);
    });
  });

  describe('error handling', () => {
    it('should propagate HTTP errors on login', () => {
      const mockRequest: LoginRequest = { email: 'test@example.com', password: 'password' };
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

    it('should propagate server errors', () => {
      const userId = '1';
      const mockError = { status: 500, statusText: 'Internal Server Error' };

      service.logout(userId).subscribe({
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
