import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { describe, it, expect, beforeEach } from 'vitest';

import { RealTimeChartComponent } from './real-time-chart.component';
import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../../../../../models/sensor-data/field-descriptor.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';
import { SensorStatus } from '../../../../../../models/sensor-status.enum';

function createSensor(overrides: Partial<Sensor> = {}): Sensor {
  return {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Test Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: SensorStatus.ACTIVE,
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

describe('RealTimeChartComponent', () => {
  let fixture: ComponentFixture<RealTimeChartComponent>;
  let component: RealTimeChartComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [RealTimeChartComponent] }).compileComponents();
    fixture = TestBed.createComponent(RealTimeChartComponent);
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

  describe('chartData', () => {
    it('should return empty data when no field is selected', () => {
      setup([], createSensor(), []);
      const data = component['chartData']();
      expect(data.labels).toEqual([]);
      expect(data.datasets).toEqual([]);
    });

    it('should map readings to chart data for single-field sensor', () => {
      setup(createReadings(3, 'bpm', 70), createSensor(), SINGLE_FIELD);
      const data = component['chartData']();
      expect(data.datasets.length).toBe(1);
      expect(data.datasets[0].data).toEqual([70, 71, 72]);
      expect(data.datasets[0].label).toBe('Heart Rate');
      expect(data.labels!.length).toBe(3);
    });

    describe('multi-value sensor', () => {
      beforeEach(() =>
        setup(
          createMultiValueReadings(3, { spo2: 95, pulseRate: 70 }),
          createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
          MULTI_FIELDS,
        ),
      );

      it('should map readings to selected field', () => {
        const data = component['chartData']();
        expect(data.datasets[0].data).toEqual([95, 96, 97]);
        expect(data.datasets[0].label).toBe('Blood Oxygen');
      });

      it('should update chart data when field selection changes', () => {
        component['onFieldChange']('pulseRate');
        const data = component['chartData']();
        expect(data.datasets[0].data).toEqual([70, 71, 72]);
        expect(data.datasets[0].label).toBe('Pulse Rate');
      });
    });

    it('should use non-ECG styling for scalar sensors', () => {
      setup(createReadings(3, 'bpm', 70), createSensor(), SINGLE_FIELD);
      const dataset = component['chartData']().datasets[0];
      expect(dataset.borderColor).toBe('#4caf50');
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

      it('should show x-axis and disable animation', () => {
        const opts = component['chartOptions']();
        expect(opts.scales!['x']!['display']).toBe(true);
        expect(opts.animation).toBe(false);
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

    it('should render canvas with chart directive attributes', () => {
      setup(createReadings(3, 'bpm', 70), createSensor(), SINGLE_FIELD);
      expect(fixture.debugElement.query(By.css('canvas')).attributes['type']).toBe('line');
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

    it('should handle readings updating over time', () => {
      const initial = createReadings(3, 'bpm', 70);
      setup(initial, createSensor(), SINGLE_FIELD);
      expect(component['chartData']().datasets[0].data).toEqual([70, 71, 72]);

      fixture.componentRef.setInput('readings', [...initial, ...createReadings(2, 'bpm', 73)]);
      fixture.detectChanges();
      expect(component['chartData']().datasets[0].data).toEqual([70, 71, 72, 73, 74]);
    });

    it('should handle field change preserving readings', () => {
      setup(
        createMultiValueReadings(3, { spo2: 95, pulseRate: 70 }),
        createSensor({ profile: SensorProfiles.PULSE_OXIMETER_SERVICE }),
        MULTI_FIELDS,
      );
      expect(component['chartData']().datasets[0].data).toEqual([95, 96, 97]);
      component['onFieldChange']('pulseRate');
      expect(component['chartData']().datasets[0].data).toEqual([70, 71, 72]);
    });
  });
});
