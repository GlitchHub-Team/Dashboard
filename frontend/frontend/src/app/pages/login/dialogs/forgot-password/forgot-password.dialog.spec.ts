import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal, WritableSignal } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';

import { ForgotPasswordDialog } from './forgot-password.dialog';
import { AuthActionsService } from '../../../../services/auth/auth-actions.service';

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

  const dialogRefMock = {
    close: vi.fn(),
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
      ],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ForgotPasswordDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have an invalid form by default', () => {
      expect(component['forgotPasswordForm'].valid).toBe(false);
    });

    it('should render the form', () => {
      const form = fixture.debugElement.query(By.css('form'));
      expect(form).toBeTruthy();
    });

    it('should render the dialog title', () => {
      const title = fixture.debugElement.query(By.css('[mat-dialog-title] h2'));
      expect(title.nativeElement.textContent).toContain('Forgot Password');
    });

    it('should render the description text', () => {
      const description = fixture.debugElement.query(By.css('mat-dialog-content p'));
      expect(description.nativeElement.textContent).toContain(
        "Enter your email and we'll send you a reset link.",
      );
    });

    it('should not render the progress bar by default', () => {
      const progressBar = fixture.debugElement.query(By.css('mat-progress-bar'));
      expect(progressBar).toBeFalsy();
    });

    it('should not render the error banner by default', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeFalsy();
    });
  });

  describe('loading state', () => {
    beforeEach(() => {
      loadingSignal.set(true);
      fixture.detectChanges();
    });

    it('should render the progress bar', () => {
      const progressBar = fixture.debugElement.query(By.css('mat-progress-bar'));
      expect(progressBar).toBeTruthy();
    });

    it('should disable the cancel button', () => {
      const cancelButton = fixture.debugElement.query(
        By.css('mat-dialog-actions button[type="button"]'),
      );
      expect(cancelButton.nativeElement.disabled).toBe(true);
    });

    it('should disable the submit button', () => {
      const submitButton = fixture.debugElement.query(
        By.css('mat-dialog-actions button[type="submit"]'),
      );
      expect(submitButton.nativeElement.disabled).toBe(true);
    });

    it('should render the spin icon in the submit button', () => {
      const spinIcon = fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin'));
      expect(spinIcon).toBeTruthy();
    });
  });

  describe('error state', () => {
    beforeEach(() => {
      errorSignal.set('Something went wrong');
      fixture.detectChanges();
    });

    it('should render the error banner', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
    });

    it('should display the error message', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner.nativeElement.textContent).toContain('Something went wrong');
    });

    it('should call clearMessages when close button is clicked', () => {
      const closeButton = fixture.debugElement.query(By.css('.error-banner button'));
      closeButton.triggerEventHandler('click');
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });

  describe('form validation', () => {
    it('should be invalid when email is empty', () => {
      expect(component['forgotPasswordForm'].valid).toBe(false);
    });

    it('should be invalid with bad email format', () => {
      component['forgotPasswordForm'].controls.email.setValue('not-an-email');
      expect(component['forgotPasswordForm'].valid).toBe(false);
      expect(component['forgotPasswordForm'].controls.email.hasError('email')).toBe(true);
    });

    it('should be valid with a proper email', () => {
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');
      expect(component['forgotPasswordForm'].valid).toBe(true);
    });
  });

  describe('form validation errors in template', () => {
    it('should show required error when email is touched and empty', () => {
      component['forgotPasswordForm'].controls.email.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement.query(By.css('mat-error'));
      expect(error).toBeTruthy();
      expect(error.nativeElement.textContent).toContain('Email is required.');
    });

    it('should show invalid email error when touched with invalid format', () => {
      component['forgotPasswordForm'].controls.email.setValue('not-an-email');
      component['forgotPasswordForm'].controls.email.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement.query(By.css('mat-error'));
      expect(error).toBeTruthy();
      expect(error.nativeElement.textContent).toContain('Please enter a valid email address.');
    });
  });

  describe('onSubmit', () => {
    it('should call forgotPassword with email when form is valid', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(of(undefined));
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(authActionsServiceMock.forgotPassword).toHaveBeenCalledWith('user@example.com');
    });

    it('should close dialog with true on success', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(of(undefined));
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should not close dialog on error', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(EMPTY);
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });

    it('should not call forgotPassword when form is invalid', () => {
      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(authActionsServiceMock.forgotPassword).not.toHaveBeenCalled();
    });

    it('should mark fields as touched when form is invalid', () => {
      expect(component['forgotPasswordForm'].controls.email.touched).toBe(false);

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(component['forgotPasswordForm'].controls.email.touched).toBe(true);
    });
  });

  describe('onCancel', () => {
    it('should call clearMessages and close dialog with false', () => {
      const cancelButton = fixture.debugElement.query(
        By.css('mat-dialog-actions button[type="button"]'),
      );
      cancelButton.triggerEventHandler('click');
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('getFieldError', () => {
    it('should return required message when field is empty', () => {
      component['forgotPasswordForm'].controls.email.setValue('');
      component['forgotPasswordForm'].controls.email.markAsTouched();

      expect(component['getFieldError']('email', 'Email')).toBe('Email is required.');
    });

    it('should return email format message when email is invalid', () => {
      component['forgotPasswordForm'].controls.email.setValue('bad-email');
      component['forgotPasswordForm'].controls.email.markAsTouched();

      expect(component['getFieldError']('email', 'Email')).toBe(
        'Please enter a valid email address.',
      );
    });

    it('should return empty string when no errors', () => {
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');

      expect(component['getFieldError']('email', 'Email')).toBe('');
    });

    it('should return empty string for unknown field', () => {
      expect(component['getFieldError']('nonexistent', 'Field')).toBe('');
    });
  });

  describe('setupAutoClear', () => {
    it('should call clearMessages when user types and error exists', () => {
      errorSignal.set('Some error');

      component['forgotPasswordForm'].controls.email.setValue('a');

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });

    it('should not call clearMessages when there is no error', () => {
      errorSignal.set(null);

      component['forgotPasswordForm'].controls.email.setValue('a');

      expect(authActionsServiceMock.clearMessages).not.toHaveBeenCalled();
    });
  });
});
