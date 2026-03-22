import { describe, it, expect, beforeEach } from 'vitest';
import { ComponentRef, Directive, input } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { RealTimeChartComponent } from './real-time-chart.component';
import { BaseChartDirective } from 'ng2-charts';
import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';
import { Status } from '../../../../../../models/gateway-sensor-status.enum';
import { ChartData, ChartOptions } from 'chart.js';

@Directive({ selector: 'canvas[baseChart]', standalone: true })
class StubBaseChart {
  type = input<string>();
  data = input<ChartData<'line'>>();
  options = input<ChartOptions<'line'>>();
}

describe('RealTimeChartComponent (Unit)', () => {
  let fixture: ComponentFixture<RealTimeChartComponent>;
  let component: RealTimeChartComponent;
  let componentRef: ComponentRef<RealTimeChartComponent>;

  const mockSensor: Sensor = { id: 'sensor-1', gatewayId: 'gw-1', name: 'Heart Rate Sensor', profile: SensorProfiles.HEART_RATE_SERVICE, status: Status.ACTIVE, dataInterval: 1000 };
  const mockReadings: SensorReading[] = [
    { timestamp: '2025-01-01T10:00:00Z', value: 72 },
    { timestamp: '2025-01-01T10:01:00Z', value: 75 },
    { timestamp: '2025-01-01T10:02:00Z', value: 70 },
  ];

  const setInputs = (sensor: Sensor, readings: SensorReading[]) => {
    componentRef.setInput('sensor', sensor);
    componentRef.setInput('readings', readings);
    fixture.detectChanges();
  };

  const getStubChart = () =>
    fixture.debugElement.query(By.directive(StubBaseChart)).injector.get(StubBaseChart);

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [RealTimeChartComponent] })
      .overrideComponent(RealTimeChartComponent, {
        remove: { imports: [BaseChartDirective] },
        add: { imports: [StubBaseChart] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(RealTimeChartComponent);
    component = fixture.componentInstance;
    componentRef = fixture.componentRef;
  });

  describe('template', () => {
    it('should render a canvas element', () => {
      setInputs(mockSensor, mockReadings);
      expect(fixture.nativeElement.querySelector('canvas')).toBeTruthy();
    });

    it('should pass chartData and chartOptions to canvas', () => {
      setInputs(mockSensor, mockReadings);
      const stub = getStubChart();
      expect(stub.data()?.datasets[0].data).toEqual([72, 75, 70]);
      expect(stub.options()?.responsive).toBe(true);
    });
  });

  describe('computed: profileDisplay', () => {
    it.each([
      [SensorProfiles.HEART_RATE_SERVICE, { label: 'Heart Rate', unit: 'bpm' }],
      [SensorProfiles.PULSE_OXIMETER_SERVICE, { label: 'Pulse Oximeter', unit: '%SpO₂' }],
      [SensorProfiles.CUSTOM_ECG_SERVICE, { label: 'ECG', unit: 'mV' }],
      [SensorProfiles.HEALTH_THERMOMETER_SERVICE, { label: 'Thermometer', unit: '°C' }],
      [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE, { label: 'Environmental', unit: '%' }],
    ] as const)('should return correct display for %s', (profile, expected) => {
      setInputs({ ...mockSensor, profile }, mockReadings);
      expect(component['profileDisplay']()).toEqual(expected);
    });
  });

  describe('computed: chartData', () => {
    it('should map readings to chart labels using toLocaleTimeString', () => {
      setInputs(mockSensor, mockReadings);
      const labels = component['chartData']().labels!;
      expect(labels).toHaveLength(3);
      expect(labels[0]).toBe(new Date('2025-01-01T10:00:00Z').toLocaleTimeString());
      expect(labels[1]).toBe(new Date('2025-01-01T10:01:00Z').toLocaleTimeString());
      expect(labels[2]).toBe(new Date('2025-01-01T10:02:00Z').toLocaleTimeString());
    });

    it('should map readings values to dataset data', () => {
      setInputs(mockSensor, mockReadings);
      const { datasets } = component['chartData']();
      expect(datasets).toHaveLength(1);
      expect(datasets[0].data).toEqual([72, 75, 70]);
    });

    it('should use profileDisplay label as dataset label', () => {
      setInputs(mockSensor, mockReadings);
      expect(component['chartData']().datasets[0].label).toBe('Heart Rate');
    });

    it('should set correct chart styling properties', () => {
      setInputs(mockSensor, mockReadings);
      const dataset = component['chartData']().datasets[0];
      expect(dataset.borderColor).toBe('#4caf50');
      expect(dataset.backgroundColor).toBe('rgba(76, 175, 80, 0.1)');
      expect(dataset.fill).toBe(true);
      expect(dataset.tension).toBe(0.3);
      expect(dataset.pointRadius).toBe(2);
    });

    it('should handle empty readings', () => {
      setInputs(mockSensor, []);
      const data = component['chartData']();
      expect(data.labels).toHaveLength(0);
      expect(data.datasets[0].data).toHaveLength(0);
    });

    it('should handle single reading', () => {
      setInputs(mockSensor, [{ timestamp: '2025-01-01T10:00:00Z', value: 80 }]);
      const data = component['chartData']();
      expect(data.labels).toHaveLength(1);
      expect(data.datasets[0].data).toEqual([80]);
    });

    it('should update when readings change', () => {
      setInputs(mockSensor, mockReadings);
      expect(component['chartData']().datasets[0].data).toEqual([72, 75, 70]);

      componentRef.setInput('readings', [...mockReadings, { timestamp: '2025-01-01T10:03:00Z', value: 68 }]);
      fixture.detectChanges();
      expect(component['chartData']().datasets[0].data).toEqual([72, 75, 70, 68]);
    });

    it('should update label when sensor profile changes', () => {
      setInputs(mockSensor, mockReadings);
      expect(component['chartData']().datasets[0].label).toBe('Heart Rate');

      componentRef.setInput('sensor', { ...mockSensor, profile: SensorProfiles.CUSTOM_ECG_SERVICE });
      fixture.detectChanges();
      expect(component['chartData']().datasets[0].label).toBe('ECG');
    });
  });

  describe('computed: chartOptions', () => {
    it('should set responsive, maintainAspectRatio, and animation', () => {
      setInputs(mockSensor, mockReadings);
      const opts = component['chartOptions']();
      expect(opts.responsive).toBe(true);
      expect(opts.maintainAspectRatio).toBe(false);
      expect(opts.animation).toBe(false);
    });

    it('should set x axis title to "Time"', () => {
      setInputs(mockSensor, mockReadings);
      const xScale = component['chartOptions']().scales?.['x'] as any;
      expect(xScale.title.display).toBe(true);
      expect(xScale.title.text).toBe('Time');
    });

    it('should set y axis title with unit', () => {
      setInputs(mockSensor, mockReadings);
      const yScale = component['chartOptions']().scales?.['y'] as any;
      expect(yScale.title.display).toBe(true);
      expect(yScale.title.text).toBe('Value (bpm)');
    });

    it('should set y axis title without unit for unknown profile', () => {
      setInputs({ ...mockSensor, profile: 'unknown-profile' as SensorProfiles }, mockReadings);
      expect((component['chartOptions']().scales?.['y'] as any).title.text).toBe('Value');
    });

    it('should enable legend at top position', () => {
      setInputs(mockSensor, mockReadings);
      const opts = component['chartOptions']();
      expect(opts.plugins?.legend?.display).toBe(true);
      expect(opts.plugins?.legend?.position).toBe('top');
    });

    it('should update y axis title when sensor profile changes', () => {
      setInputs(mockSensor, mockReadings);
      expect((component['chartOptions']().scales?.['y'] as any).title.text).toBe('Value (bpm)');

      componentRef.setInput('sensor', { ...mockSensor, profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE });
      fixture.detectChanges();
      expect((component['chartOptions']().scales?.['y'] as any).title.text).toBe('Value (°C)');
    });
  });
});
