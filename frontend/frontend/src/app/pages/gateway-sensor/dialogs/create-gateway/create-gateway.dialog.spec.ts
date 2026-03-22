import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef } from '@angular/material/dialog';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { By } from '@angular/platform-browser';
import { of, throwError, Observable } from 'rxjs';

import { CreateGatewayDialog } from './create-gateway.dialog';
import { GatewayService } from '../../../../services/gateway/gateway.service';
import { ApiError } from '../../../../models/api-error.model';

describe('CreateGatewayDialog (Unit)', () => {
  let fixture: ComponentFixture<CreateGatewayDialog>;
  let component: CreateGatewayDialog;
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let gatewayServiceMock: { addNewGateway: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    vi.resetAllMocks();
    dialogRefMock = { close: vi.fn() };
    gatewayServiceMock = { addNewGateway: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [CreateGatewayDialog, NoopAnimationsModule],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: GatewayService, useValue: gatewayServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(CreateGatewayDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  const fillValidForm = () => {
    component['gatewayForm'].setValue({ name: 'Gateway Alpha', interval: 1000 });
    fixture.detectChanges();
  };
  const submitBtn = () => fixture.debugElement.query(By.css('button[color="primary"]'));
  const cancelBtn = () => fixture.debugElement.queryAll(By.css('button')).find((btn) => btn.nativeElement.textContent.includes('Cancel'))!;

  describe('initial state', () => {
    it('should create with correct title, form defaults, and no error/spinner', () => {
      expect(component).toBeTruthy();
      expect(fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent).toContain('New Gateway');
      expect(component['gatewayForm'].value).toEqual({ name: '', interval: 1000 });
      expect(submitBtn().nativeElement.disabled).toBe(true);
      expect(submitBtn().nativeElement.textContent).toContain('Create');
      expect(cancelBtn().nativeElement.disabled).toBe(false);
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeNull();
    });
  });

  describe('form validation', () => {
    it('should be invalid when empty and valid when correctly filled', () => {
      expect(component['gatewayForm'].valid).toBe(false);
      fillValidForm();
      expect(component['gatewayForm'].valid).toBe(true);
      expect(submitBtn().nativeElement.disabled).toBe(false);
    });

    it('should be invalid when name is empty and show required error when touched', () => {
      fillValidForm();
      component['gatewayForm'].controls.name.setValue('');
      component['gatewayForm'].controls.name.markAsTouched();
      fixture.detectChanges();
      expect(component['gatewayForm'].valid).toBe(false);
      expect(fixture.debugElement.query(By.css('mat-error')).nativeElement.textContent).toContain('Name is required');
    });

    it('should be invalid when interval is below 100 and show min error when touched', () => {
      fillValidForm();
      component['gatewayForm'].controls.interval.setValue(50);
      component['gatewayForm'].controls.interval.markAsTouched();
      fixture.detectChanges();
      expect(component['gatewayForm'].valid).toBe(false);
      const minError = fixture.debugElement.queryAll(By.css('mat-error')).find((e) => e.nativeElement.textContent.includes('Minimum'));
      expect(minError!.nativeElement.textContent).toContain('Minimum 100ms');
    });
  });

  describe('submit with invalid form', () => {
    it('should mark all fields as touched and not call service', () => {
      component['onSubmit']();
      fixture.detectChanges();
      expect(gatewayServiceMock.addNewGateway).not.toHaveBeenCalled();
      expect(component['gatewayForm'].controls.name.touched).toBe(true);
      expect(component['gatewayForm'].controls.interval.touched).toBe(true);
      expect(submitBtn().nativeElement.disabled).toBe(true);
    });
  });

  describe('submit success', () => {
    it('should call service with correct config and close dialog with true', () => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(of({}));
      submitBtn().nativeElement.click();
      expect(gatewayServiceMock.addNewGateway).toHaveBeenCalledWith({ name: 'Gateway Alpha', interval: 1000 });
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });
  });

  describe('submit error', () => {
    it('should show API error message, not close dialog, and reset submitting state', () => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(throwError(() => ({ message: 'Duplicate gateway name' }) as ApiError));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      const banner = fixture.debugElement.query(By.css('.error-banner'));
      expect(banner.nativeElement.textContent).toContain('Duplicate gateway name');
      expect(dialogRefMock.close).not.toHaveBeenCalled();
      expect(component['isSubmitting']).toBe(false);
    });

    it('should show fallback error when API error has no message', () => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(throwError(() => ({ status: 500 }) as ApiError));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner')).nativeElement.textContent).toContain('Failed to create gateway');
    });
  });

  describe('submitting state', () => {
    it('should disable both buttons and show spinner while submitting', () => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(new Observable(() => {}));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(submitBtn().nativeElement.disabled).toBe(true);
      expect(cancelBtn().nativeElement.disabled).toBe(true);
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
    });
  });

  describe('cancel', () => {
    it('should close dialog with false when cancel is clicked', () => {
      cancelBtn().nativeElement.click();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('error recovery', () => {
    it('should clear error banner on successful retry', () => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(throwError(() => ({ message: 'Server error' }) as ApiError));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      gatewayServiceMock.addNewGateway.mockReturnValue(of({}));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(component['generalError']).toBe('');
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });
  });
});
