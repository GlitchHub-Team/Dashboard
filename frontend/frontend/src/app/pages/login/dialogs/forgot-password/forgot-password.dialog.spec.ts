import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal, WritableSignal } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';

import { ForgotPasswordDialog } from './forgot-password.dialog';
import { AuthActionsService } from '../../../../services/auth/auth-actions.service';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { ForgotPasswordRequest } from '../../../../models/auth/forgot-password-request.model';

describe('ForgotPasswordDialog', () => {
  let component: ForgotPasswordDialog;
  let fixture: ComponentFixture<ForgotPasswordDialog>;
  let errorSignal: WritableSignal<string | null>;
  let loadingSignal: WritableSignal<boolean>;
  let authActionsServiceMock: {
    forgotPassword: ReturnType<typeof vi.fn>;
    clearMessages: ReturnType<typeof vi.fn>;
    loading: ReturnType<WritableSignal<boolean>['asReadonly']>;
    error: ReturnType<WritableSignal<string | null>['asReadonly']>;
  };

  const dialogRefMock = { close: vi.fn() };

  const tenantServiceMock = {
    retrieveTenant: vi.fn(),
    tenantList: signal([]).asReadonly(),
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    errorSignal = signal<string | null>(null);
    loadingSignal = signal(false);

    authActionsServiceMock = {
      forgotPassword: vi.fn(),
      clearMessages: vi.fn(),
      loading: loadingSignal.asReadonly(),
      error: errorSignal.asReadonly(),
    };

    await TestBed.configureTestingModule({
      imports: [ForgotPasswordDialog, ReactiveFormsModule],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: TenantService, useValue: tenantServiceMock },
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ForgotPasswordDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  function submitForm(): void {
    fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
    fixture.detectChanges();
  }

  function fillValidForm(email = 'user@example.com', tenantId = 'tenant-01'): void {
    component['forgotPasswordForm'].controls.email.setValue(email);
    component['forgotPasswordForm'].controls.tenantId.setValue(tenantId);
  }

  describe('initial state', () => {
    it('should create with invalid form and render title, description, and form', () => {
      expect(component).toBeTruthy();
      expect(component['forgotPasswordForm'].valid).toBe(false);
      expect(fixture.debugElement.query(By.css('form'))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title] h2')).nativeElement.textContent,
      ).toContain('Reimposta Password');
      expect(
        fixture.debugElement.query(By.css('mat-dialog-content p')).nativeElement.textContent,
      ).toContain("Inserisci il tuo indirizzo email e ti invieremo un link per reimpostare la password.");
    });

    it('should not render progress bar or error banner', () => {
      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
    });

    it('should call retrieveTenant on init', () => {
      expect(tenantServiceMock.retrieveTenant).toHaveBeenCalled();
    });
  });

  describe('loading state', () => {
    it('should show progress bar, spin icon, and disable both buttons', () => {
      loadingSignal.set(true);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin')),
      ).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('mat-dialog-actions button[type="button"]')).nativeElement
          .disabled,
      ).toBe(true);
      expect(
        fixture.debugElement.query(By.css('mat-dialog-actions button[type="submit"]')).nativeElement
          .disabled,
      ).toBe(true);
    });
  });

  describe('error state', () => {
    it('should show error banner with message and call clearMessages on dismiss', () => {
      errorSignal.set('Something went wrong');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.nativeElement.textContent).toContain('Something went wrong');

      errorBanner.query(By.css('button')).triggerEventHandler('click');
      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });

  describe('form validation', () => {
    it.each([
      ['empty email', '', 'tenant-01'],
      ['invalid email format', 'not-an-email', 'tenant-01'],
      ['missing tenantId', 'user@example.com', ''],
    ])('should be invalid with %s', (_, email, tenantId) => {
      component['forgotPasswordForm'].controls.email.setValue(email);
      component['forgotPasswordForm'].controls.tenantId.setValue(tenantId);
      expect(component['forgotPasswordForm'].valid).toBe(false);
    });

    it('should be valid with a proper email and tenantId', () => {
      fillValidForm();
      expect(component['forgotPasswordForm'].valid).toBe(true);
    });
  });

  describe('form validation errors in template', () => {
    it('should show required error when email is touched and empty', () => {
      component['forgotPasswordForm'].controls.email.markAsTouched();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-error')).nativeElement.textContent).toContain(
        'Campo obbligatorio',
      );
    });

    it('should show format error when email is touched with invalid value', () => {
      component['forgotPasswordForm'].controls.email.setValue('not-an-email');
      component['forgotPasswordForm'].controls.email.markAsTouched();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-error')).nativeElement.textContent).toContain(
        'Campo obbligatorio',
      );
    });
  });

  describe('onSubmit', () => {
    it('should call forgotPassword with email and tenantId, and close dialog with true on success', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(of(undefined));
      fillValidForm('user@example.com', 'tenant-01');
      submitForm();

      const expectedRequest: ForgotPasswordRequest = {
        email: 'user@example.com',
        tenantId: 'tenant-01',
      };
      expect(authActionsServiceMock.forgotPassword).toHaveBeenCalledWith(expectedRequest);
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should not close dialog when forgotPassword does not complete', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(EMPTY);
      fillValidForm();
      submitForm();

      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });

    it('should not call forgotPassword and should mark all controls touched when form is invalid', () => {
      expect(component['forgotPasswordForm'].touched).toBe(false);
      submitForm();

      expect(authActionsServiceMock.forgotPassword).not.toHaveBeenCalled();
      expect(component['forgotPasswordForm'].controls.email.touched).toBe(true);
      expect(component['forgotPasswordForm'].controls.tenantId.touched).toBe(true);
    });
  });

  describe('onCancel', () => {
    it('should call clearMessages and close dialog with false', () => {
      fixture.debugElement
        .query(By.css('mat-dialog-actions button[type="button"]'))
        .triggerEventHandler('click');

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('getFieldError', () => {
    it.each([
      ['', 'Email is required.'],
      ['bad-email', 'Please enter a valid email address.'],
      ['user@example.com', ''],
    ])('email="%s" → "%s"', (value, expected) => {
      component['forgotPasswordForm'].controls.email.setValue(value);
      component['forgotPasswordForm'].controls.email.markAsTouched();
      expect(component['getFieldError']('email', 'Email')).toBe(expected);
    });

    it('should return empty string for unknown field', () => {
      expect(component['getFieldError']('nonexistent', 'Field')).toBe('');
    });
  });

  describe('setupAutoClear', () => {
    it('should call clearMessages on email input when error exists, but not when error is null', () => {
      errorSignal.set('Some error');
      component['forgotPasswordForm'].controls.email.setValue('a');
      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();

      authActionsServiceMock.clearMessages.mockClear();
      errorSignal.set(null);
      component['forgotPasswordForm'].controls.email.setValue('b');
      expect(authActionsServiceMock.clearMessages).not.toHaveBeenCalled();
    });

    it('should call clearMessages on tenantId change when error exists', () => {
      errorSignal.set('Some error');
      component['forgotPasswordForm'].controls.tenantId.setValue('tenant-01');
      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });
});
