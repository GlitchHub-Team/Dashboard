import { ComponentFixture, TestBed } from '@angular/core/testing';
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
  isSubmitting: boolean;
  generalError: string;
  onSave: () => void;
  onCancel: () => void;
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
    it('should create with default form values when dialog data is null', async () => {
      await createComponent(null);

      expect(component).toBeTruthy();
      expect(testApi.tenantForm.value).toEqual({
        name: '',
        canImpersonate: false,
      });
      expect(testApi.isSubmitting).toBe(false);
      expect(testApi.generalError).toBe('');
    });

    it('should patch form values from dialog data', async () => {
      await createComponent({
        name: 'Tenant Alpha',
        canImpersonate: true,
      });

      expect(testApi.tenantForm.value).toEqual({
        name: 'Tenant Alpha',
        canImpersonate: true,
      });
    });

    it('should render title and form controls', async () => {
      await createComponent(null);

      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Aggiungi Tenant');
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
    it('should call service with form config and close dialog on success', async () => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(of({ id: 'tenant-1', name: 'Tenant One' }));
      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: true });

      testApi.onSave();

      expect(tenantServiceMock.addNewTenant).toHaveBeenCalledWith({
        name: 'Tenant One',
        canImpersonate: true,
      });
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });

    it('should use false for canImpersonate when value is nullish', async () => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(of({ id: 'tenant-1', name: 'Tenant One' }));
      testApi.tenantForm.controls.name.setValue('Tenant One');
      testApi.tenantForm.controls.canImpersonate.setValue(false);

      testApi.onSave();

      expect(tenantServiceMock.addNewTenant).toHaveBeenCalledWith({
        name: 'Tenant One',
        canImpersonate: false,
      });
    });

    it('should show api error message and reset submitting state on error', async () => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(
        throwError(() => ({ status: 400, message: 'Duplicate tenant name' }) as ApiError),
      );
      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: false });

      testApi.onSave();
      fixture.detectChanges();

      expect(testApi.generalError).toBe('Duplicate tenant name');
      expect(testApi.isSubmitting).toBe(false);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
      expect(fixture.debugElement.query(By.css('.error-text')).nativeElement.textContent).toContain(
        'Duplicate tenant name',
      );
    });

    it('should show fallback error when API message is missing', async () => {
      await createComponent(null);
      tenantServiceMock.addNewTenant.mockReturnValue(
        throwError(() => ({ status: 500 }) as ApiError),
      );
      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: false });

      testApi.onSave();

      expect(testApi.generalError).toBe('Failed to save tenant');
      expect(testApi.isSubmitting).toBe(false);
    });
  });

  describe('button states', () => {
    it('should disable save button when form is invalid', async () => {
      await createComponent(null);

      const saveButton = fixture.debugElement.query(
        By.css('button[color="primary"]'),
      ).nativeElement;
      expect(saveButton.disabled).toBe(true);
    });

    it('should disable save button while submitting', async () => {
      await createComponent(null);
      testApi.tenantForm.setValue({ name: 'Tenant One', canImpersonate: false });
      testApi.isSubmitting = true;
      fixture.detectChanges();

      const saveButton = fixture.debugElement.query(
        By.css('button[color="primary"]'),
      ).nativeElement;
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
});
