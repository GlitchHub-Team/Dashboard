import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { By } from '@angular/platform-browser';

import { HistoricChartFiltersDialog } from './historic-chart-filters.dialog';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { Status } from '../../../../models/gateway-sensor-status.enum';
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
    status: Status.ACTIVE,
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

    it('should display correct title', () => {
      const title = fixture.debugElement.query(By.css('[mat-dialog-title]'));
      expect(title.nativeElement.textContent).toContain('Historic Chart Filters');
    });

    it('should display sensor name', () => {
      const sensorInfo = fixture.debugElement.query(By.css('.sensor-info'));
      expect(sensorInfo.nativeElement.textContent).toContain('Heart Rate Sensor');
    });

    it('should have form with default values', () => {
      const form = component['filtersForm'];
      expect(form.value.dataPointsCounter).toBeNull();
      expect(form.value.from).toBeNull();
      expect(form.value.fromTime).toBeNull();
      expect(form.value.to).toBeNull();
      expect(form.value.toTime).toBeNull();
      expect(form.value.lowerBound).toBeNull();
      expect(form.value.upperBound).toBeNull();
    });

    it('should have Apply button disabled (dataPointsCounter empty by default)', () => {
      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(true);
    });

    it('should have Cancel button enabled', () => {
      const cancelBtn = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('Cancel'));
      expect(cancelBtn!.nativeElement.disabled).toBe(false);
    });

    it('should render all section titles', () => {
      const titles = fixture.debugElement.queryAll(By.css('.section-title'));
      expect(titles).toHaveLength(3);
      expect(titles[0].nativeElement.textContent).toContain('Time Range');
      expect(titles[1].nativeElement.textContent).toContain('Value Bounds');
      expect(titles[2].nativeElement.textContent).toContain('Data Points');
    });
  });

  describe('form validation', () => {
    it('should be invalid with default values (dataPointsCounter required)', () => {
      expect(component['filtersForm'].valid).toBe(false);
    });

    it('should be invalid when dataPointsCounter is null', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(null);
      expect(component['filtersForm'].valid).toBe(false);
    });

    it.each([
      ['from', null],
      ['fromTime', null],
      ['to', null],
      ['toTime', null],
      ['lowerBound', null],
      ['upperBound', null],
    ] as const)('should remain valid when optional field %s is null', (field, value) => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['filtersForm'].controls[field].setValue(value as never);
      expect(component['filtersForm'].valid).toBe(true);
    });

    it('should be invalid when dataPointsCounter is 0', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(0);
      expect(component['filtersForm'].valid).toBe(false);
    });

    it('should be invalid when dataPointsCounter is negative', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(-5);
      expect(component['filtersForm'].valid).toBe(false);
    });

    it('should be valid when dataPointsCounter is 1', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(1);
      expect(component['filtersForm'].valid).toBe(true);
    });

    it('should disable Apply button when dataPointsCounter is null', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(null);
      fixture.detectChanges();

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(true);
    });

    it('should disable Apply button when dataPointsCounter is below minimum', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(0);
      fixture.detectChanges();

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(true);
    });
  });

  describe('validation error messages', () => {
    it('should show "Data points count is required" error', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(null);
      component['filtersForm'].controls.dataPointsCounter.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Data points count'));
      expect(error).toBeTruthy();
    });

    it('should show "Minimum 1 data point" error', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(0);
      component['filtersForm'].controls.dataPointsCounter.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Minimum 1'));
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
        lowerBound: 50,
        upperBound: 200,
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
        valuesInterval: { lowerBound: 50, upperBound: 200 },
        dataPointsCounter: 250,
      });
    });

    it('should include correct sensor from dialog data', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      fixture.detectChanges();

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      applyBtn.nativeElement.click();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.sensor).toEqual(mockSensor);
      expect(result.chartType).toBe(ChartType.HISTORIC);
    });

    it('should omit optional fields when only dataPointsCounter is set', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['onApply']();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.timeInterval).toBeUndefined();
      expect(result.valuesInterval).toBeUndefined();
      expect(result.dataPointsCounter).toBe(100);
    });

    it('should omit timeInterval when only one of from/to is set', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['filtersForm'].controls.from.setValue(new Date('2025-01-01'));
      component['onApply']();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.timeInterval).toBeUndefined();
    });

    it('should omit valuesInterval when only one of lowerBound/upperBound is set', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(100);
      component['filtersForm'].controls.lowerBound.setValue(10);
      component['onApply']();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.valuesInterval).toBeUndefined();
    });

    it('should mark all as touched and not close when form is invalid', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(null);
      component['onApply']();

      expect(component['filtersForm'].controls.from.touched).toBe(true);
      expect(component['filtersForm'].controls.fromTime.touched).toBe(true);
      expect(component['filtersForm'].controls.to.touched).toBe(true);
      expect(component['filtersForm'].controls.toTime.touched).toBe(true);
      expect(component['filtersForm'].controls.lowerBound.touched).toBe(true);
      expect(component['filtersForm'].controls.upperBound.touched).toBe(true);
      expect(component['filtersForm'].controls.dataPointsCounter.touched).toBe(true);
      expect(dialogRefMock.close).not.toHaveBeenCalled();
    });
  });

  describe('cancel', () => {
    it('should close dialog without result when Cancel is clicked', () => {
      const cancelBtn = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('Cancel'));
      cancelBtn!.nativeElement.click();

      expect(dialogRefMock.close).toHaveBeenCalledWith();
    });
  });
});
