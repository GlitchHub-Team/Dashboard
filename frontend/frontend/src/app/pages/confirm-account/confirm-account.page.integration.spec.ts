import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { ConfirmAccountPage } from './confirm-account.page';
import { ConfirmAccountFormComponent } from './components/confirm-account-form/confirm-account-form.component';
import { AuthActionsService } from '../../services/auth/auth-actions.service';

function createAuthActionsServiceMock() {
  return {
    confirmAccount: vi.fn().mockReturnValue(of(undefined)),
    clearMessages: vi.fn(),
    loading: signal(false),
    error: signal<string | null>(null),
  };
}

function setupTestBed(options?: { token?: string; tenantId?: string | null }) {
  const token = options?.token ?? 'confirm-token';
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
    imports: [ConfirmAccountPage, ConfirmAccountFormComponent],
    providers: [
      { provide: AuthActionsService, useValue: authMock },
      { provide: Router, useValue: routerMock },
      { provide: ActivatedRoute, useValue: activatedRouteMock },
    ],
  });

  const fixture = TestBed.createComponent(ConfirmAccountPage);
  return { fixture, authMock, routerMock };
}

function getForm(fixture: ComponentFixture<ConfirmAccountPage>) {
  return fixture.debugElement.query(By.directive(ConfirmAccountFormComponent));
}

function getConfirmForm(fixture: ComponentFixture<ConfirmAccountPage>) {
  return (getForm(fixture).componentInstance as ConfirmAccountFormComponent)['confirmAccountForm'];
}

function fillAndSubmitForm(fixture: ComponentFixture<ConfirmAccountPage>, password: string): void {
  const form = getConfirmForm(fixture);
  form.controls.newPassword.setValue(password);
  form.controls.confirmNewPassword.setValue(password);
  fixture.detectChanges();
  fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
  fixture.detectChanges();
}

function submitEmpty(fixture: ComponentFixture<ConfirmAccountPage>): void {
  fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
  fixture.detectChanges();
}

describe('ConfirmAccountPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should render container, heading, form with password fields, and no progress/error', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.confirm-account-container'))).toBeTruthy();
      expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Conferma Account');
      expect(getForm(fixture)).toBeTruthy();
      expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
      expect(fixture.nativeElement.querySelectorAll('input[type="password"]').length).toBe(2);
      expect(fixture.nativeElement.querySelector('button[type="submit"]')).toBeTruthy();
      expect(fixture.nativeElement.querySelector('mat-progress-bar')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('.error-banner')).toBeFalsy();
    });
  });

  describe('Page -> Form: Input Bindings', () => {
    it('should show progress bar, spin icon, and disable submit when loading', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.loading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-progress-bar')).toBeTruthy();
      expect(fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin'))).toBeTruthy();
      expect(fixture.nativeElement.querySelector('button[type="submit"]').disabled).toBe(true);
    });

    it('should show error banner when error is set', () => {
      const { fixture, authMock } = setupTestBed();
      (authMock.error as WritableSignal<string | null>).set('Token scaduto');
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain('Token scaduto');
    });
  });

  describe('Form -> Page: Confirm Account Flow', () => {
    it('should call confirmAccount with token and undefined tenantId, then navigate to /dashboard', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmAccount).toHaveBeenCalledWith({
        token: 'confirm-token',
        tenantId: undefined,
        newPassword: 'newSecret123',
      });
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should include tenantId from query params when present', () => {
      const { fixture, authMock, routerMock } = setupTestBed({ tenantId: 'tenant-01' });
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmAccount).toHaveBeenCalledWith({
        token: 'confirm-token',
        tenantId: 'tenant-01',
        newPassword: 'newSecret123',
      });
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });

    it('should use the token from route params', () => {
      const { fixture, authMock } = setupTestBed({ token: 'custom-token' });
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmAccount).toHaveBeenCalledWith(
        expect.objectContaining({ token: 'custom-token' }),
      );
    });

    it('should not navigate when confirmAccount returns EMPTY', () => {
      const { fixture, authMock, routerMock } = setupTestBed();
      authMock.confirmAccount.mockReturnValue(EMPTY);
      fixture.detectChanges();
      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmAccount).toHaveBeenCalled();
      expect(routerMock.navigate).not.toHaveBeenCalled();
    });

    it('should not call confirmAccount when form is empty', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      submitEmpty(fixture);

      expect(authMock.confirmAccount).not.toHaveBeenCalled();
    });

    it('should not call confirmAccount when passwords do not match', () => {
      const { fixture, authMock } = setupTestBed();
      fixture.detectChanges();
      const form = getConfirmForm(fixture);
      form.controls.newPassword.setValue('password1');
      form.controls.confirmNewPassword.setValue('password2');
      fixture.detectChanges();
      submitEmpty(fixture);

      expect(authMock.confirmAccount).not.toHaveBeenCalled();
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

    it('should show minlength error for short password', () => {
      const { fixture } = setupTestBed();
      fixture.detectChanges();
      const form = getConfirmForm(fixture);
      form.controls.newPassword.setValue('short');
      form.controls.newPassword.markAsTouched();
      fixture.detectChanges();

      const errorTexts = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errorTexts.some((t: string) => t.includes('almeno 8 caratteri'))).toBe(true);
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

  describe('Full Confirm Flow', () => {
    it('should submit, call confirmAccount, and navigate to dashboard', () => {
      const { fixture, authMock, routerMock } = setupTestBed({ tenantId: 'tenant-01' });
      fixture.detectChanges();

      fillAndSubmitForm(fixture, 'newSecret123');

      expect(authMock.confirmAccount).toHaveBeenCalledWith({
        token: 'confirm-token',
        tenantId: 'tenant-01',
        newPassword: 'newSecret123',
      });
      expect(routerMock.navigate).toHaveBeenCalledWith(['/dashboard']);
    });
  });
});
