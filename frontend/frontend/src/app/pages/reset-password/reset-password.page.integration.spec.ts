import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { By } from '@angular/platform-browser';
import { of } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { ResetPasswordPage } from './reset-password.page';
import { ResetPasswordFormComponent } from './components/reset-password-form/reset-password-form.component';
import { AuthActionsService } from '../../services/auth/auth-actions.service';

function createAuthActionsServiceMock() {
  return {
    confirmPasswordReset: vi.fn().mockReturnValue(of(undefined)),
    clearMessages: vi.fn(),
    loading: signal(false),
    error: signal<string | null>(null),
    passwordChangeResult: signal<boolean | null>(null),
  };
}

function setupTestBed(options?: { token?: string; tenantId?: string | null }) {
  const token = options?.token ?? 'reset-token';
  const tenantId = options?.tenantId ?? null;
  const authMock = createAuthActionsServiceMock();
  const routerMock = { navigate: vi.fn() };
  const activatedRouteMock = {
    snapshot: {
      paramMap: { get: vi.fn().mockImplementation((key: string) => (key === 'token' ? token : null)) },
      queryParamMap: { get: vi.fn().mockImplementation((key: string) => (key === 'tid' ? tenantId : null)) },
    },
  };

  TestBed.configureTestingModule({
    imports: [ResetPasswordPage, ResetPasswordFormComponent],
    providers: [
      { provide: AuthActionsService, useValue: authMock },
      { provide: Router, useValue: routerMock },
      { provide: ActivatedRoute, useValue: activatedRouteMock },
    ],
  });

  const fixture = TestBed.createComponent(ResetPasswordPage);
  return { fixture, authMock, routerMock };
}

function getForm(fixture: ComponentFixture<ResetPasswordPage>) {
  return fixture.debugElement.query(By.directive(ResetPasswordFormComponent));
}

function fillAndSubmitForm(fixture: ComponentFixture<ResetPasswordPage>, password: string): void {
  const form = getForm(fixture).componentInstance as ResetPasswordFormComponent;
  form['resetPasswordForm'].controls.newPassword.setValue(password);
  form['resetPasswordForm'].controls.confirmNewPassword.setValue(password);
  fixture.detectChanges();
  fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
  fixture.detectChanges();
}

