import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';
import { of, throwError } from 'rxjs';

import { SensorCommandsDialog } from './sensor-commands.dialog';
import { SensorService } from '../../../../services/sensor/sensor.service';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { SensorStatus } from '../../../../models/sensor-status.enum';
import { ApiError } from '../../../../models/api-error.model';

const COMMAND_CASES: ['interrupt' | 'resume', 'interruptSensor' | 'resumeSensor', SensorStatus][] = [
  ['interrupt', 'interruptSensor', SensorStatus.ACTIVE],
  ['resume', 'resumeSensor', SensorStatus.INACTIVE],
];

describe('SensorCommandsDialog', () => {
  let fixture: ComponentFixture<SensorCommandsDialog>;
  let component: SensorCommandsDialog;
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let sensorServiceMock: {
    interruptSensor: ReturnType<typeof vi.fn>;
    resumeSensor: ReturnType<typeof vi.fn>;
  };

  const mockSensor: Sensor = {
    id: 's-1',
    gatewayId: 'gw-1',
    name: 'Temperature',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: SensorStatus.ACTIVE,
    dataInterval: 60,
  };

  const mockInactiveSensor: Sensor = { ...mockSensor, status: SensorStatus.INACTIVE };

  const buildFixture = async (sensor: Sensor) => {
    await TestBed.configureTestingModule({
      imports: [SensorCommandsDialog],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: { sensor, mode: 'manage' } },
        { provide: SensorService, useValue: sensorServiceMock },
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(SensorCommandsDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  };

  const sendBtn = () => fixture.debugElement.query(By.css('button[color="primary"]'));
  const cancelBtn = () =>
    fixture.debugElement
      .queryAll(By.css('button'))
      .find((btn) => btn.nativeElement.textContent.includes('Annulla'))!;
  const selectCommand = (value: string) => {
    component['commandForm'].controls.command.setValue(value);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    vi.resetAllMocks();
    dialogRefMock = { close: vi.fn() };
    sensorServiceMock = {
      interruptSensor: vi.fn(),
      resumeSensor: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [SensorCommandsDialog],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: { sensor: mockSensor, mode: 'manage' } },
        { provide: SensorService, useValue: sensorServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(SensorCommandsDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('command filtering by status', () => {
    it('should only show interrupt command when sensor is ACTIVE', () => {
      expect(component['commands']).toEqual([{ value: 'interrupt', label: 'Interrompi' }]);
    });

    it('should only show resume command when sensor is INACTIVE', async () => {
      TestBed.resetTestingModule();
      await buildFixture(mockInactiveSensor);
      expect(component['commands']).toEqual([{ value: 'resume', label: 'Riprendi' }]);
    });
  });

  describe('initial state', () => {
    it('should create, show title, sensor name, and disabled Send button', () => {
      expect(component).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Comando Sensore');
      const input: HTMLInputElement = fixture.debugElement.query(
        By.css('input[disabled]'),
      ).nativeElement;
      expect(input.value).toBe('Temperature');
      expect(component['commandForm'].controls.command.value).toBe('');
      expect(sendBtn().nativeElement.disabled).toBe(true);
    });
  });

  describe('form validation', () => {
    it('should enable Send button once a command is selected', () => {
      selectCommand('interrupt');
      expect(sendBtn().nativeElement.disabled).toBe(false);
    });

    it('should show required error when command is touched without selection', () => {
      component['commandForm'].controls.command.markAsTouched();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-error')).nativeElement.textContent).toContain(
        'Campo obbligatorio',
      );
    });

    it('should mark command as touched and not call any service when form is invalid', () => {
      component['onConfirm']();
      fixture.detectChanges();
      expect(component['commandForm'].controls.command.touched).toBe(true);
      expect(sensorServiceMock.interruptSensor).not.toHaveBeenCalled();
      expect(sensorServiceMock.resumeSensor).not.toHaveBeenCalled();
    });
  });

  describe('cancel', () => {
    it('should close with false when Cancel is clicked', () => {
      cancelBtn().nativeElement.click();
      expect(dialogRefMock.close).toHaveBeenCalledWith(false);
    });
  });

  describe('command execution', () => {
    it.each(COMMAND_CASES)(
      '%s: should call %s with sensor id and close with true on success',
      async (command, method, status) => {
        if (status !== SensorStatus.ACTIVE) {
          TestBed.resetTestingModule();
          await buildFixture(mockInactiveSensor);
        }
        sensorServiceMock[method].mockReturnValue(of(void 0));
        selectCommand(command);
        sendBtn().nativeElement.click();
        expect(sensorServiceMock[method]).toHaveBeenCalledWith('s-1');
        expect(dialogRefMock.close).toHaveBeenCalledWith(true);
      },
    );

    it.each(COMMAND_CASES)(
      '%s: should set generalError, reset isSubmitting, and keep dialog open on error',
      async (command, method, status) => {
        if (status !== SensorStatus.ACTIVE) {
          TestBed.resetTestingModule();
          await buildFixture(mockInactiveSensor);
        }
        sensorServiceMock[method].mockReturnValue(
          throwError(() => ({ message: `${command} failed` }) as ApiError),
        );
        selectCommand(command);
        sendBtn().nativeElement.click();
        expect(component['generalError']()).toBe(`${command} failed`);
        expect(component['isSubmitting']()).toBe(false);
        expect(dialogRefMock.close).not.toHaveBeenCalled();
      },
    );

    it('should use fallback error message when API error has no message', () => {
      sensorServiceMock.interruptSensor.mockReturnValue(
        throwError(() => ({ status: 500 }) as ApiError),
      );
      selectCommand('interrupt');
      sendBtn().nativeElement.click();
      expect(component['generalError']()).toBe('Invio comando fallito');
    });
  });

  describe('dismissError', () => {
    it('should clear generalError', () => {
      sensorServiceMock.interruptSensor.mockReturnValue(
        throwError(() => ({ message: 'oops' }) as ApiError),
      );
      selectCommand('interrupt');
      sendBtn().nativeElement.click();
      expect(component['generalError']()).toBe('oops');

      component['dismissError']();
      expect(component['generalError']()).toBe('');
    });
  });
});
