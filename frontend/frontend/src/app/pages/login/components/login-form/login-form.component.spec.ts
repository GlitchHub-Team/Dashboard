import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LoginFormComponent } from './login-form.component';
import { LoginRequest } from '../../../../models/login-request.model';

describe('LoginFormComponent', () => {
  let component: LoginFormComponent;
  let fixture: ComponentFixture<LoginFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginFormComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have an invalid form by default', () => {
      expect(component['loginForm'].valid).toBe(false);
    });

    it('should have default input values', () => {
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
    });
  });

  describe('form validation', () => {
    it('should be invalid when both fields are empty', () => {
      expect(component['loginForm'].valid).toBe(false);
    });

    it('should be invalid when only email is filled', () => {
      component['loginForm'].controls.email.setValue('user@example.com');

      expect(component['loginForm'].valid).toBe(false);
    });

    it('should be invalid when only password is filled', () => {
      component['loginForm'].controls.password.setValue('secret123');

      expect(component['loginForm'].valid).toBe(false);
    });

    it('should be invalid with an invalid email format', () => {
      component['loginForm'].controls.email.setValue('not-an-email');
      component['loginForm'].controls.password.setValue('secret123');

      expect(component['loginForm'].valid).toBe(false);
      expect(component['loginForm'].controls.email.hasError('email')).toBe(true);
    });

    it('should be valid with valid email and password', () => {
      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');

      expect(component['loginForm'].valid).toBe(true);
    });
  });

  describe('onSubmit', () => {
    it('should emit submitLogin with form values when valid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');

      component['onSubmit']();

      const expected: LoginRequest = {
        email: 'user@example.com',
        password: 'secret123',
      };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['onSubmit']();

      expect(emitSpy).not.toHaveBeenCalled();
    });

    it('should mark all fields as touched when form is invalid', () => {
      expect(component['loginForm'].controls.email.touched).toBe(false);
      expect(component['loginForm'].controls.password.touched).toBe(false);

      component['onSubmit']();

      expect(component['loginForm'].controls.email.touched).toBe(true);
      expect(component['loginForm'].controls.password.touched).toBe(true);
    });

    it('should not emit with invalid email format', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('bad-email');
      component['loginForm'].controls.password.setValue('secret123');

      component['onSubmit']();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });

  describe('outputs', () => {
    it('should emit forgotPassword', () => {
      const emitSpy = vi.fn();
      component.forgotPassword.subscribe(emitSpy);

      component.forgotPassword.emit();

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
      fixture.componentRef.setInput('generalError', 'Invalid credentials');
      fixture.detectChanges();

      expect(component.generalError()).toBe('Invalid credentials');
    });
  });
});
