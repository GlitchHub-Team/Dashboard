import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

import { LoginFormComponent } from './login-form.component';
import { LoginRequest } from '../../../../models/auth/login-request.model';

describe('LoginFormComponent', () => {
  let component: LoginFormComponent;
  let fixture: ComponentFixture<LoginFormComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginFormComponent, ReactiveFormsModule],
      schemas: [NO_ERRORS_SCHEMA],
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

    it('should render the form', () => {
      const form = fixture.debugElement.query(By.css('form'));
      expect(form).toBeTruthy();
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

    it('should disable the forgot password button', () => {
      const forgotButton = fixture.debugElement.query(By.css('.forgot-password'));
      expect(forgotButton.nativeElement.disabled).toBe(true);
    });

    it('should render the spin icon in the submit button', () => {
      const spinIcon = fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin'));
      expect(spinIcon).toBeTruthy();
    });
  });

  describe('error state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('generalError', 'Invalid credentials');
      fixture.detectChanges();
    });

    it('should render the error banner', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
    });

    it('should display the error message', () => {
      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner.nativeElement.textContent).toContain('Invalid credentials');
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

  describe('form validation errors in template', () => {
    it('should show required error for email when touched and empty', () => {
      component['loginForm'].controls.email.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement.query(By.css('mat-error'));
      expect(error.nativeElement.textContent).toContain('Email is required');
    });

    it('should show invalid email error when touched with invalid format', () => {
      component['loginForm'].controls.email.setValue('not-an-email');
      component['loginForm'].controls.email.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement.queryAll(By.css('mat-error'));
      const emailError = errors.find((e) => e.nativeElement.textContent.includes('Invalid email'));
      expect(emailError).toBeTruthy();
    });

    it('should show required error for password when touched and empty', () => {
      component['loginForm'].controls.password.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement.queryAll(By.css('mat-error'));
      const passwordError = errors.find((e) =>
        e.nativeElement.textContent.includes('Password is required'),
      );
      expect(passwordError).toBeTruthy();
    });
  });

  describe('onSubmit', () => {
    it('should emit submitLogin with form values when valid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      const expected: LoginRequest = {
        email: 'user@example.com',
        password: 'secret123',
      };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(emitSpy).not.toHaveBeenCalled();
    });

    it('should mark all fields as touched when form is invalid', () => {
      expect(component['loginForm'].controls.email.touched).toBe(false);
      expect(component['loginForm'].controls.password.touched).toBe(false);

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(component['loginForm'].controls.email.touched).toBe(true);
      expect(component['loginForm'].controls.password.touched).toBe(true);
    });

    it('should not emit with invalid email format', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('bad-email');
      component['loginForm'].controls.password.setValue('secret123');
      fixture.detectChanges();

      const form = fixture.debugElement.query(By.css('form'));
      form.triggerEventHandler('ngSubmit');
      fixture.detectChanges();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });

  describe('forgot password', () => {
    it('should emit forgotPassword when forgot password button is clicked', () => {
      const emitSpy = vi.fn();
      component.forgotPassword.subscribe(emitSpy);

      const forgotButton = fixture.debugElement.query(By.css('.forgot-password'));
      forgotButton.triggerEventHandler('click');
      fixture.detectChanges();

      expect(emitSpy).toHaveBeenCalled();
    });
  });
});
