import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';

import { HistoricChartFiltersDialog } from './historic-chart-filters.dialog';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { SensorStatus } from '../../../../models/sensor-status.enum';
import { ChartType } from '../../../../models/chart/chart-type.enum';

describe('HistoricChartFiltersDialog (Unit)', () => {
  let fixture: ComponentFixture<HistoricChartFiltersDialog>;
  let component: HistoricChartFiltersDialog;

  let dialogRefMock: { close: ReturnType<typeof vi.fn> };

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: SensorStatus.ACTIVE,
    dataInterval: 1000,
  };

  const mockDialogData: any = {
    sensor: mockSensor,
    chartType: ChartType.HISTORIC,
  };

  beforeEach(async () => {
    vi.resetAllMocks();

    dialogRefMock = { close: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [HistoricChartFiltersDialog],
      providers: [
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: mockDialogData },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(HistoricChartFiltersDialog);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create the dialog', () => {
      expect(component).toBeTruthy();
    });

    it('should render title, sensor name, and all section titles', () => {
      expect(
        fixture.debugElement.query(By.css('[mat-dialog-title]')).nativeElement.textContent,
      ).toContain('Filtri grafico dati storici');
      expect(
        fixture.debugElement.query(By.css('.sensor-info')).nativeElement.textContent,
      ).toContain('Heart Rate Sensor');
      const titles = fixture.debugElement.queryAll(By.css('.section-title'));
      expect(titles).toHaveLength(2);
      expect(titles[0].nativeElement.textContent).toContain('Intervallo Tempo');
      expect(titles[1].nativeElement.textContent).toContain('Punti Dati');
    });

    it('should have form with default values', () => {
      const form = component['filtersForm'];
      expect(form.value.dataPointsCounter).toBe(50);
      expect(form.value.from).toBeNull();
      expect(form.value.fromTime).toBeNull();
      expect(form.value.to).toBeNull();
      expect(form.value.toTime).toBeNull();
    });

    it('should have both action buttons enabled', () => {
      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(false);
      const cancelBtn = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('Annulla'));
      expect(cancelBtn!.nativeElement.disabled).toBe(false);
    });
  });

  describe('form validation', () => {
    it('should be invalid when dataPointsCounter exceeds maximum', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(301);
      expect(component['filtersForm'].valid).toBe(false);
    });

    it.each([
      ['from', null],
      ['fromTime', null],
      ['to', null],
      ['toTime', null],
    ] as const)('should remain valid when optional field %s is null', (field, value) => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['filtersForm'].controls[field].setValue(value as never);
      expect(component['filtersForm'].valid).toBe(true);
    });

    it.each([0, -5])('should be invalid when dataPointsCounter is %s', (val) => {
      component['filtersForm'].controls.dataPointsCounter.setValue(val);
      expect(component['filtersForm'].valid).toBe(false);
    });

    it.each([1, 300])('should be valid when dataPointsCounter is %s', (val) => {
      component['filtersForm'].controls.dataPointsCounter.setValue(val);
      expect(component['filtersForm'].valid).toBe(true);
    });

    it('should be invalid when from date is not before to date', () => {
      component['filtersForm'].controls.from.setValue(new Date('2025-01-02'));
      component['filtersForm'].controls.fromTime.setValue('12:00');
      component['filtersForm'].controls.to.setValue(new Date('2025-01-01'));
      component['filtersForm'].controls.toTime.setValue('12:00');
      expect(component['filtersForm'].hasError('invalidDateRange')).toBe(true);
      expect(component['filtersForm'].valid).toBe(false);
    });

    it.each([
      [301, 'exceeds maximum'],
      [0, 'is below minimum'],
    ])('should disable Apply button when dataPointsCounter %s', (val) => {
      component['filtersForm'].controls.dataPointsCounter.setValue(val);
      fixture.detectChanges();

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(true);
    });
  });

  describe('validation error messages', () => {
    it('should show "Minimum 1 data point" error', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(0);
      component['filtersForm'].controls.dataPointsCounter.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Almeno 1'));
      expect(error).toBeTruthy();
    });

    it('should show "Maximum 300 data points" error', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(301);
      component['filtersForm'].controls.dataPointsCounter.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Al massimo 300'));
      expect(error).toBeTruthy();
    });

    it('should show invalid date range error', () => {
      component['filtersForm'].controls.from.setValue(new Date('2025-01-02'));
      component['filtersForm'].controls.to.setValue(new Date('2025-01-01'));
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('deve essere precedente'));
      expect(error).toBeTruthy();
    });
  });

  describe('apply', () => {
    it('should close dialog with complete ChartRequest', () => {
      const from = new Date('2025-01-01');
      const to = new Date('2025-01-02');

      component['filtersForm'].setValue({
        from,
        fromTime: '10:30',
        to,
        toTime: '20:00',
        dataPointsCounter: 250,
      });
      fixture.detectChanges();

      const expectedFrom = new Date(from);
      expectedFrom.setHours(10, 30, 0, 0);
      const expectedTo = new Date(to);
      expectedTo.setHours(20, 0, 0, 0);

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      applyBtn.nativeElement.click();

      expect(dialogRefMock.close).toHaveBeenCalledWith({
        sensor: mockSensor,
        chartType: ChartType.HISTORIC,
        timeInterval: { from: expectedFrom, to: expectedTo },
        dataPointsCounter: 250,
      });
    });

    it('should omit optional fields when only dataPointsCounter is set', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['onApply']();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.timeInterval).toBeUndefined();
      expect(result.dataPointsCounter).toBe(100);
    });

    it('should omit timeInterval when only one of from/to is set', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['filtersForm'].controls.from.setValue(new Date('2025-01-01'));
      component['onApply']();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.timeInterval).toBeUndefined();
    });

    it('should mark all as touched and not close when form is invalid', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(0);
      component['onApply']();

      expect(component['filtersForm'].controls.from.touched).toBe(true);
      expect(component['filtersForm'].controls.fromTime.touched).toBe(true);
      expect(component['filtersForm'].controls.to.touched).toBe(true);
      expect(component['filtersForm'].controls.toTime.touched).toBe(true);
      expect(component['filtersForm'].controls.dataPointsCounter.touched).toBe(true);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('cancel', () => {
    it('should close dialog without result when Cancel is clicked', () => {
      const cancelBtn = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('Annulla'));
      cancelBtn!.nativeElement.click();

      expect(dialogRefMock.close).toHaveBeenCalledWith();
    });
  });
});
