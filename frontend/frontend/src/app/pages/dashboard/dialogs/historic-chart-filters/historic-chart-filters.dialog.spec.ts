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
      expect(form.value.dataPointsCounter).toBe(100);
      expect(form.value.from).toBeInstanceOf(Date);
      expect(form.value.to).toBeInstanceOf(Date);
      expect(form.value.lowerBound).toBe(0);
      expect(form.value.upperBound).toBe(100);
    });

    it('should default "from" to approximately 24 hours ago', () => {
      const from = component['filtersForm'].value.from!;
      const twentyFourHoursAgo = new Date(Date.now() - 24 * 60 * 60 * 1000);
      expect(Math.abs(from.getTime() - twentyFourHoursAgo.getTime())).toBeLessThan(1000);
    });

    it('should default "to" to approximately now', () => {
      const to = component['filtersForm'].value.to!;
      expect(Math.abs(to.getTime() - Date.now())).toBeLessThan(1000);
    });

    it('should have Apply button enabled (form valid by default)', () => {
      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(false);
    });

    it('should have Cancel button enabled', () => {
      const cancelBtn = fixture.debugElement
        .queryAll(By.css('button'))
        .find((btn) => btn.nativeElement.textContent.includes('Annulla'));
      expect(cancelBtn!.nativeElement.disabled).toBe(false);
    });

    it('should render all section titles', () => {
      const titles = fixture.debugElement.queryAll(By.css('.section-title'));
      expect(titles).toHaveLength(3);
      expect(titles[0].nativeElement.textContent).toContain('Intervallo Tempo');
      expect(titles[1].nativeElement.textContent).toContain('Limiti di valore');
      expect(titles[2].nativeElement.textContent).toContain('Punti Dati');
    });
  });

  describe('form validation', () => {
    it('should be valid with default values', () => {
      expect(component['filtersForm'].valid).toBe(true);
    });

    it.each([
      ['from', null],
      ['to', null],
      ['lowerBound', null],
      ['upperBound', null],
      ['dataPointsCounter', null],
    ] as const)('should be invalid when %s is cleared', (field, value) => {
      component['filtersForm'].controls[field].setValue(value!);
      expect(component['filtersForm'].valid).toBe(false);
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

    it('should disable Apply button when form is invalid', () => {
      component['filtersForm'].controls.from.setValue(null!);
      fixture.detectChanges();

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      expect(applyBtn.nativeElement.disabled).toBe(true);
    });
  });

  describe('validation error messages', () => {
    it('should show "Start date is required" error', () => {
      component['filtersForm'].controls.from.setValue(null!);
      component['filtersForm'].controls.from.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Campo obbligatorio'));
      expect(error).toBeTruthy();
    });

    it('should show "End date is required" error', () => {
      component['filtersForm'].controls.to.setValue(null!);
      component['filtersForm'].controls.to.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Campo obbligatorio'));
      expect(error).toBeTruthy();
    });

    it('should show "Lower bound is required" error', () => {
      component['filtersForm'].controls.lowerBound.setValue(null!);
      component['filtersForm'].controls.lowerBound.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Campo obbligatorio'));
      expect(error).toBeTruthy();
    });

    it('should show "Upper bound is required" error', () => {
      component['filtersForm'].controls.upperBound.setValue(null!);
      component['filtersForm'].controls.upperBound.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Campo obbligatorio'));
      expect(error).toBeTruthy();
    });

    it('should show "Data points count is required" error', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(null!);
      component['filtersForm'].controls.dataPointsCounter.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Campo obbligatorio'));
      expect(error).toBeTruthy();
    });

    it('should show "Minimum 1 data point" error', () => {
      component['filtersForm'].controls.dataPointsCounter.setValue(0);
      component['filtersForm'].controls.dataPointsCounter.markAsTouched();
      fixture.detectChanges();

      const error = fixture.debugElement
        .queryAll(By.css('mat-error'))
        .find((e) => e.nativeElement.textContent.includes('Almeno 1'));
      expect(error).toBeTruthy();
    });
  });

  // ─────────────────────────────────────────────
  // APPLY
  // ─────────────────────────────────────────────

  describe('apply', () => {
    it('should close dialog with complete ChartRequest', () => {
      const from = new Date('2025-01-01');
      const to = new Date('2025-01-02');

      component['filtersForm'].setValue({
        from,
        to,
        lowerBound: 50,
        upperBound: 200,
        dataPointsCounter: 250,
      });
      fixture.detectChanges();

      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      applyBtn.nativeElement.click();

      expect(dialogRefMock.close).toHaveBeenCalledWith({
        sensor: mockSensor,
        chartType: ChartType.HISTORIC,
        timeInterval: { from, to },
        valuesInterval: { lowerBound: 50, upperBound: 200 },
        dataPointsCounter: 250,
      });
    });

    it('should include correct sensor from dialog data', () => {
      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      applyBtn.nativeElement.click();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.sensor).toEqual(mockSensor);
      expect(result.chartType).toBe(ChartType.HISTORIC);
    });

    it('should include default values when form is not modified', () => {
      const applyBtn = fixture.debugElement.query(By.css('button[color="primary"]'));
      applyBtn.nativeElement.click();

      const result = dialogRefMock.close.mock.calls[0][0];
      expect(result.dataPointsCounter).toBe(100);
      expect(result.valuesInterval.lowerBound).toBe(0);
      expect(result.valuesInterval.upperBound).toBe(100);
      expect(result.timeInterval.from).toBeInstanceOf(Date);
      expect(result.timeInterval.to).toBeInstanceOf(Date);
    });

    it('should mark all as touched and not close when form is invalid', () => {
      component['filtersForm'].controls.from.setValue(null!);
      component['onApply']();

      expect(component['filtersForm'].controls.from.touched).toBe(true);
      expect(component['filtersForm'].controls.to.touched).toBe(true);
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
        .find((btn) => btn.nativeElement.textContent.includes('Annulla'));
      cancelBtn!.nativeElement.click();

      expect(dialogRefMock.close).toHaveBeenCalledWith();
    });
  });
});
