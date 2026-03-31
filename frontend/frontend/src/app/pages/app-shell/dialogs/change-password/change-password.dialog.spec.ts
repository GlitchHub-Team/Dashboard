import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef } from '@angular/material/dialog';
import { WritableSignal, signal } from '@angular/core';
import { By } from '@angular/platform-browser';
import { of, EMPTY } from 'rxjs';

import { ChangePasswordDialog } from './change-password.dialog';
import { AuthActionsService } from '../../../../services/auth/auth-actions.service';

describe('ChangePasswordDialog', () => {
  let component: ChangePasswordDialog;
  let fixture: ComponentFixture<ChangePasswordDialog>;
  let loadingSignal: WritableSignal<boolean>;
  let errorSignal: WritableSignal<string | null>;
  let authActionsServiceMock: {
    confirmPasswordChange: ReturnType<typeof vi.fn>;
    clearMessages: ReturnType<typeof vi.fn>;
    loading: ReturnType<WritableSignal<boolean>['asReadonly']>;
    error: ReturnType<WritableSignal<string | null>['asReadonly']>;
  };

  const dialogRefMock = { close: vi.fn() };

  beforeEach(async () => {
    vi.resetAllMocks();
    loadingSignal = signal(false);
    errorSignal = signal<string | null>(null);
    authActionsServiceMock = {
      confirmPasswordChange: vi.fn(),
      clearMessages: vi.fn(),
      loading: loadingSignal.asReadonly(),
      error: errorSignal.asReadonly(),
    };

    await TestBed.configureTestingModule({
      imports: [ChangePasswordDialog],
      providers: [
        { provide: AuthActionsService, useValue: authActionsServiceMock },
        { provide: MatDialogRef, useValue: dialogRefMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ChangePasswordDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('Template: static structure', () => {
    it('should render title, 3 password inputs, and action buttons', () => {
      const el = fixture.nativeElement as HTMLElement;
      expect(el.querySelector('[mat-dialog-title]')!.textContent).toContain('Cambia password');
      expect(el.querySelectorAll('input').length).toBe(3);
      const btnTexts = Array.from(el.querySelectorAll('button')).map((b) => b.textContent?.trim());
      expect(btnTexts).toContain('Annulla');
      expect(btnTexts.some((t) => t?.includes('Cambia Password'))).toBe(true);
    });
  });

  describe('Template: conditional rendering', () => {
    it.each([
      { loading: true, expectBar: true },
      { loading: false, expectBar: false },
    ])(
      'should show progress bar = $expectBar when loading = $loading',
      ({ loading, expectBar }) => {
        loadingSignal.set(loading);
        fixture.detectChanges();

        const bar = fixture.nativeElement.querySelector('mat-progress-bar');
        expect(!!bar).toBe(expectBar);
      },
    );

    it('should show error banner with message when generalError is set, hide when null', () => {
      errorSignal.set('Something went wrong');
      fixture.detectChanges();
      const banner = fixture.nativeElement.querySelector('.error-banner');
      expect(banner).toBeTruthy();
      expect(banner.textContent).toContain('Something went wrong');

      errorSignal.set(null);
      fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('.error-banner')).toBeFalsy();
    });

    it('should disable both action buttons when loading', () => {
      loadingSignal.set(true);
      fixture.detectChanges();

      const buttons = Array.from(
        (fixture.nativeElement as HTMLElement).querySelectorAll<HTMLButtonElement>(
          'mat-dialog-actions button',
        ),
      );
      buttons.forEach((b) => expect(b.disabled).toBe(true));
    });

    it('should show passwordMismatch error when passwords differ and confirmNewPassword is dirty', () => {
      const { newPassword, confirmNewPassword } = component['form'].controls;
      newPassword.setValue('password1');
      confirmNewPassword.setValue('password2');
      confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      const mismatchError = fixture.nativeElement.querySelector('.field-error');
      expect(mismatchError?.textContent).toContain('Le password non coincidono');
    });

    it('should NOT show passwordMismatch error when passwords match', () => {
      const { newPassword, confirmNewPassword } = component['form'].controls;
      newPassword.setValue('samePass');
      confirmNewPassword.setValue('samePass');
      confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('.field-error')).toBeFalsy();
    });

    it('should show required error on oldPassword when touched and empty', () => {
      component['form'].controls.oldPassword.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement.queryAll(By.css('mat-error'));
      expect(errors.some((e) => e.nativeElement.textContent.includes('Campo obbligatorio'))).toBe(
        true,
      );
    });
  });

  describe('onSubmit', () => {
    function submitForm(): void {
      (fixture.nativeElement as HTMLElement)
        .querySelector('form')!
        .dispatchEvent(new Event('submit'));
      fixture.detectChanges();
    }

    function fillForm(): void {
      component['form'].controls.oldPassword.setValue('oldPass');
      component['form'].controls.newPassword.setValue('newPass');
      component['form'].controls.confirmNewPassword.setValue('newPass');
    }

    it('should mark all fields touched and not call service when form is invalid', () => {
      submitForm();

      const { oldPassword, newPassword, confirmNewPassword } = component['form'].controls;
      expect(oldPassword.touched).toBe(true);
      expect(newPassword.touched).toBe(true);
      expect(confirmNewPassword.touched).toBe(true);
      expect(authActionsServiceMock.confirmPasswordChange).not.toHaveBeenCalled();
    });

    it('should call service with correct payload and close(true) on success', () => {
      authActionsServiceMock.confirmPasswordChange.mockReturnValue(of(undefined));
      fillForm();
      submitForm();

      expect(authActionsServiceMock.confirmPasswordChange).toHaveBeenCalledWith({
        oldPassword: 'oldPass',
        newPassword: 'newPass',
      });
      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should not close dialog when service returns EMPTY', () => {
      authActionsServiceMock.confirmPasswordChange.mockReturnValue(EMPTY);
      fillForm();
      submitForm();

      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('onCancel', () => {
    it('should call clearMessages and close(false)', () => {
      const cancelBtn = Array.from(
        (fixture.nativeElement as HTMLElement).querySelectorAll<HTMLButtonElement>(
          'mat-dialog-actions button',
        ),
      ).find((b) => b.textContent?.trim() === 'Annulla');

      cancelBtn!.click();
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('dismissError', () => {
    it('should call clearMessages when error banner close button is clicked', () => {
      errorSignal.set('Something went wrong');
      fixture.detectChanges();

      (fixture.nativeElement as HTMLElement)
        .querySelector<HTMLButtonElement>('.error-banner button')!
        .click();
      fixture.detectChanges();

      expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
    });
  });

  describe('setupAutoClear', () => {
    it.each(['oldPassword', 'newPassword', 'confirmNewPassword'] as const)(
      'should call clearMessages when %s changes and an error is active',
      (field) => {
        errorSignal.set('Some error');
        component['form'].controls[field].setValue('typing');

        expect(authActionsServiceMock.clearMessages).toHaveBeenCalled();
      },
    );

    it('should NOT call clearMessages when a field changes and there is no error', () => {
      component['form'].controls.oldPassword.setValue('typing');

      expect(authActionsServiceMock.clearMessages).not.toHaveBeenCalled();
    });
  });
});
