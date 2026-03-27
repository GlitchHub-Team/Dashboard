import { ComponentFixture, TestBed } from '@angular/core/testing';
import { signal } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UserFormDialogComponent, UserFormDialogData } from './user-form.dialog';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { UserRole } from '../../../../models/user/user-role.enum';
import { User } from '../../../../models/user/user.model';

interface UserFormDialogTestApi {
  form: UserFormDialogComponent['form'];
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

  const existingUser: User = {
    id: 'user-1',
    username: 'john.doe',
    email: 'john.doe@example.com',
    role: UserRole.TENANT_USER,
    tenantId: 'tenant-1',
  };

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
  });

  describe('initialization', () => {
    it('should create and initialize form with user data', async () => {
      await createComponent({ user: existingUser, role: UserRole.TENANT_USER });

      expect(component).toBeTruthy();
      expect(testApi.form.value).toEqual({
        username: 'john.doe',
        email: 'john.doe@example.com',
        tenantId: 'tenant-1',
      });
      expect(tenantServiceMock.retrieveTenants).not.toHaveBeenCalled();
    });

    it('should initialize empty form when user is null', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });

      expect(testApi.form.value).toEqual({
        username: '',
        email: '',
        tenantId: '',
      });
    });

    it('should call retrieveTenants for TENANT_ADMIN role', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_ADMIN });

      expect(tenantServiceMock.retrieveTenants).toHaveBeenCalledTimes(1);
    });
  });

  describe('validation', () => {
    it('should require username and email', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });

      testApi.form.setValue({ username: '', email: '', tenantId: '' });

      expect(testApi.form.controls.username.hasError('required')).toBe(true);
      expect(testApi.form.controls.email.hasError('required')).toBe(true);
      expect(testApi.form.valid).toBe(false);
    });

    it('should validate email format', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });

      testApi.form.controls.email.setValue('invalid-email');
      expect(testApi.form.controls.email.hasError('email')).toBe(true);

      testApi.form.controls.email.setValue('valid@email.com');
      expect(testApi.form.controls.email.hasError('email')).toBe(false);
    });

    it('should require tenantId only for TENANT_ADMIN', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });
      testApi.form.setValue({ username: 'u', email: 'u@example.com', tenantId: '' });
      expect(testApi.form.controls.tenantId.hasError('required')).toBe(false);
      expect(testApi.form.valid).toBe(true);

      await TestBed.resetTestingModule();
      await createComponent({ user: null, role: UserRole.TENANT_ADMIN });
      testApi.form.setValue({ username: 'u', email: 'u@example.com', tenantId: '' });
      expect(testApi.form.controls.tenantId.hasError('required')).toBe(true);
      expect(testApi.form.valid).toBe(false);
    });
  });

  describe('actions', () => {
    it('should close with form value on save when form is valid', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });
      testApi.form.setValue({
        username: 'new.user',
        email: 'new.user@example.com',
        tenantId: '',
      });

      testApi.onSave();

      expect(dialogRefMock.close).toHaveBeenCalledWith({
        username: 'new.user',
        email: 'new.user@example.com',
        tenantId: '',
      });
    });

    it('should not close and should mark controls touched when save is triggered on invalid form', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });
      testApi.form.setValue({ username: '', email: '', tenantId: '' });

      testApi.onSave();

      expect(dialogRefMock.close).not.toHaveBeenCalled();
      expect(testApi.form.controls.username.touched).toBe(true);
      expect(testApi.form.controls.email.touched).toBe(true);
      expect(testApi.form.controls.tenantId.touched).toBe(true);
    });

    it('should close without payload on cancel', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });

      testApi.onCancel();

      expect(dialogRefMock.close).toHaveBeenCalledWith();
    });
  });

  describe('template behavior', () => {
    it('should not render tenant select for non admin roles', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });

      expect(
        fixture.debugElement.query(By.css('mat-select[formControlName="tenantId"]')),
      ).toBeNull();
    });

    it('should render tenant select for TENANT_ADMIN role', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_ADMIN });

      expect(
        fixture.debugElement.query(By.css('mat-select[formControlName="tenantId"]')),
      ).toBeTruthy();
    });

    it('should render edit title when user is provided', async () => {
      await createComponent({ user: existingUser, role: UserRole.TENANT_USER });

      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Modifica Utente');
    });

    it('should render add title when user is null', async () => {
      await createComponent({ user: null, role: UserRole.TENANT_USER });

      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Aggiungi Nuovo Utente');
    });
  });
});
