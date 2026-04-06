import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { TenantFormDialog } from './tenant-form.dialog';
import { TenantService } from '../../../../services/tenant/tenant.service';
import { TenantConfig } from '../../../../models/tenant/tenant-config.model';
import { ApiError } from '../../../../models/api-error.model';

interface TenantFormDialogTestApi {
  tenantForm: TenantFormDialog['tenantForm'];
  isSubmitting: WritableSignal<boolean>;
  generalError: WritableSignal<string>;
  onSave: () => void;
  onCancel: () => void;
  dismissError: () => void;
}

describe('TenantFormDialog (Unit)', () => {
  let fixture: ComponentFixture<TenantFormDialog>;
  let component: TenantFormDialog;
  let testApi: TenantFormDialogTestApi;

  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let tenantServiceMock: { addNewTenant: ReturnType<typeof vi.fn> };

  const createComponent = async (data: TenantConfig | null): Promise<void> => {
    await TestBed.configureTestingModule({
      imports: [TenantFormDialog],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: data },
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: TenantService, useValue: tenantServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(TenantFormDialog);
    component = fixture.componentInstance;
    testApi = component as unknown as TenantFormDialogTestApi;
    fixture.detectChanges();
  };

  beforeEach(() => {
    vi.clearAllMocks();
    dialogRefMock = { close: vi.fn() };
    tenantServiceMock = { addNewTenant: vi.fn() };
  });

  describe('initialization', () => {
    it('should create with default form values and render title and form controls', async () => {
      await createComponent(null);

      expect(component).toBeTruthy();
      expect(testApi.tenantForm.value).toEqual({ name: '', canImpersonate: false });
      expect(testApi.isSubmitting()).toBe(false);
      expect(testApi.generalError()).toBe('');
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Aggiungi Nuovo Tenant');
      expect(fixture.debugElement.query(By.css('input[formControlName="name"]'))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('mat-checkbox[formControlName="canImpersonate"]')),
      ).toBeTruthy();
    });
  });

  describe('validation', () => {
    it('should require tenant name', async () => {
      await createComponent(null);

      testApi.tenantForm.controls.name.setValue('');
      expect(testApi.tenantForm.controls.name.hasError('required')).toBe(true);

      testApi.tenantForm.controls.name.setValue('Tenant One');
      expect(testApi.tenantForm.controls.name.hasError('required')).toBe(false);
      expect(testApi.tenantForm.valid).toBe(true);
    });

    it('should mark controls as touched and not call service on invalid save', async () => {
      await createComponent(null);
      testApi.tenantForm.controls.name.setValue('');

      testApi.onSave();

      expect(testApi.tenantForm.controls.name.touched).toBe(true);
      expect(tenantServiceMock.addNewTenant).not.toHaveBeenCalled();
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('onSave', () => {
    it.each([
      [{ name: 'Tenant One', canImpersonate: true }, { name: 'Tenant One', canImpersonate: true }],
      [{ name: 'Tenant One', canImpersonate: false }, { name: 'Tenant One', canImpersonate: false }],
    ])('should call service with correct config and close dialog on success', async (formValue, expectedPayload) => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(of({ id: 'tenant-1', name: formValue.name }));
      testApi.tenantForm.setValue(formValue);

      testApi.onSave();

      expect(tenantServiceMock.addNewTenant).toHaveBeenCalledWith(expectedPayload);
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it.each([
      [{ status: 400, message: 'Duplicate tenant name' } as ApiError, 'Duplicate tenant name'],
      [{ status: 500 } as ApiError, 'Failed to create tenant'],
    ])('should show error and reset submitting on failure', async (error, expectedMsg) => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(throwError(() => error));
      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: false });

      testApi.onSave();
      fixture.detectChanges();

      expect(testApi.generalError()).toBe(expectedMsg);
      expect(testApi.isSubmitting()).toBe(false);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('button states', () => {
    it('should disable save button when form is invalid or while submitting', async () => {
      await createComponent(null);
      const saveButton = fixture.debugElement.query(By.css('button[color="primary"]')).nativeElement;

      expect(saveButton.disabled).toBe(true);

      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: false });
      testApi.isSubmitting.set(true);
      fixture.detectChanges();
      expect(saveButton.disabled).toBe(true);
    });
  });

  describe('onCancel', () => {
    it('should close dialog with false', async () => {
      await createComponent(null);

      testApi.onCancel();

      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('dismissError', () => {
    it('should clear error banner when close button is clicked', async () => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(throwError(() => ({ status: 500 } as ApiError)));
      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: false });
      testApi.onSave();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      fixture.debugElement.query(By.css('.error-banner button')).nativeElement.click();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
    });
  });
});
