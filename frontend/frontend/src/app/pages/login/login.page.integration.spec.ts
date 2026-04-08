import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { LoginPage } from './login.page';
import { LoginFormComponent } from './components/login-form/login-form.component';
import { ForgotPasswordDialog } from './dialogs/forgot-password/forgot-password.dialog';
import { AuthSessionService } from '../../services/auth/auth-session.service';
import { TenantService } from '../../services/tenant/tenant.service';
import { UserRole } from '../../models/user/user-role.enum';
import { Tenant } from '../../models/tenant/tenant.model';

const mockTenants: Tenant[] = [
  { id: 'tenant-01', name: 'Tenant One', canImpersonate: false },
  { id: 'tenant-02', name: 'Tenant Two', canImpersonate: false },
];

function createAuthSessionServiceMock() {
  return {
    login: vi.fn().mockReturnValue(of({ jwt: 'jwt-token' })),
    clearError: vi.fn(),
    loading: signal(false),
    error: signal<string | null>(null),
  };
}

function setupTestBed() {
  const authMock = createAuthSessionServiceMock();
  const routerMock = { navigate: vi.fn() };
  const dialogMock = { open: vi.fn() };
  const tenantServiceMock = {
    retrieveTenants: vi.fn(),
    tenantList: signal<Tenant[]>(mockTenants),
  };

  TestBed.configureTestingModule({
    imports: [LoginPage, LoginFormComponent],
    providers: [
      { provide: AuthSessionService, useValue: authMock },
      { provide: Router, useValue: routerMock },
      { provide: TenantService, useValue: tenantServiceMock },
    ],
  }).overrideProvider(MatDialog, { useValue: dialogMock });

  const fixture = TestBed.createComponent(LoginPage);
  return { fixture, authMock, routerMock, dialogMock, tenantServiceMock };
}

function getForm(fixture: ComponentFixture<LoginPage>) {
  return fixture.debugElement.query(By.directive(LoginFormComponent));
}

function getLoginForm(fixture: ComponentFixture<LoginPage>) {
  return (getForm(fixture).componentInstance as LoginFormComponent)['loginForm'];
}

function fillAndSubmitForm(
  fixture: ComponentFixture<LoginPage>,
  opts: { email: string; password: string; role: UserRole; tenantId?: string },
): void {
  const form = getLoginForm(fixture);
  form.controls.email.setValue(opts.email);
  form.controls.password.setValue(opts.password);
  form.controls.role.setValue(opts.role);
  if (opts.tenantId) form.controls.tenantId.setValue(opts.tenantId);
  fixture.detectChanges();
  fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
  fixture.detectChanges();
}

function submitEmpty(fixture: ComponentFixture<LoginPage>): void {
  fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
  fixture.detectChanges();
}

