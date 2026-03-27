import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

import { ResetPasswordFormComponent } from './reset-password-form.component';
import { ForgotPasswordResponse } from '../../../../models/auth/forgot-password.model';

describe('ResetPasswordFormComponent', () => {
  let component: ResetPasswordFormComponent;
  let fixture: ComponentFixture<ResetPasswordFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ResetPasswordFormComponent, ReactiveFormsModule],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ResetPasswordFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  function submitForm(): void {
    fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
    fixture.detectChanges();
  }

  describe('initial state', () => {
    it('should create with invalid form and default input values', () => {
      expect(component).toBeTruthy();
      expect(component['resetPasswordForm'].valid).toBe(false);
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
      expect(component.success()).toBe(false);
    });

    it('should render form but not success banner, progress bar, or error banner', () => {
      expect(fixture.debugElement.query(By.css('form'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.success-banner'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
    });
  });

  describe('loading state', () => {
    it('should show progress bar, spin icon, and disable both buttons', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin')),
      ).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"]')).nativeElement.disabled,
      ).toBe(true);
      expect(
        fixture.debugElement.query(By.css('button[type="button"]')).nativeElement.disabled,
      ).toBe(true);
    });
  });

  describe('success state', () => {
    it('should show success banner with message, hide form, and emit goToLogin on button click', () => {
      fixture.componentRef.setInput('success', true);
      fixture.detectChanges();

      const successBanner = fixture.debugElement.query(By.css('.success-banner'));
      expect(successBanner).toBeTruthy();
      expect(successBanner.nativeElement.textContent).toContain('Reimpostazione della password riuscita.');
      expect(fixture.debugElement.query(By.css('form'))).toBeFalsy();

      const goToLoginButton = fixture.debugElement.query(By.css('button'));
      expect(goToLoginButton.nativeElement.textContent).toContain('Torna indietro');

      const emitSpy = vi.fn();
      component.goToLogin.subscribe(emitSpy);
      goToLoginButton.triggerEventHandler('click');
      expect(emitSpy).toHaveBeenCalled();
    });
  });

  describe('error state', () => {
    it('should show error banner with message and emit dismissError on close', () => {
      fixture.componentRef.setInput('generalError', 'Something went wrong');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.nativeElement.textContent).toContain('Something went wrong');

      const emitSpy = vi.fn();
      component.dismissError.subscribe(emitSpy);
      errorBanner.query(By.css('button')).triggerEventHandler('click');
      expect(emitSpy).toHaveBeenCalled();
    });
  });

  describe('form validation', () => {
    it.each([
      ['both empty', '', ''],
      ['only newPassword', 'secret123', ''],
      ['only confirmNewPassword', '', 'secret123'],
      ['mismatched passwords', 'secret123', 'different'],
    ])('should be invalid with %s', (_, newPassword, confirmNewPassword) => {
      component['resetPasswordForm'].controls.newPassword.setValue(newPassword);
      component['resetPasswordForm'].controls.confirmNewPassword.setValue(confirmNewPassword);
      expect(component['resetPasswordForm'].valid).toBe(false);
    });

    it('should be valid and have no mismatch error when passwords match', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      expect(component['resetPasswordForm'].valid).toBe(true);
      expect(component['resetPasswordForm'].hasError('passwordMismatch')).toBe(false);
    });

    it('should have passwordMismatch error when passwords do not match', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      expect(component['resetPasswordForm'].hasError('passwordMismatch')).toBe(true);
    });

    it('should not have mismatch error when confirm field is empty', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('');
      expect(component['resetPasswordForm'].hasError('passwordMismatch')).toBe(false);
    });
  });

  describe('form validation errors in template', () => {
    it('should show required errors for each field when touched and empty', () => {
      component['resetPasswordForm'].controls.newPassword.markAsTouched();
      component['resetPasswordForm'].controls.confirmNewPassword.markAsTouched();
      fixture.detectChanges();

      const errorTexts = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errorTexts.some((t) => t.includes('Campo obbligatorio'))).toBe(true);
      expect(errorTexts.some((t) => t.includes('Campo obbligatorio'))).toBe(true);
    });

    it('should show mismatch error when passwords differ and confirm is dirty, hide when they match', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      component['resetPasswordForm'].controls.confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      const mismatchError = fixture.debugElement.query(By.css('.field-error'));
      expect(mismatchError).toBeTruthy();
      expect(mismatchError.nativeElement.textContent).toContain('Le password non coincidono');

      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.field-error'))).toBeFalsy();
    });
  });

  describe('onSubmit', () => {
    it('should emit submitReset with ForgotPasswordResponse when form is valid', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      submitForm();

      const expected: ForgotPasswordResponse = { token: '', newPassword: 'secret123' };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit and should mark both fields touched when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
      expect(component['resetPasswordForm'].controls.newPassword.touched).toBe(true);
      expect(component['resetPasswordForm'].controls.confirmNewPassword.touched).toBe(true);
    });

    it('should not emit when passwords do not match', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });

  describe('back to login', () => {
    it('should emit goToLogin when back to login button is clicked', () => {
      const emitSpy = vi.fn();
      component.goToLogin.subscribe(emitSpy);

      const backButton = fixture.debugElement
        .queryAll(By.css('button[type="button"]'))
        .find((b) => b.nativeElement.textContent.includes('Torna indietro'));
      backButton!.triggerEventHandler('click');

      expect(emitSpy).toHaveBeenCalled();
    });
  });
});