describe('ResetPasswordPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should render container, heading, and form', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.reset-container'))).toBeTruthy();
      expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Reimposta Password');
      expect(getForm(fixture)).toBeTruthy();
    });

    it('should render password fields and submit button', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
      expect(fixture.nativeElement.querySelectorAll('input[type="password"]').length).toBe(2);
      expect(fixture.nativeElement.querySelector('button[type="submit"]')).toBeTruthy();
    });

    it('should not show progress bar, success banner, or error banner', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-progress-bar')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.success-banner')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.error-banner')).toBeFalsy();
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
      expect(fixture.nativeElement.querySelector('button[type="button"]').disabled).toBe(true);
    });

    it('should show error banner when error is set', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.error as WritableSignal<string | null>).set('Token non valido');
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain('Token non valido');
    });

    it('should show success banner and hide form when passwordChangeResult is true', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.passwordChangeResult as WritableSignal<boolean | null>).set(true);
      fixture.detectChanges();

      const successBanner = fixture.nativeElement.querySelector('.success-banner');
      expect(successBanner).toBeTruthy();
      expect(successBanner.textContent).toContain('Reimpostazione della password riuscita');
      expect(fixture.nativeElement.querySelector('form')).toBeFalsy();
    });

    it.each([false, null])('should not show success banner when passwordChangeResult is %s', (value) => {
      const { fixture, authMock } = setupTestBed();
      (authMock.passwordChangeResult as WritableSignal<boolean | null>).set(value);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.success-banner')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
    });
  });

  describe('Form -> Page: Submit Reset Flow', () => {
    it('should call confirmPasswordReset with token and undefined tenantId', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmPasswordReset).toHaveBeenCalledWith({
        token: 'reset-token',
        tenantId: undefined,
        newPassword: 'newSecret123',
      });
    });

    it('should include tenantId from query params when present', () => {
      const { fixture, authMock } = setupTestBed({ tenantId: 'tenant-01' });
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmPasswordReset).toHaveBeenCalledWith({
        token: 'reset-token',
        tenantId: 'tenant-01',
        newPassword: 'newSecret123',
      });
    });

    it('should use the token from route params', () => {
      const { fixture, authMock } = setupTestBed({ token: 'custom-token' });
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmPasswordReset).toHaveBeenCalledWith(
        expect.objectContaining({ token: 'custom-token' }),
      );
    });

    it('should not call confirmPasswordReset when form is empty', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(authMock.confirmPasswordReset).not.toHaveBeenCalled();
    });

    it('should not call confirmPasswordReset when passwords do not match', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      const form = getForm(fixture).componentInstance as ResetPasswordFormComponent;
      form['resetPasswordForm'].controls.newPassword.setValue('password1');
      form['resetPasswordForm'].controls.confirmNewPassword.setValue('password2');
      fixture.detectChanges();
      fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(authMock.confirmPasswordReset).not.toHaveBeenCalled();
    });
  });

  describe('Form Validation Errors in Template', () => {
    it('should show required errors on empty submit', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      const errorTexts = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errorTexts.some((t: string) => t.includes('Campo obbligatorio'))).toBe(true);
    });

    it('should show mismatch error when passwords differ and confirm is dirty', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      const form = getForm(fixture).componentInstance as ResetPasswordFormComponent;
      form['resetPasswordForm'].controls.newPassword.setValue('secret123');
      form['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      form['resetPasswordForm'].controls.confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      const mismatchError = fixture.nativeElement.querySelector('.field-error');
      expect(mismatchError).toBeTruthy();
      expect(mismatchError.textContent).toContain('Le password non coincidono');
    });

    it('should hide mismatch error when passwords match', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      const form = getForm(fixture).componentInstance as ResetPasswordFormComponent;
      form['resetPasswordForm'].controls.newPassword.setValue('secret123');
      form['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      form['resetPasswordForm'].controls.confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.field-error')).toBeFalsy();
    });
  });

  describe('Error Banner Dismiss', () => {
    it('should call clearMessages when close button is clicked', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.error as WritableSignal<string | null>).set('Something went wrong');
      fixture.detectChanges();

      fixture.debugElement.query(By.css('.error-banner button')).triggerEventHandler('click');
      fixture.detectChanges();

      expect(authMock.clearMessages).toHaveBeenCalled();
    });
  });

  describe('Navigation: Go to Login', () => {
    it('should call clearMessages and navigate to /login from form back button', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      fixture.detectChanges();

      const backButton = fixture.debugElement
        .queryAll(By.css('button[type="button"]'))
        .find((b) => b.nativeElement.textContent.includes('Torna indietro'));
      backButton!.triggerEventHandler('click');
      fixture.detectChanges();

      expect(authMock.clearMessages).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });

    it('should call clearMessages and navigate to /login from success back button', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      (authMock.passwordChangeResult as WritableSignal<boolean | null>).set(true);
      fixture.detectChanges();

      const backButton = fixture.debugElement
        .queryAll(By.css('button'))
        .find((b) => b.nativeElement.textContent.includes('Torna indietro'));
      backButton!.triggerEventHandler('click');
      fixture.detectChanges();

      expect(authMock.clearMessages).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });
  });

  describe('Full Reset Flow', () => {
    it('should submit, show success banner, then navigate to login', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      fixture.detectChanges();

      fillAndSubmitForm(fixture, 'newSecret123');
      expect(authMock.confirmPasswordReset).toHaveBeenCalled();

      (authMock.passwordChangeResult as WritableSignal<boolean | null>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.success-banner')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('form')).toBeFalsy();

      const backButton = fixture.debugElement
        .queryAll(By.css('button'))
        .find((b) => b.nativeElement.textContent.includes('Torna indietro'));
      backButton!.triggerEventHandler('click');
      fixture.detectChanges();

      expect(authMock.clearMessages).toHaveBeenCalled();
      expect(routerMock.navigate).toHaveBeenCalledWith(['/login']);
    });
  });
});
