import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

import { LoginFormComponent } from './login-form.component';
import { LoginRequest } from '../../../../models/auth/login-request.model';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { signal } from '@angular/core';

describe('LoginFormComponent', () => {
  let component: LoginFormComponent;
  let fixture: ComponentFixture<LoginFormComponent>;

  const tenantServiceMock = {
    retrieveTenant: vi.fn(),
    tenantList: signal([]).asReadonly(),
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LoginFormComponent, ReactiveFormsModule],
      providers: [{ provide: TenantService, useValue: tenantServiceMock }],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create with invalid form and default input values', () => {
      expect(component).toBeTruthy();
      expect(component['loginForm'].valid).toBe(false);
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
    });

    it('should call retrieveTenant on init', () => {
      expect(tenantServiceMock.retrieveTenant).toHaveBeenCalled();
    });

    it('should render the form but not the progress bar or error banner', () => {
      expect(fixture.debugElement.query(By.css('form'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
    });
  });

  describe('loading state', () => {
    beforeEach(() => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();
    });

    it('should show progress bar, spin icon, and disable buttons', () => {
      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"] mat-icon.spin')),
      ).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('button[type="submit"]')).nativeElement.disabled,
      ).toBe(true);
      expect(fixture.debugElement.query(By.css('.forgot-password')).nativeElement.disabled).toBe(
        true,
      );
    });
  });

  describe('error state', () => {
    it('should show error banner with message and emit dismissError on close', () => {
      fixture.componentRef.setInput('generalError', 'Invalid credentials');
      fixture.detectChanges();

      const errorBanner = fixture.debugElement.query(By.css('.error-banner'));
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.nativeElement.textContent).toContain('Invalid credentials');

      const emitSpy = vi.fn();
      component.dismissError.subscribe(emitSpy);
      errorBanner.query(By.css('button')).triggerEventHandler('click');
      expect(emitSpy).toHaveBeenCalled();
    });
  });

  describe('form validation', () => {
    it.each([
      ['empty fields', '', '', ''],
      ['email only', 'user@example.com', '', ''],
      ['password only', '', 'secret123', ''],
      ['email and password, no tenantId', 'user@example.com', 'secret123', ''],
      ['invalid email format', 'not-an-email', 'secret123', 'tenant-01'],
    ])('should be invalid with %s', (_, email, password, tenantId) => {
      component['loginForm'].controls.email.setValue(email);
      component['loginForm'].controls.password.setValue(password);
      component['loginForm'].controls.tenantId.setValue(tenantId);
      expect(component['loginForm'].valid).toBe(false);
    });

    it('should be valid with valid email, password, and tenantId', () => {
      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.tenantId.setValue('tenant-01');
      expect(component['loginForm'].valid).toBe(true);
    });
  });

  describe('form validation errors in template', () => {
    it('should show required error for email when touched and empty', () => {
      component['loginForm'].controls.email.markAsTouched();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-error')).nativeElement.textContent).toContain(
        'Campo obbligatorio',
      );
    });

    it('should show field errors when touched with invalid values', () => {
      component['loginForm'].controls.email.setValue('not-an-email');
      component['loginForm'].controls.email.markAsTouched();
      component['loginForm'].controls.password.markAsTouched();
      fixture.detectChanges();

      const errors = fixture.debugElement.queryAll(By.css('mat-error'));
      const texts = errors.map((e) => e.nativeElement.textContent);
      expect(texts.some((t) => t.includes('Indirizzo email non valido'))).toBe(true);
      expect(texts.some((t) => t.includes('Campo obbligatorio'))).toBe(true);
    });
  });

  describe('onSubmit', () => {
    function submitForm(): void {
      fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
      fixture.detectChanges();
    }

    it('should emit submitLogin with correct values when form is valid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.tenantId.setValue('tenant-01');
      submitForm();

      const expected: LoginRequest = {
        email: 'user@example.com',
        password: 'secret123',
        tenantId: 'tenant-01',
      };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit and should mark fields touched when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
      expect(component['loginForm'].controls.email.touched).toBe(true);
      expect(component['loginForm'].controls.password.touched).toBe(true);
    });

    it('should not emit with invalid email format', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('bad-email');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.tenantId.setValue('tenant-01');
      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
    });
  });

  describe('forgot password', () => {
    it('should emit forgotPassword when button is clicked', () => {
      const emitSpy = vi.fn();
      component.forgotPassword.subscribe(emitSpy);

      fixture.debugElement.query(By.css('.forgot-password')).triggerEventHandler('click');
      expect(emitSpy).toHaveBeenCalled();
    });
  });
});
