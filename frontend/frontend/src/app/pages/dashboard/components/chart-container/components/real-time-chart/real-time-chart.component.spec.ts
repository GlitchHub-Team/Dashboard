import { describe, it, expect, beforeEach } from 'vitest';
import { NO_ERRORS_SCHEMA, ComponentRef } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';

import { RealTimeChartComponent } from './real-time-chart.component';
import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';
import { Status } from '../../../../../../models/gateway-sensor-status.enum';

describe('RealTimeChartComponent', () => {
  let fixture: ComponentFixture<RealTimeChartComponent>;
  let component: RealTimeChartComponent;
  let componentRef: ComponentRef<RealTimeChartComponent>;

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 1000,
  };

  const mockReadings: SensorReading[] = [
    { timestamp: '2025-01-01T10:00:00Z', value: 72 },
    { timestamp: '2025-01-01T10:01:00Z', value: 75 },
    { timestamp: '2025-01-01T10:02:00Z', value: 70 },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RealTimeChartComponent],
    })
      .overrideComponent(RealTimeChartComponent, {
        set: {
          imports: [],
          schemas: [NO_ERRORS_SCHEMA],
        },
      })
      .compileComponents();

    fixture = TestBed.createComponent(RealTimeChartComponent);
    component = fixture.componentInstance;
    componentRef = fixture.componentRef;
  });

  function setInputs(sensor: Sensor, readings: SensorReading[]): void {
    componentRef.setInput('sensor', sensor);
    componentRef.setInput('readings', readings);
    fixture.detectChanges();
  }

  function query(selector: string): HTMLElement | null {
    return fixture.nativeElement.querySelector(selector);
  }

  describe('template', () => {
    it('should render a canvas element', () => {
      setInputs(mockSensor, mockReadings);
      expect(query('canvas')).not.toBeNull();
    });
  });

  describe('computed: profileDisplay', () => {
    it('should return correct display for HEART_RATE_SERVICE', () => {
      setInputs(mockSensor, mockReadings);
      expect(component['profileDisplay']()).toEqual({ label: 'Heart Rate', unit: 'bpm' });
    });

    it('should return correct display for PULSE_OXIMETER_SERVICE', () => {
      const sensor: Sensor = { ...mockSensor, profile: SensorProfiles.PULSE_OXIMETER_SERVICE };
      setInputs(sensor, mockReadings);
      expect(component['profileDisplay']()).toEqual({ label: 'Pulse Oximeter', unit: '%SpO₂' });
    });

    it('should return correct display for CUSTOM_ECG_SERVICE', () => {
      const sensor: Sensor = { ...mockSensor, profile: SensorProfiles.CUSTOM_ECG_SERVICE };
      setInputs(sensor, mockReadings);
      expect(component['profileDisplay']()).toEqual({ label: 'ECG', unit: 'mV' });
    });

    it('should return correct display for HEALTH_THERMOMETER_SERVICE', () => {
      const sensor: Sensor = { ...mockSensor, profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE };
      setInputs(sensor, mockReadings);
      expect(component['profileDisplay']()).toEqual({ label: 'Thermometer', unit: '°C' });
    });

    it('should return correct display for ENVIRONMENTAL_SENSING_SERVICE', () => {
      const sensor: Sensor = {
        ...mockSensor,
        profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
      };
      setInputs(sensor, mockReadings);
      expect(component['profileDisplay']()).toEqual({ label: 'Environmental', unit: '%' });
    });
  });

  describe('computed: chartData', () => {
    it('should map readings to chart labels using toLocaleTimeString', () => {
      setInputs(mockSensor, mockReadings);
      const data = component['chartData']();

      expect(data.labels).toHaveLength(3);
      expect(data.labels![0]).toBe(new Date('2025-01-01T10:00:00Z').toLocaleTimeString());
      expect(data.labels![1]).toBe(new Date('2025-01-01T10:01:00Z').toLocaleTimeString());
      expect(data.labels![2]).toBe(new Date('2025-01-01T10:02:00Z').toLocaleTimeString());
    });

    it('should map readings values to dataset data', () => {
      setInputs(mockSensor, mockReadings);
      const data = component['chartData']();

      expect(data.datasets).toHaveLength(1);
      expect(data.datasets[0].data).toEqual([72, 75, 70]);
    });

    it('should use profileDisplay label as dataset label', () => {
      setInputs(mockSensor, mockReadings);
      const data = component['chartData']();

      expect(data.datasets[0].label).toBe('Heart Rate');
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

      const newReadings: SensorReading[] = [
        ...mockReadings,
        { timestamp: '2025-01-01T10:03:00Z', value: 68 },
      ];
      componentRef.setInput('readings', newReadings);
      fixture.detectChanges();

      expect(component['chartData']().datasets[0].data).toEqual([72, 75, 70, 68]);
    });

    it('should update label when sensor profile changes', () => {
      setInputs(mockSensor, mockReadings);
      expect(component['chartData']().datasets[0].label).toBe('Heart Rate');

      const ecgSensor: Sensor = { ...mockSensor, profile: SensorProfiles.CUSTOM_ECG_SERVICE };
      componentRef.setInput('sensor', ecgSensor);
      fixture.detectChanges();

      expect(component['chartData']().datasets[0].label).toBe('ECG');
    });
  });

  describe('computed: chartOptions', () => {
    it('should set responsive to true', () => {
      setInputs(mockSensor, mockReadings);
      const options = component['chartOptions']();
      expect(options.responsive).toBe(true);
    });

    it('should set maintainAspectRatio to false', () => {
      setInputs(mockSensor, mockReadings);
      const options = component['chartOptions']();
      expect(options.maintainAspectRatio).toBe(false);
    });

    it('should disable animation', () => {
      setInputs(mockSensor, mockReadings);
      const options = component['chartOptions']();
      expect(options.animation).toBe(false);
    });

    it('should set x axis title to "Time"', () => {
      setInputs(mockSensor, mockReadings);
      const options = component['chartOptions']();
      const xScale = options.scales?.['x'] as any;
      expect(xScale.title.display).toBe(true);
      expect(xScale.title.text).toBe('Time');
    });

    it('should set y axis title with unit when unit exists', () => {
      setInputs(mockSensor, mockReadings);
      const options = component['chartOptions']();
      const yScale = options.scales?.['y'] as any;
      expect(yScale.title.display).toBe(true);
      expect(yScale.title.text).toBe('Value (bpm)');
    });

    it('should set y axis title without unit when unit is empty', () => {
      const unknownSensor: Sensor = {
        ...mockSensor,
        profile: 'unknown-profile' as SensorProfiles,
      };
      setInputs(unknownSensor, mockReadings);
      const options = component['chartOptions']();
      const yScale = options.scales?.['y'] as any;
      expect(yScale.title.text).toBe('Value');
    });

    it('should enable legend at top position', () => {
      setInputs(mockSensor, mockReadings);
      const options = component['chartOptions']();
      expect(options.plugins?.legend?.display).toBe(true);
      expect(options.plugins?.legend?.position).toBe('top');
    });

    it('should update y axis title when sensor profile changes', () => {
      setInputs(mockSensor, mockReadings);
      let options = component['chartOptions']();
      let yScale = options.scales?.['y'] as any;
      expect(yScale.title.text).toBe('Value (bpm)');

      const thermSensor: Sensor = {
        ...mockSensor,
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
      };
      componentRef.setInput('sensor', thermSensor);
      fixture.detectChanges();

      options = component['chartOptions']();
      yScale = options.scales?.['y'] as any;
      expect(yScale.title.text).toBe('Value (°C)');
    });
  });
});
