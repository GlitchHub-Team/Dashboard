import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal } from '@angular/core';
import { By } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { of, EMPTY } from 'rxjs';

import { LoginPage } from './login.page';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { ForgotPasswordDialog } from './dialogs/forgot-password/forgot-password.dialog';
import { LoginRequest } from '../../models/auth/login-request.model';
import { AuthResponse } from '../../models/auth/auth-response.model';
import { UserRole } from '../../models/user/user-role.enum';

describe('LoginPage', () => {
  let component: LoginPage;
  let fixture: ComponentFixture<LoginPage>;
  // ESLint whining
  let loginFormDebug: any;

  const authSessionServiceMock = {
    login: vi.fn(),
    clearError: vi.fn(),
    loading: signal(false).asReadonly(),
    error: signal<string | null>(null).asReadonly(),
    isAuthenticated: signal(false).asReadonly(),
  };

  const routerMock = { navigate: vi.fn() };
  const dialogMock = { open: vi.fn() };

  const mockRequest: LoginRequest = {
    email: 'user@example.com',
    password: 'secret123',
    userRole: UserRole.SUPER_ADMIN,
  };
  const mockResponse: AuthResponse = {
    token: 'jwt-token',
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
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginPage);
    component = fixture.componentInstance;
    fixture.detectChanges();

    loginFormDebug = fixture.debugElement.query(By.css('app-login-form'));
  });

  it('should create and render shell, heading, and login form', () => {
    expect(component).toBeTruthy();
    expect(fixture.debugElement.query(By.css('.login-container'))).toBeTruthy();
    expect(fixture.debugElement.query(By.css('h1')).nativeElement.textContent).toContain('Accedi');
    expect(loginFormDebug).toBeTruthy();
  });

  describe('onLogin', () => {
    it('should call login and navigate to /dashboard on success', () => {
      authSessionServiceMock.login.mockReturnValue(of(mockResponse));

      loginFormDebug.triggerEventHandler('submitLogin', mockRequest);
      fixture.detectChanges();

      expect(authSessionServiceMock.login).toHaveBeenCalledWith(mockRequest);
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should call login but not navigate when observable completes without value', () => {
      authSessionServiceMock.login.mockReturnValue(EMPTY);

      loginFormDebug.triggerEventHandler('submitLogin', mockRequest);
      fixture.detectChanges();

      expect(authSessionServiceMock.login).toHaveBeenCalledWith(mockRequest);
      expect(routerMock.navigate).not.toHaveBeenCalled();
    });
  });

  describe('onForgotPassword', () => {
    it('should open ForgotPasswordDialog with correct config', () => {
      loginFormDebug.triggerEventHandler('forgotPassword');
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(ForgotPasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });

  describe('onDismissError', () => {
    it('should call clearError', () => {
      loginFormDebug.triggerEventHandler('dismissError');
      fixture.detectChanges();

      expect(authSessionServiceMock.clearError).toHaveBeenCalled();
    });
  });
});
