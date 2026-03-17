import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { of, EMPTY } from 'rxjs';

import { LoginPage } from './login.page';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { ForgotPasswordDialog } from './dialogs/forgot-password/forgot-password.dialog';
import { LoginRequest } from '../../models/login-request.model';
import { AuthResponse } from '../../models/auth-response.model';
import { UserRole } from '../../models/user-role.enum';

describe('LoginPage', () => {
  let component: LoginPage;
  let fixture: ComponentFixture<LoginPage>;

  const authSessionServiceMock = {
    login: vi.fn(),
    clearError: vi.fn(),
    loading: signal(false).asReadonly(),
    error: signal<string | null>(null).asReadonly(),
    isAuthenticated: signal(false).asReadonly(),
  };

  const routerMock = {
    navigate: vi.fn(),
  };

  const dialogMock = {
    open: vi.fn(),
  };

  beforeEach(async () => {
    vi.resetAllMocks();

    await TestBed.configureTestingModule({
      imports: [LoginPage],
      providers: [
        { provide: AuthSessionService, useValue: authSessionServiceMock },
        { provide: Router, useValue: routerMock },
        { provide: MatDialog, useValue: dialogMock },
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('onLogin', () => {
    const mockRequest: LoginRequest = {
      email: 'user@example.com',
      password: 'secret123',
    };

    const mockResponse: AuthResponse = {
      token: 'jwt-token',
      user: {
        id: '1',
        email: 'user@example.com',
        role: UserRole.SUPER_ADMIN,
        tenantId: 'tenant-1',
      },
    };

    it('should call authSessionService.login with the request', () => {
      authSessionServiceMock.login.mockReturnValue(of(mockResponse));

      component['onLogin'](mockRequest);

      expect(authSessionServiceMock.login).toHaveBeenCalledWith(mockRequest);
    });

    it('should navigate to /dashboard on success', () => {
      authSessionServiceMock.login.mockReturnValue(of(mockResponse));

      component['onLogin'](mockRequest);

      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should not navigate on error', () => {
      // The real service catches the error and returns EMPTY
      authSessionServiceMock.login.mockReturnValue(EMPTY);

      component['onLogin'](mockRequest);

      expect(routerMock.navigate).not.toHaveBeenCalled();
    });

    it('should call login even if it errors', () => {
      authSessionServiceMock.login.mockReturnValue(EMPTY);

      component['onLogin'](mockRequest);

      expect(authSessionServiceMock.login).toHaveBeenCalledWith(mockRequest);
    });
  });

  describe('onForgotPassword', () => {
    it('should open ForgotPasswordDialog', () => {
      component['onForgotPassword']();

      expect(dialogMock.open).toHaveBeenCalledWith(ForgotPasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });

  describe('onDismissError', () => {
    it('should call clearError', () => {
      component['onDismissError']();

      expect(authSessionServiceMock.clearError).toHaveBeenCalled();
    });
  });
});
