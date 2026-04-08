import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { describe, it, expect, beforeEach } from 'vitest';

import { HistoricChartComponent } from './historic-chart.component';
import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../../../../../models/sensor-data/field-descriptor.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';
import { Status } from '../../../../../../models/gateway-sensor-status.enum';
import { SENSOR_VISIBLE_POINTS } from '../../../../../../models/chart/sensor-visible-points.model';

function createSensor(overrides: Partial<Sensor> = {}): Sensor {
  return {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Test Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 1000,
    ...overrides,
  };
}

function createReadings(count: number, fieldKey: string, baseValue: number): SensorReading[] {
  return Array.from({ length: count }, (_, i) => ({
    timestamp: new Date(Date.now() - (count - i) * 1000).toISOString(),
    value: { [fieldKey]: baseValue + i },
  }));
}

function createMultiValueReadings(count: number, fields: Record<string, number>): SensorReading[] {
  return Array.from({ length: count }, (_, i) => {
    const value: Record<string, number> = {};
    for (const [key, base] of Object.entries(fields)) value[key] = base + i;
    return { timestamp: new Date(Date.now() - (count - i) * 1000).toISOString(), value };
  });
}

const SINGLE_FIELD: FieldDescriptor[] = [{ key: 'bpm', label: 'Heart Rate', unit: 'bpm' }];
const MULTI_FIELDS: FieldDescriptor[] = [
  { key: 'spo2', label: 'Blood Oxygen', unit: '%' },
  { key: 'pulseRate', label: 'Pulse Rate', unit: 'bpm' },
];
const ECG_FIELDS: FieldDescriptor[] = [{ key: 'ecg', label: 'ECG Waveform', unit: 'mV' }];

const HR_VP = SENSOR_VISIBLE_POINTS[SensorProfiles.HEART_RATE_SERVICE] ?? 50;
const ECG_VP = SENSOR_VISIBLE_POINTS[SensorProfiles.CUSTOM_ECG_SERVICE] ?? 50;

