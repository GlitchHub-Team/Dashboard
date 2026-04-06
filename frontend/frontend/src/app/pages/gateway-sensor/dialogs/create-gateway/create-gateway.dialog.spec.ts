import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef } from '@angular/material/dialog';
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
      imports: [CreateGatewayDialog],
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
  const cancelBtn = () => fixture.debugElement.queryAll(By.css('button')).find((btn) => btn.nativeElement.textContent.includes('Annulla'))!;

  describe('initial state', () => {
    it('should create with correct title, form defaults, and no error/spinner', () => {
      expect(component).toBeTruthy();
      expect(fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent).toContain('Nuovo Gateway');
      expect(component['gatewayForm'].value).toEqual({ name: '', interval: 1000 });
      expect(submitBtn().nativeElement.disabled).toBe(true);
      expect(submitBtn().nativeElement.textContent).toContain('Crea');
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

    it.each([
      { control: 'name', value: '', expectedError: 'Campo obbligatorio' },
      { control: 'interval', value: 50, expectedError: 'Almeno 100ms' },
    ])('should be invalid and show $expectedError for invalid $control', ({ control, value, expectedError }) => {
      fillValidForm();
      (component['gatewayForm'].controls as any)[control].setValue(value);
      (component['gatewayForm'].controls as any)[control].markAsTouched();
      fixture.detectChanges();
      expect(component['gatewayForm'].valid).toBe(false);
      const error = fixture.debugElement.queryAll(By.css('mat-error')).find((e) => e.nativeElement.textContent.includes(expectedError));
      expect(error!.nativeElement.textContent).toContain(expectedError);
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
    it.each([
      [{ message: 'Duplicate gateway name' } as ApiError, 'Duplicate gateway name'],
      [{ status: 500 } as ApiError, 'Failed to create gateway. Please try again.'],
    ])('should show error banner and not close dialog', (error, expectedMsg) => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(throwError(() => error));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner')).nativeElement.textContent).toContain(expectedMsg);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
      expect(component['isSubmitting']()).toBe(false);
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

  describe('dismissError', () => {
    it('should clear error banner when close button is clicked', () => {
      fillValidForm();
      gatewayServiceMock.addNewGateway.mockReturnValue(throwError(() => ({ message: 'Server error' }) as ApiError));
      submitBtn().nativeElement.click();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      fixture.debugElement.query(By.css('.error-banner button')).nativeElement.click();
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
      expect(component['generalError']()).toBe('');
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
      expect(component['generalError']()).toBe('');
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });
  });
});
