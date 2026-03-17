import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ResetPasswordFormComponent } from './reset-password-form.component';

describe('ResetPasswordFormComponent', () => {
  let component: ResetPasswordFormComponent;
  let fixture: ComponentFixture<ResetPasswordFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ResetPasswordFormComponent],
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

  describe('onSubmit', () => {
    it('should emit submitReset with password when form is valid', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('secret123');

      component['onSubmit']();

      expect(emitSpy).toHaveBeenCalledWith('secret123');
    });

    it('should not emit when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['onSubmit']();

      expect(emitSpy).not.toHaveBeenCalled();
    });

    it('should mark all fields as touched when form is invalid', () => {
      expect(component['resetPasswordForm'].controls.newPassword.touched).toBe(false);
      expect(component['resetPasswordForm'].controls.confirmNewPassword.touched).toBe(false);

      component['onSubmit']();

      expect(component['resetPasswordForm'].controls.newPassword.touched).toBe(true);
      expect(component['resetPasswordForm'].controls.confirmNewPassword.touched).toBe(true);
    });

    it('should not emit when passwords do not match', () => {
      const emitSpy = vi.fn();
      component.submitReset.subscribe(emitSpy);

      component['resetPasswordForm'].controls.newPassword.setValue('secret123');
      component['resetPasswordForm'].controls.confirmNewPassword.setValue('different');

      component['onSubmit']();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });

  describe('outputs', () => {
    it('should emit goToLogin', () => {
      const emitSpy = vi.fn();
      component.goToLogin.subscribe(emitSpy);

      component.goToLogin.emit();

      expect(emitSpy).toHaveBeenCalled();
    });

    it('should emit dismissError', () => {
      const emitSpy = vi.fn();
      component.dismissError.subscribe(emitSpy);

      component.dismissError.emit();

      expect(emitSpy).toHaveBeenCalled();
    });
  });

  describe('inputs', () => {
    it('should accept loading input', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(component.loading()).toBe(true);
    });

    it('should accept generalError input', () => {
      fixture.componentRef.setInput('generalError', 'Something went wrong');
      fixture.detectChanges();

      expect(component.generalError()).toBe('Something went wrong');
    });

    it('should accept success input', () => {
      fixture.componentRef.setInput('success', true);
      fixture.detectChanges();

      expect(component.success()).toBe(true);
    });
  });
});
