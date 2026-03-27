import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, throwError, Observable } from 'rxjs';

import { CreateSensorDialog } from './create-sensor.dialog';
import { SensorService } from '../../../../services/sensor/sensor.service';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { ApiError } from '../../../../models/api-error.model';
import { sensorProfilesMapper } from '../../../../utils/sensor-profile.utils';

describe('CreateSensorDialog (Unit)', () => {
  let fixture: ComponentFixture<CreateSensorDialog>;
  let component: CreateSensorDialog;

  const mockDialogData = { id: 'gw-1', name: 'Gateway Alpha' };
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let sensorServiceMock: { addNewSensor: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    vi.resetAllMocks();
    dialogRefMock = { close: vi.fn() };
    sensorServiceMock = { addNewSensor: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [CreateSensorDialog],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: mockDialogData },
        { provide: SensorService, useValue: sensorServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(CreateSensorDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  const fillValidForm = () => {
    component['sensorForm'].setValue({
      name: 'Heart Rate Sensor',
      profile: SensorProfiles.HEART_RATE_SERVICE,
      interval: 1000,
    });
    fixture.detectChanges();
  };
  const submitBtn = () => fixture.debugElement.query(By.css('button[color="primary"]'));
  const cancelBtn = () =>
    fixture.debugElement
      .queryAll(By.css('button'))
      .find((btn) => btn.nativeElement.textContent.includes('Cancel'))!;

  describe('initial state', () => {
    it('should create with correct title, form defaults, gateway display, and no error/spinner', () => {
      expect(component).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('New Sensor');
      expect(component['sensorForm'].value).toEqual({ name: '', profile: null, interval: 1000 });
      expect(fixture.debugElement.query(By.css('input[disabled]')).nativeElement.value).toBe(
        'Gateway Alpha',
      );
      expect(fixture.debugElement.query(By.css('mat-hint')).nativeElement.textContent).toContain(
        'gw-1',
      );
      expect(submitBtn().nativeElement.disabled).toBe(true);
      expect(cancelBtn().nativeElement.disabled).toBe(false);
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeNull();
    });

    it('should populate profile dropdown with all sensor profiles', () => {
      expect(component['profiles']).toHaveLength(Object.keys(SensorProfiles).length);
    });
  });

  describe('form validation', () => {
    it('should be invalid when empty and valid when all fields filled', () => {
      expect(component['sensorForm'].valid).toBe(false);
      fillValidForm();
      expect(component['sensorForm'].valid).toBe(true);
      expect(submitBtn().nativeElement.disabled).toBe(false);
    });

    it('should be invalid and show min error when interval is below 100', () => {
      fillValidForm();
      component['sensorForm'].controls.interval.setValue(50);
      component['sensorForm'].controls.interval.markAsTouched();
      fixture.detectChanges();
      expect(component['sensorForm'].valid).toBe(false);
      const minError = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Minimum'));
      expect(minError!.nativeElement.textContent).toContain('Minimum 100ms');
    });
  });

  describe('submit with invalid form', () => {
    it('should mark all fields as touched and not call service', () => {
      component['onSubmit']();
      fixture.detectChanges();
      expect(sensorServiceMock.addNewSensor).not.toHaveBeenCalled();
      expect(component['sensorForm'].controls.name.touched).toBe(true);
      expect(component['sensorForm'].controls.profile.touched).toBe(true);
      expect(component['sensorForm'].controls.interval.touched).toBe(true);
      expect(submitBtn().nativeElement.disabled).toBe(true);
    });
  });

  describe('submit success', () => {
    beforeEach(() => {
      fillValidForm();
      sensorServiceMock.addNewSensor.mockReturnValue(of({}));
    });

    it('should call service with correct config, close dialog, and reset submitting state', () => {
      submitBtn().nativeElement.click();
      expect(sensorServiceMock.addNewSensor).toHaveBeenCalledWith({
        gatewayId: 'gw-1',
        name: 'Heart Rate Sensor',
        profile: sensorProfilesMapper.toBackend(SensorProfiles.HEART_RATE_SERVICE),
        dataInterval: 1000,
      });
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
      expect(component['isSubmitting']).toBe(false);
    });
  });

  describe('submit error', () => {
    it('should show API error message, not close dialog, and reset submitting state', () => {
      fillValidForm();
      sensorServiceMock.addNewSensor.mockReturnValue(
        throwError(() => ({ message: 'Duplicate sensor name' }) as ApiError),
      );
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      const banner = fixture.debugElement.query(By.css('.error-banner'));
      expect(banner.nativeElement.textContent).toContain('Duplicate sensor name');
      expect(dialogRefMock.close).not.toHaveBeenCalled();
      expect(component['isSubmitting']).toBe(false);
    });

    it('should show fallback error when API error has no message', () => {
      fillValidForm();
      sensorServiceMock.addNewSensor.mockReturnValue(
        throwError(() => ({ status: 500 }) as ApiError),
      );
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(
        fixture.debugElement.query(By.css('.error-banner')).nativeElement.textContent,
      ).toContain('Failed to create sensor');
    });
  });

  describe('submitting state', () => {
    it('should disable both buttons and show spinner while submitting', () => {
      fillValidForm();
      sensorServiceMock.addNewSensor.mockReturnValue(new Observable(() => {}));
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
      sensorServiceMock.addNewSensor.mockReturnValue(
        throwError(() => ({ message: 'Server error' }) as ApiError),
      );
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      sensorServiceMock.addNewSensor.mockReturnValue(of({}));
      submitBtn().nativeElement.click();
      fixture.detectChanges();
      expect(component['generalError']).toBe('');
      expect(dialogRefMock.close).toHaveBeenCalledWith(true);
    });
  });
});
