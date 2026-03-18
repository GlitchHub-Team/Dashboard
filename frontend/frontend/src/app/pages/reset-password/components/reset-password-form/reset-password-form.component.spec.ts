import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

import { ResetPasswordFormComponent } from './reset-password-form.component';

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

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have an invalid form by default', () => {
      expect(component['resetPasswordForm'].valid).toBe(false);
    });

    it('should have default input values', () => {
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
      expect(component.success()).toBe(false);
    });

    it('should render the form by default', () => {
      const form = fixture.debugElement.query(By.css('form'));
      expect(form).toBeTruthy();
    });

    it('should not render the success banner by default', () => {
      const successBanner = fixture.debugElement.query(By.css('.success-banner'));
      expect(successBanner).toBeFalsy();
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
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();
    });

    it('should render the progress bar', () => {
      const progressBar = fixture.debugElement.query(By.css('mat-progress-bar'));
      expect(progressBar).toBeTruthy();
    });

    it('should disable the submit button', () => {
      const submitButton = fixture.debugElement.query(By.css('button[type="submit"]'));
      expect(submitButton.nativeElement.disabled).toBe(true);
    });

    it('should disable the back to login button', () => {
      const backButton = fixture.debugElement.query(By.css('button[type="button"]'));
      expect(backButton.nativeElement.disabled).toBe(true);
    });

    it('should render the spin icon in the submit button', () => {
      const spinIcon = fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin'));
      expect(spinIcon).toBeTruthy();
    });
  });

  describe('success state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('success', true);
      fixture.detectChanges();
    });

    it('should render the success banner', () => {
      const successBanner = fixture.debugElement.query(By.css('.success-banner'));
      expect(successBanner).toBeTruthy();
    });

    it('should display success message', () => {
      const successBanner = fixture.debugElement.query(By.css('.success-banner'));
      expect(successBanner.nativeElement.textContent).toContain('Password reset successfully');
    });

    it('should not render the form', () => {
      const form = fixture.debugElement.query(By.css('form'));
      expect(form).toBeFalsy();
    });

    it('should render go to login button', () => {
      const goToLoginButton = fixture.debugElement.query(By.css('button'));
      expect(goToLoginButton.nativeElement.textContent).toContain('Go to Login');
    });

    it('should emit goToLogin when go to login button is clicked', () => {
      const emitSpy = vi.fn();
      component.goToLogin.subscribe(emitSpy);

      const goToLoginButton = fixture.debugElement.query(By.css('button'));
      goToLoginButton.triggerEventHandler('click');
      fixture.detectChanges();

      expect(emitSpy).toHaveBeenCalled();
    });
  });

  describe('error state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('generalError', 'Something went wrong');
      fixture.detectChanges();
    });

    it('should render the error banner', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
    });

    it('should display error message', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner.nativeElement.textContent).toContain('Something went wrong');
    });

    it('should emit dismissError when close button is clicked', () => {
      const emitSpy = vi.fn();
      component.dismissError.subscribe(emitSpy);

      const closeButton = fixture.debugElement.query(By.css('.error-banner button'));
      closeButton.triggerEventHandler('click');
      fixture.detectChanges();

      expect(emitSpy).toHaveBeenCalled();
    });
  });

  describe('form validation', () => {
    it('should be invalid when both fields are empty', () => {
      expect(component['resetPasswordForm'].valid).toBe(false);
    });

    it('should be invalid when only newPassword is filled', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      expect(component['resetPasswordForm'].valid).toBe(false);
    });

    it('should be invalid when only confirmNewPassword is filled', () => {
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      expect(component['resetPasswordForm'].valid).toBe(false);
    });

    it('should be valid when both fields match', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      expect(component['resetPasswordForm'].valid).toBe(true);
    });

    it('should be invalid when passwords do not match', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      expect(component['resetPasswordForm'].hasError('passwordMismatch')).toBe(true);
    });

    it('should not have mismatch error when one field is empty', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('');
      expect(component['resetPasswordForm'].hasError('passwordMismatch')).toBe(false);
    });
  });

  describe('form validation errors in template', () => {
    it('should show required error for newPassword when touched and empty', () => {
      component['resetPasswordForm'].controls.newPassword.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement.query(By.css('mat-error'));
      expect(error.nativeElement.textContent).toContain('Password is required');
    });

    it('should show required error for confirmNewPassword when touched and empty', () => {
      component['resetPasswordForm'].controls.confirmNewPassword.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement.queryAll(By.css('mat-error'));
      const confirmError = errors.find((e) =>
        e.nativeElement.textContent.includes('Confirm password is required'),
      );
      expect(confirmError).toBeTruthy();
    });

    it('should show password mismatch error when passwords differ and confirm is dirty', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      component['resetPasswordForm'].controls.confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      const mismatchError = fixture.debugElement.query(By.css('.field-error'));
      expect(mismatchError).toBeTruthy();
      expect(mismatchError.nativeElement.textContent).toContain('Passwords do not match');
    });

    it('should not show password mismatch error when passwords match', () => {
      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.markAsDirty();
      fixture.detectChanges();

      const mismatchError = fixture.debugElement.query(By.css('.field-error'));
      expect(mismatchError).toBeFalsy();
    });
  });

  describe('onSubmit', () => {
    it('should emit submitReset with password when form is valid', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(emitSpy).toHaveBeenCalledWith('secret123');
    });

    it('should not emit when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(emitSpy).not.toHaveBeenCalled();
    });

    it('should mark all fields as touched when form is invalid', () => {
      expect(component['resetPasswordForm'].controls.newPassword.touched).toBe(false);
      expect(component['resetPasswordForm'].controls.confirmNewPassword.touched).toBe(false);

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(component['resetPasswordForm'].controls.newPassword.touched).toBe(true);
      expect(component['resetPasswordForm'].controls.confirmNewPassword.touched).toBe(true);
    });

    it('should not emit when passwords do not match', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });

  describe('back to login', () => {
    it('should emit goToLogin when back to login button is clicked', () => {
      const emitSpy = vi.fn();
      component.goToLogin.subscribe(emitSpy);

      const buttons = fixture.debugElement.queryAll(By.css('button[type="button"]'));
      const backButton = buttons.find((b) => b.nativeElement.textContent.includes('Back to Login'));
      backButton!.triggerEventHandler('click');
      fixture.detectChanges();

      expect(emitSpy).toHaveBeenCalled();
    });
  });
});