describe('LoginPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should render container, heading, form with fields, and no progress/error', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.login-container'))).toBeTruthy();
      expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Accedi');
      expect(getForm(fixture)).toBeTruthy();
      expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('input[type="email"]')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('input[type="password"]')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('button[type="submit"]')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('mat-progress-bar')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.error-banner')).toBeFalsy();
    });

    it('should call retrieveTenants on init', () => {
      const { fixture, tenantServiceMock } = setupTestBed();
      fixture.detectChanges();

      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalled();
    });
  });

  describe('Page -> Form: Input Bindings', () => {
    it('should show progress bar, spin icon, and disable buttons when loading', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.loading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-progress-bar')).toBeTruthy();
      expect(fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin'))).toBeTruthy();
      expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBe(true);
      expect(fixture.nativeElement.querySelector('.forgot-password').disabled).toBe(true);
    });

    it('should show error banner when error is set', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.error as WritableSignal<string | null>).set('Credenziali non valide');
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain('Credenziali non valide');
    });
  });

  describe('Role and Tenant Dropdown', () => {
    it('should not show tenant dropdown initially', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      const formFields = fixture.nativeElement.querySelectorAll('mat-form-field');
      expect(formFields.length).toBe(3);
    });

    it('should show tenant dropdown when TENANT_ADMIN role is selected', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      getLoginForm(fixture).controls.role.setValue(UserRole.TENANT_ADMIN);
      fixture.detectChanges();

      const formFields = fixture.nativeElement.querySelectorAll('mat-form-field');
      expect(formFields.length).toBe(4);
    });

    it('should hide tenant dropdown when switching from TENANT_ADMIN to SUPER_ADMIN', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      const form = getLoginForm(fixture);
      form.controls.role.setValue(UserRole.TENANT_ADMIN);
      fixture.detectChanges();
      expect(fixture.nativeElement.querySelectorAll('mat-form-field').length).toBe(4);

      form.controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
      expect(fixture.nativeElement.querySelectorAll('mat-form-field').length).toBe(3);
    });
  });

  describe('Form -> Page: Login Flow', () => {
    it('should call login and navigate to /dashboard for Super Admin', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      fixture.detectChanges();
      fillAndSubmitForm(fixture, {
        email: 'admin@example.com',
        password: 'secret123',
        role: UserRole.SUPER_ADMIN,
      });

      expect(authMock.login).toHaveBeenCalledWith({
        email: 'admin@example.com',
        password: 'secret123',
        tenantId: undefined,
      });
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should call login with tenantId for Tenant Admin', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      fixture.detectChanges();
      fillAndSubmitForm(fixture, {
        email: 'user@example.com',
        password: 'secret123',
        role: UserRole.TENANT_ADMIN,
        tenantId: 'tenant-01',
      });

      expect(authMock.login).toHaveBeenCalledWith({
        email: 'user@example.com',
        password: 'secret123',
        tenantId: 'tenant-01',
      });
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should not navigate when login returns EMPTY', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      authMock.login.mockReturnValue(EMPTY);
      fixture.detectChanges();
      fillAndSubmitForm(fixture, {
        email: 'admin@example.com',
        password: 'secret123',
        role: UserRole.SUPER_ADMIN,
      });

      expect(authMock.login).toHaveBeenCalled();
      expect(routerMock.navigate).not.toHaveBeenCalled();
    });

    it('should not call login when form is empty', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      submitEmpty(fixture);

      expect(authMock.login).not.toHaveBeenCalled();
    });

    it('should not call login with invalid email', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      fillAndSubmitForm(fixture, {
        email: 'not-an-email',
        password: 'secret123',
        role: UserRole.SUPER_ADMIN,
      });

      expect(authMock.login).not.toHaveBeenCalled();
    });

    it('should not call login when Tenant Admin has no tenantId', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      const form = getLoginForm(fixture);
      form.controls.email.setValue('user@example.com');
      form.controls.password.setValue('secret123');
      form.controls.role.setValue(UserRole.TENANT_ADMIN);
      fixture.detectChanges();
      submitEmpty(fixture);

      expect(authMock.login).not.toHaveBeenCalled();
    });
  });

  describe('Form Validation Errors in Template', () => {
    it('should show required errors on empty submit', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      submitEmpty(fixture);

      const errorTexts = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errorTexts.some((t: string) => t.includes('Campo obbligatorio'))).toBe(true);
    });

    it('should show invalid email error', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      const form = getLoginForm(fixture);
      form.controls.email.setValue('not-an-email');
      form.controls.email.markAsTouched();
      fixture.detectChanges();

      const errorTexts = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errorTexts.some((t: string) => t.includes('Indirizzo email non valido'))).toBe(true);
    });
  });

  describe('Error Banner Dismiss', () => {
    it('should call clearError when close button is clicked', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.error as WritableSignal<string | null>).set('Invalid credentials');
      fixture.detectChanges();

      fixture.debugElement.query(By.css('.error-banner button')).triggerEventHandler('click');
      fixture.detectChanges();

      expect(authMock.clearError).toHaveBeenCalled();
    });
  });

  describe('Forgot Password', () => {
    it('should open ForgotPasswordDialog when forgot password button is clicked', () => {
      const { fixture, dialogMock } = setupTestBed();
      fixture.detectChanges();

      fixture.debugElement.query(By.css('.forgot-password')).triggerEventHandler('click');
      fixture.detectChanges();

      expect(dialogMock.open).toHaveBeenCalledWith(ForgotPasswordDialog, {
        width: '800px',
        disableClose: true,
      });
    });
  });

  describe('Full Login Flow', () => {
    it('should submit, call login, and navigate to dashboard', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      fixture.detectChanges();

      fillAndSubmitForm(fixture, {
        email: 'admin@example.com',
        password: 'secret123',
        role: UserRole.SUPER_ADMIN,
      });

      expect(authMock.login).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });
  });
});
