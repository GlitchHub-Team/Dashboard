import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef } from '@angular/material/dialog';
import { signal, WritableSignal } from '@angular/core';
import { of, EMPTY } from 'rxjs';

import { ForgotPasswordDialog } from './forgot-password.dialog';
import { AuthActionsService } from '../../../../services/auth/auth-actions.service';

describe('ForgotPasswordDialog', () => {
  let component: ForgotPasswordDialog;
  let fixture: ComponentFixture<ForgotPasswordDialog>;

  // Writable so we can change values during tests
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
      imports: [ForgotPasswordDialog],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: MatDialogRef, useValue: dialogRefMock },
      ],
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

  describe('onSubmit', () => {
    it('should call forgotPassword with email when form is valid', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(of(undefined));
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');

      component['onSubmit']();

      expect(authActionsServiceMock.forgotPassword).toHaveBeenCalledWith('user@example.com');
    });

    it('should close dialog with true on success', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(of(undefined));
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');

      component['onSubmit']();

      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should not close dialog on error', () => {
      authActionsServiceMock.forgotPassword.mockReturnValue(EMPTY);
      component['forgotPasswordForm'].controls.email.setValue('user@example.com');

      component['onSubmit']();

      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });

    it('should not call forgotPassword when form is invalid', () => {
      component['onSubmit']();

      expect(authActionsServiceMock.forgotPassword).not.toHaveBeenCalled();
    });

    it('should mark fields as touched when form is invalid', () => {
      expect(component['forgotPasswordForm'].controls.email.touched).toBe(false);

      component['onSubmit']();

      expect(component['forgotPasswordForm'].controls.email.touched).toBe(true);
    });
  });

  describe('onCancel', () => {
    it('should call clearMessages', () => {
      component['onCancel']();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });

    it('should close dialog with false', () => {
      component['onCancel']();

      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('dismissError', () => {
    it('should call clearMessages', () => {
      component['dismissError']();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
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
      // Simulate an existing error
      errorSignal.set('Some error');

      // Type in the email field
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
