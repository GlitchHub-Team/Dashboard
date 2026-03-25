import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

import { ConfirmAccountFormComponent } from './confirm-account-form.component';
import { ConfirmAccountResponse } from '../../../../models/auth/confirm-account.model';

describe('ConfirmAccountFormComponent', () => {
  let component: ConfirmAccountFormComponent;
  let fixture: ComponentFixture<ConfirmAccountFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ConfirmAccountFormComponent, ReactiveFormsModule],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmAccountFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  function submitForm(): void {
    fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
    fixture.detectChanges();
  }

  describe('initial state', () => {
    it('should create with an invalid form and default input values', () => {
      expect(component).toBeTruthy();
      expect(component['confirmAccountForm'].valid).toBe(false);
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
    });

    it('should render the form without progress bar or error banner', () => {
      expect(fixture.debugElement.query(By.css('form'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
    });
  });

  describe('loading state', () => {
    it('should show progress bar, spin icon, and disable the submit button', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin')),
      ).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"]')).nativeElement.disabled,
      ).toBe(true);
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

    it('should hide error banner when generalError is null', () => {
      fixture.componentRef.setInput('generalError', null);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
    });
  });

  describe('form validation', () => {
    it.each([
      ['both empty', '', ''],
      ['only newPassword filled', 'secret123', ''],
      ['only confirmNewPassword filled', '', 'secret123'],
      ['mismatched passwords', 'secret123', 'different'],
    ])('should be invalid with %s', (_, newPassword, confirmNewPassword) => {
      component['confirmAccountForm'].controls.newPassword.setValue(newPassword);
      component['confirmAccountForm'].controls.confirmNewPassword.setValue(confirmNewPassword);
      expect(component['confirmAccountForm'].valid).toBe(false);
    });

    it('should be valid when both passwords match', () => {
      component['confirmAccountForm'].controls.newPassword.setValue('secret123');
      component['confirmAccountForm'].controls.confirmNewPassword.setValue('secret123');
      expect(component['confirmAccountForm'].valid).toBe(true);
    });

    it('should have passwordMismatch error when passwords differ', () => {
      component['confirmAccountForm'].controls.newPassword.setValue('secret123');
      component['confirmAccountForm'].controls.confirmNewPassword.setValue('different');
      expect(component['confirmAccountForm'].hasError('passwordMismatch')).toBe(true);
    });

    it('should not have passwordMismatch error when confirm field is empty', () => {
      component['confirmAccountForm'].controls.newPassword.setValue('secret123');
      component['confirmAccountForm'].controls.confirmNewPassword.setValue('');
      expect(component['confirmAccountForm'].hasError('passwordMismatch')).toBe(false);
    });

    it('should not have passwordMismatch error when passwords match', () => {
      component['confirmAccountForm'].controls.newPassword.setValue('secret123');
      component['confirmAccountForm'].controls.confirmNewPassword.setValue('secret123');
      expect(component['confirmAccountForm'].hasError('passwordMismatch')).toBe(false);
    });
  });

  describe('form validation errors in template', () => {
    it('should show required error for newPassword when touched and empty', () => {
      component['confirmAccountForm'].controls.newPassword.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errors.some((t) => t.includes('New Password is required'))).toBe(true);
    });

    it('should show required error for confirmNewPassword when touched and empty', () => {
      component['confirmAccountForm'].controls.confirmNewPassword.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .map((e) => e.nativeElement.textContent);
      expect(errors.some((t) => t.includes('Confirm New Password is required'))).toBe(true);
    });
  });

  describe('onSubmit', () => {
    it('should emit submitConfirmAccount with correct payload when form is valid', () => {
      const emitSpy = vi.fn();
      component.submitConfirmAccount.subscribe(emitSpy);

      component['confirmAccountForm'].controls.newPassword.setValue('secret123');
      component['confirmAccountForm'].controls.confirmNewPassword.setValue('secret123');
      submitForm();

      const expected: ConfirmAccountResponse = { token: '', newPassword: 'secret123' };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit and should mark both fields as touched when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitConfirmAccount.subscribe(emitSpy);

      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
      expect(component['confirmAccountForm'].controls.newPassword.touched).toBe(true);
      expect(component['confirmAccountForm'].controls.confirmNewPassword.touched).toBe(true);
    });

    it('should not emit when passwords do not match', () => {
      const emitSpy = vi.fn();
      component.submitConfirmAccount.subscribe(emitSpy);

      component['confirmAccountForm'].controls.newPassword.setValue('secret123');
      component['confirmAccountForm'].controls.confirmNewPassword.setValue('different');
      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });
});
