import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, NO_ERRORS_SCHEMA, ComponentRef } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';

import { ChartContainerComponent } from './chart-container.component';
import { SensorChartService } from '../../../../services/sensor-chart/sensor-chart.service';
import { ChartRequest } from '../../../../models/chart/chart-request.model';
import { ChartType } from '../../../../models/chart/chart-type.enum';
import { Sensor } from '../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../models/sensor/sensor-profiles.enum';
import { SensorReading } from '../../../../models/sensor-data/sensor-reading.model';
import { Status } from '../../../../models/gateway-sensor-status.enum';

describe('ChartContainerComponent', () => {
  let fixture: ComponentFixture<ChartContainerComponent>;
  let component: ChartContainerComponent;
  let componentRef: ComponentRef<ChartContainerComponent>;
  let loadingSig: WritableSignal<boolean>;
  let connectionStatusSig: WritableSignal<'connected' | 'connecting' | 'disconnected'>;
  let errorSig: WritableSignal<string | null>;
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

  const mockRealtimeRequest: ChartRequest = {
    sensor: mockSensor,
    chartType: ChartType.REALTIME,
  };

  beforeEach(async () => {
    loadingSig = signal(false);
    connectionStatusSig = signal<'connected' | 'connecting' | 'disconnected'>('disconnected');
    errorSig = signal<string | null>(null);

    chartServiceMock = {
      historicReadings: signal<SensorReading[]>([]),
      liveReadings: signal<SensorReading[]>([]),
      loading: loadingSig,
      connectionStatus: connectionStatusSig,
      error: errorSig,
      startChart: vi.fn(),
      stopChart: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [ChartContainerComponent],
      schemas: [NO_ERRORS_SCHEMA],
      providers: [{ provide: SensorChartService, useValue: chartServiceMock }],
    }).compileComponents();

    fixture = TestBed.createComponent(ChartContainerComponent);
    component = fixture.componentInstance;
    componentRef = fixture.componentRef;
  });

  function query(selector: string): HTMLElement | null {
    return fixture.nativeElement.querySelector(selector);
  }

  function setChartRequest(req: ChartRequest | null): void {
    componentRef.setInput('chartRequest', req);
    fixture.detectChanges();
  }

  describe('when chartRequest is null', () => {
    beforeEach(() => setChartRequest(null));

    it('should show placeholder and hide mat-card', () => {
      const placeholder = query('.chart-placeholder');
      expect(placeholder).not.toBeNull();
      expect(placeholder!.textContent).toContain('No chart selected');
      expect(query('mat-card')).toBeNull();
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
      const title = query('mat-card-title');
      expect(title!.textContent).toContain('Heart Rate Sensor');
      expect(title!.textContent).toContain('Heart Rate');
    });

    it('should display correct profile label for ECG sensor', () => {
      setChartRequest({
        sensor: { ...mockSensor, name: 'ECG Sensor', profile: SensorProfiles.CUSTOM_ECG_SERVICE },
        chartType: ChartType.HISTORIC,
        timeInterval: { from: new Date(), to: new Date() },
      });
      const title = query('mat-card-title');
      expect(title!.textContent).toContain('ECG Sensor');
      expect(title!.textContent).toContain('ECG');
    });
  });

  describe('close button', () => {
    it('should call stopChart and emit chartClosed when clicked', () => {
      setChartRequest(mockHistoricRequest);
      const closedSpy = vi.fn();
      component.chartClosed.subscribe(closedSpy);

      query('button[mat-icon-button]')!.click();

      expect(chartServiceMock.stopChart).toHaveBeenCalled();
      expect(closedSpy).toHaveBeenCalledOnce();
    });
  });

  describe('connection status', () => {
    it('should not show connection status for historic chart', () => {
      setChartRequest(mockHistoricRequest);
      expect(query('.connection-status')).toBeNull();
    });

    it.each([
      ['connected', 'Connected', 'status-connected'],
      ['connecting', 'Connecting...', 'status-connecting'],
      ['disconnected', 'Disconnected', 'status-disconnected'],
    ] as const)('should show "%s" label and class for realtime chart', (status, label, cls) => {
      setChartRequest(mockRealtimeRequest);
      connectionStatusSig.set(status);
      fixture.detectChanges();

      const el = query('.connection-status');
      expect(el).not.toBeNull();
      expect(el!.textContent).toContain(label);
      expect(el!.classList.contains(cls)).toBe(true);
    });
  });

  describe('content states', () => {
    it('should show error message and hide spinner/charts when error is set', () => {
      setChartRequest(mockHistoricRequest);
      errorSig.set('Failed to load data');
      fixture.detectChanges();

      const err = query('.chart-error');
      expect(err!.textContent).toContain('Failed to load data');
      expect(query('mat-spinner')).toBeNull();
      expect(query('app-historic-chart')).toBeNull();
      expect(query('app-real-time-chart')).toBeNull();
    });

    it('should show spinner and hide charts/error when loading', () => {
      setChartRequest(mockHistoricRequest);
      loadingSig.set(true);
      fixture.detectChanges();

      expect(query('mat-spinner')).not.toBeNull();
      expect(query('.chart-error')).toBeNull();
      expect(query('app-historic-chart')).toBeNull();
      expect(query('app-real-time-chart')).toBeNull();
    });

    it('should render app-historic-chart when historic and not loading/error', () => {
      setChartRequest(mockHistoricRequest);
      expect(query('app-historic-chart')).not.toBeNull();
      expect(query('app-real-time-chart')).toBeNull();
    });

    it('should render app-real-time-chart when realtime and not loading/error', () => {
      setChartRequest(mockRealtimeRequest);
      expect(query('app-real-time-chart')).not.toBeNull();
      expect(query('app-historic-chart')).toBeNull();
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
    ] as const)('chartRequest=%s to %s() === %s', (req, prop, expected) => {
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
    ] as const)('connectionStatus=%s to %s() === "%s"', (status, prop, expected) => {
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
      expect(query('.chart-placeholder')).not.toBeNull();

      setChartRequest(mockHistoricRequest);
      expect(query('.chart-placeholder')).toBeNull();
      expect(query('mat-card')).not.toBeNull();
    });

    it('should transition between loading, chart and error', () => {
      setChartRequest(mockHistoricRequest);
      loadingSig.set(true);
      fixture.detectChanges();
      expect(query('mat-spinner')).not.toBeNull();

      loadingSig.set(false);
      fixture.detectChanges();
      expect(query('app-historic-chart')).not.toBeNull();

      errorSig.set('Something broke');
      fixture.detectChanges();
      expect(query('app-historic-chart')).toBeNull();
      expect(query('.chart-error')!.textContent).toContain('Something broke');
    });

    it('should switch between historic and realtime charts', () => {
      setChartRequest(mockHistoricRequest);
      expect(query('app-historic-chart')).not.toBeNull();

      setChartRequest(mockRealtimeRequest);
      expect(query('app-historic-chart')).toBeNull();
      expect(query('app-real-time-chart')).not.toBeNull();
    });
  });
});
