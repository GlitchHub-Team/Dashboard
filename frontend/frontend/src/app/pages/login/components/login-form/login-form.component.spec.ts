import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA, signal } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';

import { LoginFormComponent } from './login-form.component';
import { LoginRequest } from '../../../../models/auth/login-request.model';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { UserRole } from '../../../../models/user/user-role.enum';

describe('LoginFormComponent', () => {
  let component: LoginFormComponent;
  let fixture: ComponentFixture<LoginFormComponent>;

  const tenantServiceMock = {
    retrieveTenants: vi.fn(),
    tenantList: signal([{ id: 'tenant-01', name: 'Tenant One' }]).asReadonly(),
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

  function submitForm(): void {
    fixture.debugElement.query(By.css('form')).triggerEventHandler('ngSubmit');
    fixture.detectChanges();
  }

  describe('initial state', () => {
    it('should create with invalid form and default input values', () => {
      expect(component).toBeTruthy();
      expect(component['loginForm'].valid).toBe(false);
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
    });

    it('should call retrieveTenants on init', () => {
      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalled();
    });

    it('should render the form but not the progress bar or error banner', () => {
      expect(fixture.debugElement.query(By.css('form'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('mat-progress-bar'))).toBeFalsy();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeFalsy();
    });

    it('should not show the tenant dropdown before a role is selected', () => {
      expect(component['showTenantDropdown']).toBe(false);
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

  describe('role dropdown and tenant visibility', () => {
    it('should not show tenant dropdown when Super Admin is selected', () => {
      component['loginForm'].controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
      expect(component['showTenantDropdown']).toBe(false);
    });

    it('should show tenant dropdown when Tenant Admin is selected', () => {
      component['loginForm'].controls.role.setValue(UserRole.TENANT_ADMIN);
      fixture.detectChanges();
      expect(component['showTenantDropdown']).toBe(true);
    });

    it('should show tenant dropdown when Tenant User is selected', () => {
      component['loginForm'].controls.role.setValue(UserRole.TENANT_USER);
      fixture.detectChanges();
      expect(component['showTenantDropdown']).toBe(true);
    });

    it('should clear tenantId when switching to Super Admin', () => {
      component['loginForm'].controls.role.setValue(UserRole.TENANT_ADMIN);
      component['loginForm'].controls.tenantId.setValue('tenant-01');
      component['loginForm'].controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
      expect(component['loginForm'].controls.tenantId.value).toBe('');
    });

    it('should make tenantId not required when Super Admin is selected', () => {
      component['loginForm'].controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
      expect(component['loginForm'].controls.tenantId.errors).toBeNull();
    });
  });

  describe('form validation', () => {
    it.each([
      ['empty fields', '', '', '' as UserRole],
      ['email only', 'user@example.com', '', '' as UserRole],
      ['password only', '', 'secret123', '' as UserRole],
      ['email and password, no role', 'user@example.com', 'secret123', '' as UserRole],
      ['invalid email format', 'not-an-email', 'secret123', UserRole.SUPER_ADMIN],
    ])('should be invalid with %s', (_, email, password, role) => {
      component['loginForm'].controls.email.setValue(email);
      component['loginForm'].controls.password.setValue(password);
      component['loginForm'].controls.role.setValue(role);
      expect(component['loginForm'].valid).toBe(false);
    });

    it('should be valid for Super Admin without tenantId', () => {
      component['loginForm'].controls.email.setValue('admin@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
      expect(component['loginForm'].valid).toBe(true);
    });

    it('should be valid for Tenant Admin with email, password, role, and tenantId', () => {
      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.role.setValue(UserRole.TENANT_ADMIN);
      component['loginForm'].controls.tenantId.setValue('tenant-01');
      fixture.detectChanges();
      expect(component['loginForm'].valid).toBe(true);
    });

    it('should be invalid for Tenant User without tenantId', () => {
      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.role.setValue(UserRole.TENANT_USER);
      fixture.detectChanges();
      expect(component['loginForm'].valid).toBe(false);
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
    it('should emit submitLogin with mapped userRole and no tenantId for Super Admin', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('admin@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
      submitForm();

      const expected: LoginRequest = {
        email: 'admin@example.com',
        password: 'secret123',
        userRole: 'super_admin',
        tenantId: undefined,
      };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should emit submitLogin with mapped userRole and tenantId for Tenant Admin', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('user@example.com');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.role.setValue(UserRole.TENANT_ADMIN);
      component['loginForm'].controls.tenantId.setValue('tenant-01');
      fixture.detectChanges();
      submitForm();

      const expected: LoginRequest = {
        email: 'user@example.com',
        password: 'secret123',
        userRole: 'tenant_admin',
        tenantId: 'tenant-01',
      };
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit and should mark all fields touched when form is invalid', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      submitForm();

      expect(emitSpy).not.toHaveBeenCalled();
      expect(component['loginForm'].controls.email.touched).toBe(true);
      expect(component['loginForm'].controls.password.touched).toBe(true);
      expect(component['loginForm'].controls.role.touched).toBe(true);
    });

    it('should not emit with invalid email format', () => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);

      component['loginForm'].controls.email.setValue('bad-email');
      component['loginForm'].controls.password.setValue('secret123');
      component['loginForm'].controls.role.setValue(UserRole.SUPER_ADMIN);
      fixture.detectChanges();
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
