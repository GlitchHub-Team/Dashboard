import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, ComponentRef, Component, input, output } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { ChartContainerComponent } from './chart-container.component';
import { HistoricChartComponent } from './components/historic-chart/historic-chart.component';
import { RealTimeChartComponent } from './components/real-time-chart/real-time-chart.component';
import { SensorChartService } from '../../../../services/sensor-chart/sensor-chart.service';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { SensorReading } from '../../../../models/sensor-data/sensor-reading.model';
import { Status } from '../../../../models/gateway-sensor-status.enum';

@Component({ selector: 'app-historic-chart', template: '', standalone: true })
class StubHistoricChart {
  readings = input<SensorReading[]>();
  sensor = input<Sensor>();
}

@Component({ selector: 'app-real-time-chart', template: '', standalone: true })
class StubRealTimeChart {
  readings = input<SensorReading[]>();
  sensor = input<Sensor>();
}

describe('ChartContainerComponent (Unit)', () => {
  let fixture: ComponentFixture<ChartContainerComponent>;
  let component: ChartContainerComponent;
  let componentRef: ComponentRef<ChartContainerComponent>;

  let loadingSig: WritableSignal<boolean>;
  let connectionStatusSig: WritableSignal<'connected' | 'connecting' | 'disconnected'>;
  let errorSig: WritableSignal<string | null>;
  let historicReadingsSig: WritableSignal<SensorReading[]>;
  let liveReadingsSig: WritableSignal<SensorReading[]>;

  let chartServiceMock: {
    historicReadings: WritableSignal<SensorReading[]>;
    liveReadings: WritableSignal<SensorReading[]>;
    loading: WritableSignal<boolean>;
    connectionStatus: WritableSignal<'connected' | 'connecting' | 'disconnected'>;
    error: WritableSignal<string | null>;
    startChart: ReturnType<typeof vi.fn>;
    stopChart: ReturnType<typeof vi.fn>;
  };

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 1000,
  };
  const mockHistoricRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.HISTORIC,
    timeInterval: { from: new Date('2025-01-01'), to: new Date('2025-01-02') },
  };
  const mockRealtimeRequest: ChartRequest = { sensor: mockSensor, chartType: ChartType.REALTIME };
  const mockReadings: SensorReading[] = [
    { timestamp: '2025-01-01T10:00:00Z', value: 72 } as SensorReading,
    { timestamp: '2025-01-01T10:01:00Z', value: 75 } as SensorReading,
  ];

  const setChartRequest = (req: ChartRequest | null) => {
    componentRef.setInput('chartRequest', req);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    vi.resetAllMocks();

    loadingSig = signal(false);
    connectionStatusSig = signal<'connected' | 'connecting' | 'disconnected'>('disconnected');
    errorSig = signal<string | null>(null);
    historicReadingsSig = signal<SensorReading[]>([]);
    liveReadingsSig = signal<SensorReading[]>([]);

    chartServiceMock = {
      historicReadings: historicReadingsSig,
      liveReadings: liveReadingsSig,
      loading: loadingSig,
      connectionStatus: connectionStatusSig,
      error: errorSig,
      startChart: vi.fn(),
      stopChart: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [ChartContainerComponent],
      providers: [{ provide: SensorChartService, useValue: chartServiceMock }],
    })
      .overrideComponent(ChartContainerComponent, {
        remove: { imports: [HistoricChartComponent, RealTimeChartComponent] },
        add: { imports: [StubHistoricChart, StubRealTimeChart] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(ChartContainerComponent);
    component = fixture.componentInstance;
    componentRef = fixture.componentRef;
  });

  describe('when chartRequest is null', () => {
    beforeEach(() => setChartRequest(null));

    it('should show placeholder and hide mat-card', () => {
      const placeholder = fixture.debugElement.query(By.css('.chart-placeholder'));
      expect(placeholder).toBeTruthy();
      expect(placeholder.nativeElement.textContent).toContain('No chart selected');
      expect(fixture.debugElement.query(By.css('mat-card'))).toBeNull();
    });

    it('should not call startChart', () => {
      expect(chartServiceMock.startChart).not.toHaveBeenCalled();
    });
  });

  describe('effect: startChart', () => {
    it('should call startChart for each request type and update on change', () => {
      setChartRequest(mockHistoricRequest);
      expect(chartServiceMock.startChart).toHaveBeenCalledWith(mockHistoricRequest);

      setChartRequest(mockRealtimeRequest);
      expect(chartServiceMock.startChart).toHaveBeenCalledTimes(2);
      expect(chartServiceMock.startChart).toHaveBeenLastCalledWith(mockRealtimeRequest);
    });
  });

  describe('card header', () => {
    it('should display sensor name and profile label for heart-rate sensor', () => {
      setChartRequest(mockHistoricRequest);
      const title = fixture.debugElement.query(By.css('mat-card-title'));
      expect(title.nativeElement.textContent).toContain('Heart Rate Sensor');
      expect(title.nativeElement.textContent).toContain('Heart Rate');
    });

    it('should display correct profile label for ECG sensor', () => {
      setChartRequest({
        sensor: { ...mockSensor, name: 'ECG Sensor', profile: SensorProfiles.CUSTOM_ECG_SERVICE },
        chartType: ChartType.HISTORIC,
        timeInterval: { from: new Date(), to: new Date() },
      });
      const title = fixture.debugElement.query(By.css('mat-card-title'));
      expect(title.nativeElement.textContent).toContain('ECG Sensor');
      expect(title.nativeElement.textContent).toContain('ECG');
    });
  });

  describe('close button', () => {
    it('should call stopChart and emit chartClosed when clicked', () => {
      setChartRequest(mockHistoricRequest);
      const closedSpy = vi.fn();
      component.chartClosed.subscribe(closedSpy);

      fixture.debugElement.query(By.css('button[mat-icon-button]')).nativeElement.click();

      expect(chartServiceMock.stopChart).toHaveBeenCalled();
      expect(closedSpy).toHaveBeenCalledOnce();
    });
  });

  describe('connection status', () => {
    it('should not show connection status for historic chart', () => {
      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.css('.connection-status'))).toBeNull();
    });

    it.each([
      ['connected', 'Connected', 'status-connected'],
      ['connecting', 'Connecting...', 'status-connecting'],
      ['disconnected', 'Disconnected', 'status-disconnected'],
    ] as const)('should show "%s" label and class for realtime chart', (status, label, cls) => {
      setChartRequest(mockRealtimeRequest);
      connectionStatusSig.set(status);
      fixture.detectChanges();

      const el = fixture.debugElement.query(By.css('.connection-status'));
      expect(el).toBeTruthy();
      expect(el.nativeElement.textContent).toContain(label);
      expect(el.nativeElement.classList.contains(cls)).toBe(true);
    });
  });

  describe('content states', () => {
    it('should show error message and hide spinner/charts when error is set', () => {
      setChartRequest(mockHistoricRequest);
      errorSig.set('Failed to load data');
      fixture.detectChanges();

      expect(
        fixture.debugElement.query(By.css('.chart-error')).nativeElement.textContent,
      ).toContain('Failed to load data');
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubRealTimeChart))).toBeNull();
    });

    it('should show spinner and hide charts/error when loading', () => {
      setChartRequest(mockHistoricRequest);
      loadingSig.set(true);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.chart-error'))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubRealTimeChart))).toBeNull();
    });

    it('should render historic chart when not loading/error', () => {
      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubRealTimeChart))).toBeNull();
    });

    it('should render realtime chart when not loading/error', () => {
      setChartRequest(mockRealtimeRequest);
      expect(fixture.debugElement.query(By.directive(StubRealTimeChart))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
    });
  });

  describe('input bindings', () => {
    it('should pass historicReadings and sensor to historic chart', () => {
      historicReadingsSig.set(mockReadings);
      setChartRequest(mockHistoricRequest);

      const chart = fixture.debugElement.query(By.directive(StubHistoricChart))
        .componentInstance as StubHistoricChart;
      expect(chart.readings()).toEqual(mockReadings);
      expect(chart.sensor()).toEqual(mockSensor);
    });

    it('should pass liveReadings and sensor to realtime chart', () => {
      liveReadingsSig.set(mockReadings);
      setChartRequest(mockRealtimeRequest);

      const chart = fixture.debugElement.query(By.directive(StubRealTimeChart))
        .componentInstance as StubRealTimeChart;
      expect(chart.readings()).toEqual(mockReadings);
      expect(chart.sensor()).toEqual(mockSensor);
    });

    it('should update historic chart when readings change', () => {
      setChartRequest(mockHistoricRequest);
      const chart = fixture.debugElement.query(By.directive(StubHistoricChart))
        .componentInstance as StubHistoricChart;
      expect(chart.readings()).toEqual([]);

      historicReadingsSig.set(mockReadings);
      fixture.detectChanges();
      expect(chart.readings()).toEqual(mockReadings);
    });

    it('should update realtime chart when readings change', () => {
      setChartRequest(mockRealtimeRequest);
      const chart = fixture.debugElement.query(By.directive(StubRealTimeChart))
        .componentInstance as StubRealTimeChart;
      expect(chart.readings()).toEqual([]);

      liveReadingsSig.set(mockReadings);
      fixture.detectChanges();
      expect(chart.readings()).toEqual(mockReadings);
    });
  });

  describe('computed properties', () => {
    it.each([
      [mockHistoricRequest, 'isHistoricChart', true],
      [mockRealtimeRequest, 'isHistoricChart', false],
      [null, 'isHistoricChart', false],
      [mockRealtimeRequest, 'isLiveChart', true],
      [mockHistoricRequest, 'isLiveChart', false],
      [null, 'isLiveChart', false],
    ] as const)('chartRequest=%s → %s() === %s', (req, prop, expected) => {
      setChartRequest(req);
      expect(component[prop]()).toBe(expected);
    });

    it.each([
      ['connected', 'statusLabel', 'Connected'],
      ['connecting', 'statusLabel', 'Connecting...'],
      ['disconnected', 'statusLabel', 'Disconnected'],
      ['connected', 'statusClass', 'status-connected'],
      ['connecting', 'statusClass', 'status-connecting'],
      ['disconnected', 'statusClass', 'status-disconnected'],
    ] as const)('connectionStatus=%s → %s() === "%s"', (status, prop, expected) => {
      connectionStatusSig.set(status);
      expect(component[prop]()).toBe(expected);
    });

    it('profileDisplay returns correct value for known profiles', () => {
      setChartRequest(mockHistoricRequest);
      expect(component['profileDisplay']()).toEqual({ label: 'Heart Rate', unit: 'bpm' });

      setChartRequest(null);
      expect(component['profileDisplay']()).toBeNull();

      setChartRequest({
        sensor: { ...mockSensor, profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE },
        chartType: ChartType.HISTORIC,
        timeInterval: { from: new Date(), to: new Date() },
      });
      expect(component['profileDisplay']()).toEqual({ label: 'Thermometer', unit: '°C' });
    });
  });

  describe('ngOnDestroy', () => {
    it('should call stopChart on destroy', () => {
      setChartRequest(mockHistoricRequest);
      chartServiceMock.stopChart.mockClear();
      fixture.destroy();
      expect(chartServiceMock.stopChart).toHaveBeenCalledOnce();
    });
  });

  describe('signal reactivity', () => {
    it('should transition from placeholder to card when request is set', () => {
      setChartRequest(null);
      expect(fixture.debugElement.query(By.css('.chart-placeholder'))).toBeTruthy();

      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.css('.chart-placeholder'))).toBeNull();
      expect(fixture.debugElement.query(By.css('mat-card'))).toBeTruthy();
    });

    it('should transition between loading, chart and error', () => {
      setChartRequest(mockHistoricRequest);

      loadingSig.set(true);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();

      loadingSig.set(false);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();

      errorSig.set('Something broke');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
      expect(
        fixture.debugElement.query(By.css('.chart-error')).nativeElement.textContent,
      ).toContain('Something broke');
    });

    it('should switch between historic and realtime charts', () => {
      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();

      setChartRequest(mockRealtimeRequest);
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubRealTimeChart))).toBeTruthy();
    });
  });
});
