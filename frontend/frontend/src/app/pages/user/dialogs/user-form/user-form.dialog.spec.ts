import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
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
  isSubmitting: boolean;
  generalError: string;
  onSave: () => void;
  onCancel: () => void;
}

describe('UserFormDialogComponent', () => {
  let fixture: ComponentFixture<UserFormDialogComponent>;
  let component: UserFormDialogComponent;
  let testApi: UserFormDialogTestApi;

  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let tenantServiceMock: {
    retrieveTenants: ReturnType<typeof vi.fn>;
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
      tenantList: signal(tenantList).asReadonly(),
    };
    userServiceMock = { addNewUser: vi.fn() };
  });

  describe('initialization', () => {
    it('should create with default form values', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      expect(component).toBeTruthy();
      expect(testApi.form.value).toEqual({ username: '', email: '', tenantId: '' });
      expect(testApi.isSubmitting).toBe(false);
      expect(testApi.generalError).toBe('');
    });

    it('should not call retrieveTenants for TENANT_USER role', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      expect(tenantServiceMock.retrieveTenants).not.toHaveBeenCalled();
    });

    it('should call retrieveTenants for TENANT_ADMIN role', async () => {
      await createComponent({ role: UserRole.TENANT_ADMIN });

      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalledTimes(1);
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

    it('should not require tenantId for TENANT_USER', async () => {
      await createComponent({ role: UserRole.TENANT_USER });
      testApi.form.setValue({ username: 'u', email: 'u@example.com', tenantId: '' });

      expect(testApi.form.controls.tenantId.hasError('required')).toBe(false);
      expect(testApi.form.valid).toBe(true);
    });

    it('should require tenantId for TENANT_ADMIN', async () => {
      await createComponent({ role: UserRole.TENANT_ADMIN });
      testApi.form.setValue({ username: 'u', email: 'u@example.com', tenantId: '' });

      expect(testApi.form.controls.tenantId.hasError('required')).toBe(true);
      expect(testApi.form.valid).toBe(false);
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
    it('should call service with form config and close dialog with true on success', async () => {
      await createComponent({ role: UserRole.TENANT_USER, tenantId: 'tenant-1' });
      userServiceMock.addNewUser.mockReturnValue(
        of({ id: 'u1', username: 'new.user', email: 'new.user@example.com' }),
      );
      testApi.form.setValue({ username: 'new.user', email: 'new.user@example.com', tenantId: '' });

      testApi.onSave();

      expect(userServiceMock.addNewUser).toHaveBeenCalledWith(
        { username: 'new.user', email: 'new.user@example.com' },
        UserRole.TENANT_USER,
        'tenant-1',
      );
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should use tenantId from form for TENANT_ADMIN role', async () => {
      await createComponent({ role: UserRole.TENANT_ADMIN });
      userServiceMock.addNewUser.mockReturnValue(
        of({ id: 'u1', username: 'admin', email: 'admin@example.com' }),
      );
      testApi.form.setValue({
        username: 'admin',
        email: 'admin@example.com',
        tenantId: 'tenant-2',
      });

      testApi.onSave();

      expect(userServiceMock.addNewUser).toHaveBeenCalledWith(
        { username: 'admin', email: 'admin@example.com' },
        UserRole.TENANT_ADMIN,
        'tenant-2',
      );
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should show api error message and reset submitting state on error', async () => {
      await createComponent({ role: UserRole.TENANT_USER });
      userServiceMock.addNewUser.mockReturnValue(
        throwError(() => ({ status: 400, message: 'Username already exists' }) as ApiError),
      );
      testApi.form.setValue({ username: 'new.user', email: 'new.user@example.com', tenantId: '' });

      testApi.onSave();

      expect(testApi.generalError).toBe('Username already exists');
      expect(testApi.isSubmitting).toBe(false);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });

    it('should show fallback error when API message is missing', async () => {
      await createComponent({ role: UserRole.TENANT_USER });
      userServiceMock.addNewUser.mockReturnValue(throwError(() => ({ status: 500 }) as ApiError));
      testApi.form.setValue({ username: 'new.user', email: 'new.user@example.com', tenantId: '' });

      testApi.onSave();

      expect(testApi.generalError).toBe('Failed to create user');
      expect(testApi.isSubmitting).toBe(false);
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

  describe('template behavior', () => {
    it('should not render tenant select for non admin roles', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      expect(
        fixture.debugElement.query(By.css('mat-select[formControlName="tenantId"]')),
      ).toBeNull();
    });

    it('should render tenant select for TENANT_ADMIN role', async () => {
      await createComponent({ role: UserRole.TENANT_ADMIN });

      expect(
        fixture.debugElement.query(By.css('mat-select[formControlName="tenantId"]')),
      ).toBeTruthy();
    });

    it('should render add title', async () => {
      await createComponent({ role: UserRole.TENANT_USER });

      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Aggiungi Nuovo');
    });
  });
});