describe('HistoricChartComponent', () => {
  let fixture: ComponentFixture<HistoricChartComponent>;
  let component: HistoricChartComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [HistoricChartComponent],
    }).compileComponents();
    fixture = TestBed.createComponent(HistoricChartComponent);
    component = fixture.componentInstance;
  });

  function setup(readings: SensorReading[], sensor: Sensor, fields: FieldDescriptor[]) {
    fixture.componentRef.setInput('readings', readings);
    fixture.componentRef.setInput('sensor', sensor);
    fixture.componentRef.setInput('fields', fields);
    fixture.detectChanges();
  }

  describe('creation', () => {
    it('should create the component', () => {
      setup([], createSensor(), SINGLE_FIELD);
      expect(component).toBeTruthy();
    });
  });

  describe('field selection', () => {
    it('should auto-select the first field on init', () => {
      setup(createReadings(5, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(component['selectedField']()).toBe('bpm');
    });

    describe('multi-field', () => {
      beforeEach(() =>
        setup(
          createMultiValueReadings(5, { spo2: 95, pulseRate: 70 }),
          createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
          MULTI_FIELDS,
        ),
      );

      it('should auto-select first field', () => {
        expect(component['selectedField']()).toBe('spo2');
      });

      it('should update selectedField when onFieldChange is called', () => {
        component['onFieldChange']('pulseRate');
        expect(component['selectedField']()).toBe('pulseRate');
      });

      it('should resolve the correct field descriptor after change', () => {
        component['onFieldChange']('pulseRate');
        expect(component['selectedFieldDescriptor']()).toEqual({
          key: 'pulseRate',
          label: 'Pulse Rate',
          unit: 'bpm',
        });
      });
    });
  });

  describe('hasMultipleFields', () => {
    it('should return false for single-field sensors', () => {
      setup(createReadings(5, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(component['hasMultipleFields']()).toBe(false);
    });

    it('should return true for multi-field sensors', () => {
      setup(
        createMultiValueReadings(5, { spo2: 95, pulseRate: 70 }),
        createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
        MULTI_FIELDS,
      );
      expect(component['hasMultipleFields']()).toBe(true);
    });
  });

  describe('visiblePoints', () => {
    it('should use configured visible points for heart rate', () => {
      setup(createReadings(5, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(component['visiblePoints']()).toBe(HR_VP);
    });

    it('should use configured visible points for ECG', () => {
      setup(
        createReadings(5, 'ecg', 100),
        createSensor({ profile: SensorProfiles.CUSTOM_ECG_SERVICE }),
        ECG_FIELDS,
      );
      expect(component['visiblePoints']()).toBe(ECG_VP);
    });

    it('should fallback to 50 for unknown sensor profile', () => {
      setup(
        createReadings(5, 'bpm', 70),
        createSensor({ profile: 'unknown_profile' as SensorProfiles }),
        SINGLE_FIELD,
      );
      expect(component['visiblePoints']()).toBe(50);
    });
  });

  describe('scrolling', () => {
    describe('with readings fitting within visible points', () => {
      beforeEach(() => setup(createReadings(10, 'bpm', 70), createSensor(), SINGLE_FIELD));

      it('should not allow scrolling', () => {
        expect(component['canScroll']()).toBe(false);
      });

      it('should have maxOffset of 0', () => {
        expect(component['maxOffset']()).toBe(0);
      });
    });

    describe('with readings exceeding visible points', () => {
      beforeEach(() => setup(createReadings(HR_VP + 50, 'bpm', 0), createSensor(), SINGLE_FIELD));

      it('should allow scrolling', () => {
        expect(component['canScroll']()).toBe(true);
      });

      it('should start with offset at 0', () => {
        expect(component['offset']()).toBe(0);
      });

      it('should scroll right by one step', () => {
        const step = component['scrollStep']();
        component['onScrollRight']();
        expect(component['offset']()).toBe(step);
      });

      it('should scroll left by one step', () => {
        const step = component['scrollStep']();
        component['onOffsetChange'](step * 2);
        component['onScrollLeft']();
        expect(component['offset']()).toBe(step);
      });

      it('should not scroll right beyond maxOffset', () => {
        for (let i = 0; i < 10; i++) component['onScrollRight']();
        expect(component['offset']()).toBeLessThanOrEqual(component['maxOffset']());
      });
    });

    it('should calculate maxOffset correctly', () => {
      const total = HR_VP + 30;
      setup(createReadings(total, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(component['maxOffset']()).toBe(total - HR_VP);
    });

    it('should calculate scrollStep as quarter of visible points', () => {
      setup(createReadings(5, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(component['scrollStep']()).toBe(
        Math.max(1, Math.floor(component['visiblePoints']() / 4)),
      );
    });

    it('should update offset via onOffsetChange', () => {
      setup(createReadings(100, 'bpm', 70), createSensor(), SINGLE_FIELD);
      component['onOffsetChange'](25);
      expect(component['offset']()).toBe(25);
    });

    it('should not scroll left below 0', () => {
      setup(createReadings(100, 'bpm', 70), createSensor(), SINGLE_FIELD);
      component['onScrollLeft']();
      expect(component['offset']()).toBe(0);
    });

    it('should slice readings according to offset and visible points', () => {
      setup(createReadings(HR_VP + 20, 'bpm', 0), createSensor(), SINGLE_FIELD);

      expect(component['visibleReadings']().length).toBe(HR_VP);
      expect(component['visibleReadings']()[0].value['bpm']).toBe(0);

      component['onOffsetChange'](10);
      expect(component['visibleReadings']().length).toBe(HR_VP);
      expect(component['visibleReadings']()[0].value['bpm']).toBe(10);
    });
  });

  describe('chartData', () => {
    it('should return empty data when no field is selected', () => {
      setup([], createSensor(), []);
      const data = component['chartData']();
      expect(data.labels).toEqual([]);
      expect(data.datasets).toEqual([]);
    });

    it('should map visible readings to chart data', () => {
      setup(createReadings(10, 'bpm', 70), createSensor(), SINGLE_FIELD);
      const data = component['chartData']();
      expect(data.datasets.length).toBe(1);
      expect(data.datasets[0].label).toBe('Heart Rate');
      expect(data.datasets[0].data).toEqual([70, 71, 72, 73, 74, 75, 76, 77, 78, 79]);
    });

    it('should only include visible window in chart data', () => {
      setup(createReadings(HR_VP + 20, 'bpm', 0), createSensor(), SINGLE_FIELD);
      const data = component['chartData']();
      expect((data.datasets[0].data as number[]).length).toBe(HR_VP);
      expect((data.datasets[0].data as number[])[0]).toBe(0);
    });

    it('should reflect offset change in chart data', () => {
      setup(createReadings(HR_VP + 20, 'bpm', 0), createSensor(), SINGLE_FIELD);
      component['onOffsetChange'](10);
      expect((component['chartData']().datasets[0].data as number[])[0]).toBe(10);
    });

    it('should map selected field for multi-value sensor', () => {
      setup(
        createMultiValueReadings(5, { spo2: 95, pulseRate: 70 }),
        createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
        MULTI_FIELDS,
      );
      expect(component['chartData']().datasets[0].data).toEqual([95, 96, 97, 98, 99]);
      component['onFieldChange']('pulseRate');
      expect(component['chartData']().datasets[0].data).toEqual([70, 71, 72, 73, 74]);
    });

    it('should use non-ECG styling for scalar sensors', () => {
      setup(createReadings(3, 'bpm', 70), createSensor(), SINGLE_FIELD);
      const dataset = component['chartData']().datasets[0];
      expect(dataset.borderColor).toBe('#3f51b5');
      expect(dataset.fill).toBe(true);
      expect(dataset.tension).toBe(0.3);
      expect(dataset.pointRadius).toBe(2);
      expect(dataset.borderWidth).toBe(2);
    });

    it('should use ECG styling for ECG sensor', () => {
      setup(
        createReadings(3, 'ecg', 100),
        createSensor({ profile: SensorProfiles.CUSTOM_ECG_SERVICE }),
        ECG_FIELDS,
      );
      const dataset = component['chartData']().datasets[0];
      expect(dataset.borderColor).toBe('#00ff88');
      expect(dataset.fill).toBe(false);
      expect(dataset.tension).toBe(0.2);
      expect(dataset.pointRadius).toBe(0);
      expect(dataset.borderWidth).toBe(1.5);
    });
  });

  describe('chartOptions', () => {
    describe('scalar sensor', () => {
      beforeEach(() => setup(createReadings(3, 'bpm', 70), createSensor(), SINGLE_FIELD));

      it('should show x-axis, enable animation, and enable tooltip', () => {
        const opts = component['chartOptions']();
        expect(opts.scales!['x']!['display']).toBe(true);
        expect(opts.animation).toEqual({ duration: 300 });
        expect(opts.plugins!.tooltip!['enabled']).toBe(true);
      });

      it('should set y-axis title from selected field', () => {
        expect((component['chartOptions']().scales!['y'] as any).title.text).toBe(
          'Heart Rate (bpm)',
        );
      });
    });

    it('should update y-axis title when field changes', () => {
      setup(
        createMultiValueReadings(3, { spo2: 95, pulseRate: 70 }),
        createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
        MULTI_FIELDS,
      );
      component['onFieldChange']('pulseRate');
      expect((component['chartOptions']().scales!['y'] as any).title.text).toBe('Pulse Rate (bpm)');
    });

    it('should fallback y-axis title when no field selected', () => {
      setup([], createSensor(), []);
      expect((component['chartOptions']().scales!['y'] as any).title.text).toBe('Value');
    });
  });

  describe('template - dropdown', () => {
    it('should not render dropdown for single-field sensor', () => {
      setup(createReadings(5, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(fixture.debugElement.query(By.css('mat-select'))).toBeNull();
    });

    describe('multi-field sensor', () => {
      beforeEach(() =>
        setup(
          createMultiValueReadings(5, { spo2: 95, pulseRate: 70 }),
          createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
          MULTI_FIELDS,
        ),
      );

      it('should render dropdown', () => {
        expect(fixture.debugElement.query(By.css('mat-select'))).not.toBeNull();
      });

      it('should render correct number of options', () => {
        fixture.debugElement.query(By.css('.mat-mdc-select-trigger')).nativeElement.click();
        fixture.detectChanges();
        expect(fixture.debugElement.queryAll(By.css('mat-option')).length).toBe(2);
      });

      it('should display field labels with units in options', () => {
        fixture.debugElement.query(By.css('.mat-mdc-select-trigger')).nativeElement.click();
        fixture.detectChanges();
        const texts = fixture.debugElement
          .queryAll(By.css('mat-option'))
          .map((o) => o.nativeElement.textContent.trim());
        expect(texts).toContain('Blood Oxygen (%)');
        expect(texts).toContain('Pulse Rate (bpm)');
      });
    });
  });

  describe('template - canvas', () => {
    it('should always render the chart canvas', () => {
      setup([], createSensor(), SINGLE_FIELD);
      expect(fixture.debugElement.query(By.css('canvas'))).not.toBeNull();
    });

    it('should render canvas with line chart type', () => {
      setup(createReadings(3, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(fixture.debugElement.query(By.css('canvas')).attributes['type']).toBe('line');
    });
  });

  describe('template - scroll controls', () => {
    it('should not render scroll controls when readings fit', () => {
      setup(createReadings(10, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeNull();
    });

    describe('when readings exceed visible points', () => {
      beforeEach(() => setup(createReadings(HR_VP + 20, 'bpm', 70), createSensor(), SINGLE_FIELD));

      it('should render scroll controls', () => {
        expect(fixture.debugElement.query(By.css('.scroll-controls'))).not.toBeNull();
      });

      it('should render left and right scroll buttons', () => {
        expect(
          fixture.debugElement.queryAll(By.css('.scroll-controls button[mat-icon-button]')).length,
        ).toBe(2);
      });

      it('should render the slider', () => {
        expect(fixture.debugElement.query(By.css('mat-slider'))).not.toBeNull();
      });
    });

    it('should scroll right when right button is clicked', () => {
      setup(createReadings(HR_VP + 50, 'bpm', 0), createSensor(), SINGLE_FIELD);
      fixture.debugElement
        .queryAll(By.css('.scroll-controls button[mat-icon-button]'))[1]
        .nativeElement.click();
      fixture.detectChanges();
      expect(component['offset']()).toBeGreaterThan(0);
    });

    it('should scroll left when left button is clicked', () => {
      setup(createReadings(HR_VP + 50, 'bpm', 0), createSensor(), SINGLE_FIELD);
      component['onOffsetChange'](20);
      fixture.detectChanges();
      fixture.debugElement
        .queryAll(By.css('.scroll-controls button[mat-icon-button]'))[0]
        .nativeElement.click();
      fixture.detectChanges();
      expect(component['offset']()).toBeLessThan(20);
    });
  });

  describe('edge cases', () => {
    it('should handle empty readings array', () => {
      setup([], createSensor(), SINGLE_FIELD);
      const data = component['chartData']();
      expect(data.datasets[0].data).toEqual([]);
      expect(data.labels).toEqual([]);
    });

    it('should handle single reading', () => {
      setup(createReadings(1, 'bpm', 72), createSensor(), SINGLE_FIELD);
      const data = component['chartData']();
      expect(data.datasets[0].data).toEqual([72]);
      expect(data.labels!.length).toBe(1);
    });

    it('should handle readings exactly equal to visible points', () => {
      setup(createReadings(HR_VP, 'bpm', 0), createSensor(), SINGLE_FIELD);
      expect(component['canScroll']()).toBe(false);
      expect(component['maxOffset']()).toBe(0);
      expect((component['chartData']().datasets[0].data as number[]).length).toBe(HR_VP);
    });

    it('should handle readings one more than visible points', () => {
      setup(createReadings(HR_VP + 1, 'bpm', 0), createSensor(), SINGLE_FIELD);
      expect(component['canScroll']()).toBe(true);
      expect(component['maxOffset']()).toBe(1);
    });

    it('should handle field change preserving offset', () => {
      const vp = SENSOR_VISIBLE_POINTS[SensorProfiles.PULSE_OXIMETER_SERVICE] ?? 50;
      setup(
        createMultiValueReadings(vp + 20, { spo2: 80, pulseRate: 60 }),
        createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
        MULTI_FIELDS,
      );
      component['onOffsetChange'](10);
      component['onFieldChange']('pulseRate');
      expect(component['offset']()).toBe(10);
      expect((component['chartData']().datasets[0].data as number[])[0]).toBe(70);
    });

    it('should handle rapid readings input changes', () => {
      setup(createReadings(10, 'bpm', 70), createSensor(), SINGLE_FIELD);
      fixture.componentRef.setInput('readings', createReadings(20, 'bpm', 50));
      fixture.detectChanges();
      fixture.componentRef.setInput('readings', createReadings(30, 'bpm', 30));
      fixture.detectChanges();
      const data = component['chartData']();
      expect((data.datasets[0].data as number[])[0]).toBe(30);
      expect(data.datasets[0].data.length).toBe(30);
    });
  });
});
