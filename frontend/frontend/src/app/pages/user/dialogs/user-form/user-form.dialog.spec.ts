import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal, WritableSignal } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UserFormDialogComponent, UserFormDialogData } from './user-form.dialog';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { UserService } from '../../../../services/user/user.service';
import { UserRole } from '../../../../models/user/user-role.enum';
import { ApiError } from '../../../../models/api-error.model';

interface UserFormDialogTestApi {
  form: UserFormDialogComponent['form'];
  isSubmitting: WritableSignal<boolean>;
  generalError: WritableSignal<string>;
  lockedTenantName: WritableSignal<string | null>;
  isTenantIdLocked: boolean;
  onSave: () => void;
  onCancel: () => void;
  dismissError: () => void;
}

describe('UserFormDialogComponent', () => {
  let fixture: ComponentFixture<UserFormDialogComponent>;
  let component: UserFormDialogComponent;
  let testApi: UserFormDialogTestApi;

  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let tenantServiceMock: {
    retrieveTenants: ReturnType<typeof vi.fn>;
    getTenant: ReturnType<typeof vi.fn>;
    tenantList: ReturnType<ReturnType<typeof signal>['asReadonly']>;
  };
  let userServiceMock: { addNewUser: ReturnType<typeof vi.fn> };

  const tenantList = [
    { id: 'tenant-1', name: 'Tenant One', canImpersonate: true },
    { id: 'tenant-2', name: 'Tenant Two', canImpersonate: false },
  ];

  const createComponent = async (data: UserFormDialogData): Promise<void> => {
    await TestBed.configureTestingModule({
      imports: [UserFormDialogComponent],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: data },
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: TenantService, useValue: tenantServiceMock },
        { provide: UserService, useValue: userServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(UserFormDialogComponent);
    component = fixture.componentInstance;
    testApi = component as unknown as UserFormDialogTestApi;
    fixture.detectChanges();
  };

  beforeEach(() => {
    vi.clearAllMocks();
    dialogRefMock = { close: vi.fn() };
    tenantServiceMock = {
      retrieveTenants: vi.fn(),
      getTenant: vi.fn().mockReturnValue(of({ id: 'tenant-1', name: 'Tenant One', canImpersonate: true })),
      tenantList: signal(tenantList).asReadonly(),
    };
    userServiceMock = { addNewUser: vi.fn() };
  });

  describe('initialization', () => {
    it('should create with default form values', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      expect(component).toBeTruthy();
      expect(testApi.form.value).toEqual({ username: '', email: '', tenantId: '' });
      expect(testApi.isSubmitting()).toBe(false);
      expect(testApi.generalError()).toBe('');
    });

    it.each([
      [UserRole.TENANT_USER, 0],
      [UserRole.TENANT_ADMIN, 1],
    ])('should call retrieveTenants %i time(s) for role %s when no tenantId provided', async (role, expectedCalls) => {
      await createComponent({ role });
      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalledTimes(expectedCalls);
    });

    describe('locked tenantId (TENANT_ADMIN with tenantId pre-set)', () => {
      it('should pre-fill tenantId form control and NOT call retrieveTenants', async () => {
        await createComponent({ role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' });

        expect(testApi.form.controls.tenantId.value).toBe('tenant-1');
        expect(tenantServiceMock.retrieveTenants).not.toHaveBeenCalled();
      });

      it('should call getTenant to resolve the tenant name', async () => {
        await createComponent({ role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' });

        expect(tenantServiceMock.getTenant).toHaveBeenCalledWith('tenant-1');
        expect(testApi.lockedTenantName()).toBe('Tenant One');
      });

      it('should expose isTenantIdLocked as true', async () => {
        await createComponent({ role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' });

        expect(testApi.isTenantIdLocked).toBe(true);
      });

      it('should expose isTenantIdLocked as false when tenantId is absent', async () => {
        await createComponent({ role: UserRole.TENANT_ADMIN });

        expect(testApi.isTenantIdLocked).toBe(false);
      });
    });
  });

  describe('validation', () => {
    it('should require username and email', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      testApi.form.setValue({ username: '', email: '', tenantId: '' });

      expect(testApi.form.controls.username.hasError('required')).toBe(true);
      expect(testApi.form.controls.email.hasError('required')).toBe(true);
      expect(testApi.form.valid).toBe(false);
    });

    it('should validate email format', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      testApi.form.controls.email.setValue('invalid-email');
      expect(testApi.form.controls.email.hasError('email')).toBe(true);

      testApi.form.controls.email.setValue('valid@email.com');
      expect(testApi.form.controls.email.hasError('email')).toBe(false);
    });

    it.each([
      { role: UserRole.TENANT_USER, required: false, formValid: true },
      { role: UserRole.TENANT_ADMIN, required: true, formValid: false },
    ])('should tenantId required=$required for role $role', async ({ role, required, formValid }) => {
      await createComponent({ role });
      testApi.form.setValue({ username: 'u', email: 'u@example.com', tenantId: '' });

      expect(testApi.form.controls.tenantId.hasError('required')).toBe(required);
      expect(testApi.form.valid).toBe(formValid);
    });

    it('should mark controls as touched and not call service on invalid save', async () => {
      await createComponent({ role: UserRole.TENANT_USER });
      testApi.form.setValue({ username: '', email: '', tenantId: '' });

      testApi.onSave();

      expect(testApi.form.controls.username.touched).toBe(true);
      expect(testApi.form.controls.email.touched).toBe(true);
      expect(userServiceMock.addNewUser).not.toHaveBeenCalled();
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('onSave', () => {
    it.each([
      {
        label: 'TENANT_USER: tenantId from dialog data',
        data: { role: UserRole.TENANT_USER, tenantId: 'tenant-1' } as UserFormDialogData,
        formValue: { username: 'new.user', email: 'new.user@example.com', tenantId: '' },
        expectedArgs: [{ username: 'new.user', email: 'new.user@example.com' }, UserRole.TENANT_USER, 'tenant-1'],
      },
      {
        label: 'TENANT_ADMIN: tenantId from form',
        data: { role: UserRole.TENANT_ADMIN } as UserFormDialogData,
        formValue: { username: 'admin', email: 'admin@example.com', tenantId: 'tenant-2' },
        expectedArgs: [{ username: 'admin', email: 'admin@example.com' }, UserRole.TENANT_ADMIN, 'tenant-2'],
      },
      {
        label: 'TENANT_ADMIN: locked tenantId pre-filled from data',
        data: { role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' } as UserFormDialogData,
        formValue: { username: 'admin', email: 'admin@example.com', tenantId: 'tenant-1' },
        expectedArgs: [{ username: 'admin', email: 'admin@example.com' }, UserRole.TENANT_ADMIN, 'tenant-1'],
      },
    ])('should call service with correct config and close dialog ($label)', async ({ data, formValue, expectedArgs }) => {
      await createComponent(data);
      userServiceMock.addNewUser.mockReturnValue(of({ id: 'u1' }));
      testApi.form.setValue(formValue);

      testApi.onSave();

      expect(userServiceMock.addNewUser).toHaveBeenCalledWith(...expectedArgs);
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it.each([
      [{ status: 400, message: 'Username already exists' } as ApiError, 'Username already exists'],
      [{ status: 500 } as ApiError, 'Failed to create user'],
    ])('should show error and reset submitting on failure', async (error, expectedMsg) => {
      await createComponent({ role: UserRole.TENANT_USER });
      userServiceMock.addNewUser.mockReturnValue(throwError(() => error));
      testApi.form.setValue({ username: 'new.user', email: 'new.user@example.com', tenantId: '' });

      testApi.onSave();

      expect(testApi.generalError()).toBe(expectedMsg);
      expect(testApi.isSubmitting()).toBe(false);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('button states', () => {
    it('should disable save button when form is invalid', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      const saveButton = fixture.debugElement.query(
        By.css('button[color="primary"]'),
      ).nativeElement;
      expect(saveButton.disabled).toBe(true);
    });
  });

  describe('onCancel', () => {
    it('should close dialog without payload', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      testApi.onCancel();

      expect(dialogRefMock.close).toHaveBeenCalledWith();
    });
  });

  describe('dismissError', () => {
    it('should clear error banner when close button is clicked', async () => {
      await createComponent({ role: UserRole.TENANT_USER });
      userServiceMock.addNewUser.mockReturnValue(throwError(() => ({ status: 500 } as ApiError)));
      testApi.form.setValue({ username: 'u', email: 'u@example.com', tenantId: '' });
      testApi.onSave();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      fixture.debugElement.query(By.css('.error-banner button')).nativeElement.click();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
    });
  });

  describe('template behavior', () => {
    it.each([
      [UserRole.TENANT_USER, false],
      [UserRole.TENANT_ADMIN, true],
    ])('should render tenant select: %s for role %s when no tenantId provided', async (role, shouldRender) => {
      await createComponent({ role });
      const select = fixture.debugElement.query(By.css('mat-select[formControlName="tenantId"]'));
      if (shouldRender) {
        expect(select).toBeTruthy();
      } else {
        expect(select).toBeNull();
      }
    });

    it('should render a disabled input (not mat-select) when tenantId is locked', async () => {
      await createComponent({ role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' });
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-select[formControlName="tenantId"]'))).toBeNull();
      const disabledInput = fixture.debugElement.query(By.css('input[disabled]'));
      expect(disabledInput).toBeTruthy();
    });

    it('should show tenant name in locked input once resolved', async () => {
      await createComponent({ role: UserRole.TENANT_ADMIN, tenantId: 'tenant-1' });
      fixture.detectChanges();

      const disabledInput: HTMLInputElement = fixture.debugElement.query(By.css('input[disabled]')).nativeElement;
      expect(disabledInput.value).toBe('Tenant One');
    });

    it('should render add title', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Aggiungi Nuovo');
    });
  });
});
