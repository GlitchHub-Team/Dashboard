import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ReactiveFormsModule } from '@angular/forms';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';

import { LoginFormComponent } from './login-form.component';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { Tenant } from '../../../../models/tenant/tenant.model';
import { ApiError } from '../../../../models/api-error.model';

describe('LoginFormComponent', () => {
  let component: LoginFormComponent;
  let fixture: ComponentFixture<LoginFormComponent>;

  const mockTenants: Tenant[] = [{ id: 'tenant-01', name: 'Tenant One', canImpersonate: false }];

  let tenantServiceMock: { getAllTenants: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    tenantServiceMock = { getAllTenants: vi.fn().mockReturnValue(of(mockTenants)) };

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
    it('should create with invalid form, call getAllTenants, populate displayedTenants, and render form without progress/error', () => {
      expect(component).toBeTruthy();
      expect(component['loginForm'].valid).toBe(false);
      expect(component.loading()).toBe(false);
      expect(component.generalError()).toBeNull();
      expect(tenantServiceMock.getAllTenants).toHaveBeenCalled();
      expect(component['displayedTenants']()).toEqual(mockTenants);
      expect(component['tenantLoadingError']()).toBeNull();
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
      { label: 'empty fields', email: '', password: '', tenantId: '', expected: false },
      { label: 'email only', email: 'user@example.com', password: '', tenantId: '', expected: false },
      { label: 'password only', email: '', password: 'secret123', tenantId: '', expected: false },
      { label: 'invalid email format', email: 'not-an-email', password: 'secret123', tenantId: '', expected: false },
      { label: 'email and password without tenantId', email: 'user@example.com', password: 'secret123', tenantId: '', expected: true },
      { label: 'email and password with tenantId', email: 'user@example.com', password: 'secret123', tenantId: 'tenant-01', expected: true },
    ])('should be valid=$expected for $label', ({ email, password, tenantId, expected }) => {
      component['loginForm'].controls.email.setValue(email);
      component['loginForm'].controls.password.setValue(password);
      if (tenantId) component['loginForm'].controls.tenantId.setValue(tenantId);
      fixture.detectChanges();
      expect(component['loginForm'].valid).toBe(expected);
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
    it.each([
      { label: 'without tenantId', email: 'admin@example.com', password: 'secret123', tenantId: undefined as string | undefined, expected: { email: 'admin@example.com', password: 'secret123', tenantId: undefined } },
      { label: 'with tenantId', email: 'user@example.com', password: 'secret123', tenantId: 'tenant-01', expected: { email: 'user@example.com', password: 'secret123', tenantId: 'tenant-01' } },
    ])('should emit submitLogin $label', ({ email, password, tenantId, expected }) => {
      const emitSpy = vi.fn();
      component.submitLogin.subscribe(emitSpy);
      component['loginForm'].controls.email.setValue(email);
      component['loginForm'].controls.password.setValue(password);
      if (tenantId !== undefined) component['loginForm'].controls.tenantId.setValue(tenantId);
      fixture.detectChanges();
      submitForm();
      expect(emitSpy).toHaveBeenCalledWith(expected);
    });

    it('should not emit and should mark all fields touched when form is invalid', () => {
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

describe('LoginFormComponent - getAllTenants error', () => {
  it.each([
    [{ status: 500, message: 'Server error' } as ApiError, 'Server error'],
    [{ status: 500 } as ApiError, 'Failed to fetch tenants'],
  ])('should set tenantLoadingError when getAllTenants fails', async (error, expected) => {
    const errorTenantServiceMock = { getAllTenants: vi.fn().mockReturnValue(throwError(() => error)) };

    await TestBed.configureTestingModule({
      imports: [LoginFormComponent, ReactiveFormsModule],
      providers: [{ provide: TenantService, useValue: errorTenantServiceMock }],
      schemas: [NO_ERRORS_SCHEMA],
    }).compileComponents();

    const errorFixture = TestBed.createComponent(LoginFormComponent);
    const errorComponent = errorFixture.componentInstance;
    errorFixture.detectChanges();

    expect(errorComponent['tenantLoadingError']()).toBe(expected);
    expect(errorComponent['displayedTenants']()).toEqual([]);
  });
});
