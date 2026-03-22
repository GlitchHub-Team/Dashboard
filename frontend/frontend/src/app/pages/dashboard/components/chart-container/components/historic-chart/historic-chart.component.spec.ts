import { describe, it, expect, beforeEach } from 'vitest';
import { ComponentRef, Directive, input } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { ChartData, ChartOptions } from 'chart.js';

import { HistoricChartComponent } from './historic-chart.component';
import { BaseChartDirective } from 'ng2-charts';
import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';
import { Status } from '../../../../../../models/gateway-sensor-status.enum';

@Directive({ selector: 'canvas[baseChart]', standalone: true })
class StubBaseChart {
  type = input<string>();
  data = input<ChartData<'line'>>();
  options = input<ChartOptions<'line'>>();
}

describe('HistoricChartComponent (Unit)', () => {
  let fixture: ComponentFixture<HistoricChartComponent>;
  let component: HistoricChartComponent;
  let componentRef: ComponentRef<HistoricChartComponent>;

  const mockSensor: Sensor = { id: 'sensor-1', gatewayId: 'gw-1', name: 'Heart Rate Sensor', profile: SensorProfiles.HEART_RATE_SERVICE, status: Status.ACTIVE, dataInterval: 1000 };

  const generateReadings = (count: number): SensorReading[] =>
    Array.from({ length: count }, (_, i) => ({ timestamp: new Date(2025, 0, 1, 10, i).toISOString(), value: 70 + (i % 10) }));

  const smallReadings = generateReadings(10);
  const exactReadings = generateReadings(50);
  const largeReadings = generateReadings(100);

  const setInputs = (sensor: Sensor, readings: SensorReading[]) => {
    componentRef.setInput('sensor', sensor);
    componentRef.setInput('readings', readings);
    fixture.detectChanges();
  };

  const getStubChart = () =>
    fixture.debugElement.query(By.directive(StubBaseChart)).injector.get(StubBaseChart);

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [HistoricChartComponent] })
      .overrideComponent(HistoricChartComponent, {
        remove: { imports: [BaseChartDirective] },
        add: { imports: [StubBaseChart] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(HistoricChartComponent);
    component = fixture.componentInstance;
    componentRef = fixture.componentRef;
  });

  describe('template', () => {
    it('should render canvas and hide scroll controls when readings fit in VISIBLE_POINTS', () => {
      setInputs(mockSensor, smallReadings);
      expect(fixture.debugElement.query(By.directive(StubBaseChart))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeNull();
    });

    it('should hide scroll controls when readings equal VISIBLE_POINTS', () => {
      setInputs(mockSensor, exactReadings);
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeNull();
    });

    it('should render scroll controls with two buttons and a slider when readings exceed VISIBLE_POINTS', () => {
      setInputs(mockSensor, largeReadings);
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeTruthy();
      expect(fixture.debugElement.queryAll(By.css('button[mat-icon-button]'))).toHaveLength(2);
      expect(fixture.debugElement.query(By.css('mat-slider'))).toBeTruthy();
    });

    it('should pass chartData and chartOptions to canvas directive', () => {
      setInputs(mockSensor, smallReadings);
      const stub = getStubChart();
      expect(stub.data()?.datasets[0].data).toEqual(smallReadings.map((r) => r.value));
      expect(stub.options()?.responsive).toBe(true);
    });

    it('should update canvas data when offset changes', () => {
      setInputs(mockSensor, largeReadings);
      const dataBefore = getStubChart().data()?.datasets[0].data;
      component['offset'].set(25);
      fixture.detectChanges();
      expect(getStubChart().data()?.datasets[0].data).not.toEqual(dataBefore);
    });
  });

  describe('computed: maxOffset and canScroll', () => {
    it.each<[string, SensorReading[], number, boolean]>([
      ['small (10)', smallReadings, 0, false],
      ['exact (50)', exactReadings, 0, false],
      ['large (100)', largeReadings, 50, true],
      ['empty', [], 0, false],
    ])('readings=%s → maxOffset=%i, canScroll=%s', (_, readings, maxOff, canScr) => {
      setInputs(mockSensor, readings);
      expect(component['maxOffset']()).toBe(maxOff);
      expect(component['canScroll']()).toBe(canScr);
    });
  });

  describe('computed: scrollStep', () => {
    it('should return floor(VISIBLE_POINTS / 4) = 12', () => {
      setInputs(mockSensor, smallReadings);
      expect(component['scrollStep']()).toBe(12);
    });
  });

  describe('computed: visibleReadings', () => {
    it('should return all readings when fewer than VISIBLE_POINTS, or empty for empty input', () => {
      setInputs(mockSensor, smallReadings);
      expect(component['visibleReadings']()).toEqual(smallReadings);

      setInputs(mockSensor, []);
      expect(component['visibleReadings']()).toEqual([]);
    });

    it.each<[number, number, number]>([
      [0, 0, 49],
      [25, 25, 74],
      [50, 50, 99],
    ])('offset=%i → window largeReadings[%i..%i]', (offset, start, end) => {
      setInputs(mockSensor, largeReadings);
      component['offset'].set(offset);
      const visible = component['visibleReadings']();
      expect(visible).toHaveLength(50);
      expect(visible[0]).toEqual(largeReadings[start]);
      expect(visible[49]).toEqual(largeReadings[end]);
    });
  });

  describe('computed: profileDisplay', () => {
    it.each<[SensorProfiles, { label: string; unit: string }]>([
      [SensorProfiles.HEART_RATE_SERVICE, { label: 'Heart Rate', unit: 'bpm' }],
      [SensorProfiles.CUSTOM_ECG_SERVICE, { label: 'ECG', unit: 'mV' }],
      [SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE, { label: 'Environmental', unit: '%' }],
    ])('profile %s → %o', (profile, expected) => {
      setInputs({ ...mockSensor, profile }, smallReadings);
      expect(component['profileDisplay']()).toEqual(expected);
    });
  });

  describe('computed: chartData', () => {
    it('should map visible readings to labels, dataset values, and correct styling', () => {
      setInputs(mockSensor, smallReadings);
      const data = component['chartData']();
      const ds = data.datasets[0];
      expect(data.labels).toHaveLength(10);
      expect(data.labels![0]).toBe(new Date(smallReadings[0].timestamp).toLocaleTimeString());
      expect(ds.data).toEqual(smallReadings.map((r) => r.value));
      expect(ds.label).toBe('Heart Rate');
      expect(ds.borderColor).toBe('#3f51b5');
      expect(ds.backgroundColor).toBe('rgba(63, 81, 181, 0.1)');
      expect(ds.fill).toBe(true);
      expect(ds.tension).toBe(0.3);
      expect(ds.pointRadius).toBe(2);
    });

    it('should handle empty readings', () => {
      setInputs(mockSensor, []);
      const data = component['chartData']();
      expect(data.labels).toHaveLength(0);
      expect(data.datasets[0].data).toHaveLength(0);
    });

    it('should only include VISIBLE_POINTS readings and update labels when offset changes', () => {
      setInputs(mockSensor, largeReadings);
      const data = component['chartData']();
      expect(data.labels).toHaveLength(50);

      const firstLabelBefore = data.labels![0];
      component['offset'].set(25);
      expect(component['chartData']().labels![0]).not.toBe(firstLabelBefore);
    });
  });

  describe('computed: chartOptions', () => {
    it('should set responsive, aspect ratio, axes titles, and legend', () => {
      setInputs(mockSensor, smallReadings);
      const options = component['chartOptions']();
      const xScale = options.scales?.['x'] as any;
      const yScale = options.scales?.['y'] as any;
      expect(options.responsive).toBe(true);
      expect(options.maintainAspectRatio).toBe(false);
      expect(xScale.title.display).toBe(true);
      expect(xScale.title.text).toBe('Time');
      expect(yScale.title.display).toBe(true);
      expect(yScale.title.text).toBe('Value (bpm)');
      expect(options.plugins?.legend?.display).toBe(true);
      expect(options.plugins?.legend?.position).toBe('top');
    });

    it('should set y axis title without unit for unknown profile', () => {
      setInputs({ ...mockSensor, profile: 'unknown-profile' as SensorProfiles }, smallReadings);
      expect((component['chartOptions']().scales?.['y'] as any).title.text).toBe('Value');
    });
  });

  describe('onOffsetChange / onScrollLeft / onScrollRight', () => {
    beforeEach(() => setInputs(mockSensor, largeReadings));

    it('onOffsetChange sets offset to the given value', () => {
      component['onOffsetChange'](30);
      expect(component['offset']()).toBe(30);

      component['onOffsetChange'](0);
      expect(component['offset']()).toBe(0);
    });

    it.each<[number, number]>([
      [30, 18],
      [5, 0],
      [0, 0],
    ])('onScrollLeft: offset %i → %i', (initial, expected) => {
      component['offset'].set(initial);
      component['onScrollLeft']();
      expect(component['offset']()).toBe(expected);
    });

    it.each<[number, number]>([
      [0, 12],
      [45, 50],
      [50, 50],
    ])('onScrollRight: offset %i → %i', (initial, expected) => {
      component['offset'].set(initial);
      component['onScrollRight']();
      expect(component['offset']()).toBe(expected);
    });
  });

  describe('scroll controls interaction', () => {
    it('should scroll left and right when buttons are clicked', () => {
      setInputs(mockSensor, largeReadings);
      component['offset'].set(30);
      fixture.detectChanges();

      const [leftBtn, rightBtn] = fixture.debugElement.queryAll(By.css('button[mat-icon-button]')).map((b) => b.nativeElement);

      leftBtn.click();
      fixture.detectChanges();
      expect(component['offset']()).toBe(18);

      component['offset'].set(0);
      fixture.detectChanges();
      rightBtn.click();
      fixture.detectChanges();
      expect(component['offset']()).toBe(12);
    });
  });

  describe('signal reactivity', () => {
    it('should show and hide scroll controls as readings grow and shrink', () => {
      setInputs(mockSensor, smallReadings);
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeNull();

      componentRef.setInput('readings', largeReadings);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeTruthy();

      componentRef.setInput('readings', smallReadings);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.scroll-controls'))).toBeNull();
    });

    it('should handle visibleReadings without error when maxOffset decreases', () => {
      setInputs(mockSensor, largeReadings);
      component['offset'].set(50);

      componentRef.setInput('readings', generateReadings(60));
      fixture.detectChanges();
      expect(component['visibleReadings']().length).toBeLessThanOrEqual(50);
    });
  });
});
