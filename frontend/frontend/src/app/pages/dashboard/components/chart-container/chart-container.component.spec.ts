import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, ComponentRef, Component, input } from '@angular/core';
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
  let connectionStatusSig: WritableSignal<
    'connected' | 'connecting' | 'disconnected' | 'reconnecting'
  >;
  let errorSig: WritableSignal<string | null>;
  let historicReadingsSig: WritableSignal<SensorReading[]>;
  let liveReadingsSig: WritableSignal<SensorReading[]>;
  let chartServiceMock: Record<string, unknown>;

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
    connectionStatusSig = signal<'connected' | 'connecting' | 'disconnected' | 'reconnecting'>(
      'disconnected',
    );
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

  describe('no chart request', () => {
    it('should show placeholder, hide card, and not call startChart', () => {
      setChartRequest(null);

      const placeholder = fixture.debugElement.query(By.css('.chart-placeholder'));
      expect(placeholder).toBeTruthy();
      expect(placeholder.nativeElement.textContent).toContain('Nessun grafico selezionato');
      expect(fixture.debugElement.query(By.css('mat-card'))).toBeNull();
      expect(chartServiceMock['startChart']).not.toHaveBeenCalled();
    });
  });

  describe('startChart effect', () => {
    it('should call startChart on each request and on change', () => {
      setChartRequest(mockHistoricRequest);
      expect(chartServiceMock['startChart']).toHaveBeenCalledWith(mockHistoricRequest);

      setChartRequest(mockRealtimeRequest);
      expect(chartServiceMock['startChart']).toHaveBeenCalledTimes(2);
      expect(chartServiceMock['startChart']).toHaveBeenLastCalledWith(mockRealtimeRequest);
    });
  });

  describe('card header', () => {
    it.each([
      {
        name: 'Heart Rate Sensor',
        profile: SensorProfiles.HEART_RATE_SERVICE,
        label: 'Heart Rate',
      },
      { name: 'ECG Sensor', profile: SensorProfiles.CUSTOM_ECG_SERVICE, label: 'ECG' },
    ])('should display "$name" with "$label" profile label', ({ name, profile, label }) => {
      setChartRequest({
        sensor: { ...mockSensor, name, profile },
        chartType: ChartType.HISTORIC,
        timeInterval: { from: new Date(), to: new Date() },
      });

      const title = fixture.debugElement.query(By.css('mat-card-title'));
      expect(title.nativeElement.textContent).toContain(name);
      expect(title.nativeElement.textContent).toContain(label);
    });
  });

  describe('close button', () => {
    it('should call stopChart and emit chartClosed', () => {
      setChartRequest(mockHistoricRequest);
      const closedSpy = vi.fn();
      component.chartClosed.subscribe(closedSpy);

      fixture.debugElement.query(By.css('button[mat-icon-button]')).nativeElement.click();

      expect(chartServiceMock['stopChart']).toHaveBeenCalled();
      expect(closedSpy).toHaveBeenCalledOnce();
    });
  });

  describe('connection status', () => {
    it('should not show for historic chart', () => {
      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.css('.connection-status'))).toBeNull();
    });

    it.each([
      ['connected', 'Connected', 'status-connected'],
      ['connecting', 'Connecting...', 'status-connecting'],
      ['disconnected', 'Disconnected', 'status-disconnected'],
      ['reconnecting', 'Reconnecting...', 'status-reconnecting'],
    ] as const)('should show "%s" with correct label and class', (status, label, cls) => {
      setChartRequest(mockRealtimeRequest);
      connectionStatusSig.set(status);
      fixture.detectChanges();

      const el = fixture.debugElement.query(By.css('.connection-status'));
      expect(el.nativeElement.textContent).toContain(label);
      expect(el.nativeElement.classList.contains(cls)).toBe(true);
    });
  });

  describe('content states', () => {
    it('should show only spinner when loading', () => {
      setChartRequest(mockHistoricRequest);
      loadingSig.set(true);
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.chart-error'))).toBeNull();
      expect(fixture.debugElement.query(By.css('.chart-warning'))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
    });

    it('should show chart-error when error is set and not reconnecting', () => {
      setChartRequest(mockHistoricRequest);
      errorSig.set('Failed to load data');
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.chart-error'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.chart-warning'))).toBeNull();
    });

    it('should show chart-warning when reconnecting', () => {
      setChartRequest(mockRealtimeRequest);
      connectionStatusSig.set('reconnecting');
      errorSig.set('Connection lost. Retry 1/3...');
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.chart-warning'))).toBeTruthy();
      expect(fixture.debugElement.query(By.css('.chart-error'))).toBeNull();
    });

    it('should show error and chart together when readings exist', () => {
      historicReadingsSig.set(mockReadings);
      setChartRequest(mockHistoricRequest);
      errorSig.set('Something went wrong');
      fixture.detectChanges();

      expect(fixture.debugElement.query(By.css('.chart-error'))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();
    });

    it.each([
      { type: 'historic', request: mockHistoricRequest, stub: StubHistoricChart },
      { type: 'realtime', request: mockRealtimeRequest, stub: StubRealTimeChart },
    ])('should not render $type chart when readings are empty', ({ request, stub }) => {
      setChartRequest(request);
      expect(fixture.debugElement.query(By.directive(stub))).toBeNull();
    });

    it.each([
      {
        type: 'historic',
        request: mockHistoricRequest,
        stub: StubHistoricChart,
        sig: () => historicReadingsSig,
      },
      {
        type: 'realtime',
        request: mockRealtimeRequest,
        stub: StubRealTimeChart,
        sig: () => liveReadingsSig,
      },
    ])('should render $type chart when readings exist', ({ request, stub, sig }) => {
      sig().set(mockReadings);
      setChartRequest(request);
      expect(fixture.debugElement.query(By.directive(stub))).toBeTruthy();
    });
  });

  describe('input bindings', () => {
    it.each([
      {
        type: 'historic',
        request: mockHistoricRequest,
        stub: StubHistoricChart,
        sig: () => historicReadingsSig,
      },
      {
        type: 'realtime',
        request: mockRealtimeRequest,
        stub: StubRealTimeChart,
        sig: () => liveReadingsSig,
      },
    ])(
      'should pass readings and sensor to $type chart and update reactively',
      ({ request, stub, sig }) => {
        setChartRequest(request);
        expect(fixture.debugElement.query(By.directive(stub))).toBeNull();

        sig().set(mockReadings);
        fixture.detectChanges();

        const chart = fixture.debugElement.query(By.directive(stub)).componentInstance;
        expect(chart.readings()).toEqual(mockReadings);
        expect(chart.sensor()).toEqual(mockSensor);
      },
    );
  });

  describe('computed properties', () => {
    it.each([
      [mockHistoricRequest, 'isHistoricChart', true],
      [mockRealtimeRequest, 'isHistoricChart', false],
      [mockRealtimeRequest, 'isLiveChart', true],
      [mockHistoricRequest, 'isLiveChart', false],
    ] as const)('%s → %s() === %s', (req, prop, expected) => {
      setChartRequest(req);
      expect(component[prop]()).toBe(expected);
    });

    it.each([
      ['connected', 'statusLabel', 'Connected'],
      ['connecting', 'statusLabel', 'Connecting...'],
      ['disconnected', 'statusLabel', 'Disconnected'],
      ['reconnecting', 'statusLabel', 'Reconnecting...'],
      ['connected', 'statusClass', 'status-connected'],
      ['connecting', 'statusClass', 'status-connecting'],
      ['disconnected', 'statusClass', 'status-disconnected'],
      ['reconnecting', 'statusClass', 'status-reconnecting'],
    ] as const)('connectionStatus=%s → %s() === "%s"', (status, prop, expected) => {
      connectionStatusSig.set(status);
      expect(component[prop]()).toBe(expected);
    });

    it.each([
      {
        profile: SensorProfiles.HEART_RATE_SERVICE,
        expected: { label: 'Heart Rate', unit: 'bpm' },
      },
      {
        profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
        expected: { label: 'Thermometer', unit: '°C' },
      },
    ])('profileDisplay for $profile', ({ profile, expected }) => {
      setChartRequest({
        sensor: { ...mockSensor, profile },
        chartType: ChartType.HISTORIC,
        timeInterval: { from: new Date(), to: new Date() },
      });
      expect(component['profileDisplay']()).toEqual(expected);
    });

    it('profileDisplay returns null when no request', () => {
      setChartRequest(null);
      expect(component['profileDisplay']()).toBeNull();
    });
  });

  describe('ngOnDestroy', () => {
    it('should call stopChart', () => {
      setChartRequest(mockHistoricRequest);
      (chartServiceMock['stopChart'] as ReturnType<typeof vi.fn>).mockClear();
      fixture.destroy();
      expect(chartServiceMock['stopChart']).toHaveBeenCalledOnce();
    });
  });

  describe('signal reactivity', () => {
    it('should transition from placeholder to card', () => {
      setChartRequest(null);
      expect(fixture.debugElement.query(By.css('.chart-placeholder'))).toBeTruthy();

      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.css('.chart-placeholder'))).toBeNull();
      expect(fixture.debugElement.query(By.css('mat-card'))).toBeTruthy();
    });

    it('should transition through loading, chart and error states', () => {
      historicReadingsSig.set(mockReadings);
      setChartRequest(mockHistoricRequest);

      loadingSig.set(true);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('mat-spinner'))).toBeTruthy();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();

      loadingSig.set(false);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();

      errorSig.set('Something broke');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();
      expect(
        fixture.debugElement.query(By.css('.chart-error')).nativeElement.textContent,
      ).toContain('Something broke');
    });

    it('should switch between historic and realtime charts', () => {
      historicReadingsSig.set(mockReadings);
      setChartRequest(mockHistoricRequest);
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeTruthy();

      liveReadingsSig.set(mockReadings);
      setChartRequest(mockRealtimeRequest);
      expect(fixture.debugElement.query(By.directive(StubHistoricChart))).toBeNull();
      expect(fixture.debugElement.query(By.directive(StubRealTimeChart))).toBeTruthy();
    });
  });
});
